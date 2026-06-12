package api

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// DownloadSkill 返回一个处理器，将嵌入的 skill-sublinkpro/ 目录即时打包为 zip 并下发。
// skillFS 由 main.go 通过 //go:embed 注入，避免 api 包反向 import main 包。
func DownloadSkill(skillFS fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		tmpFile, err := os.CreateTemp("", "skill-sublinkpro-*.zip")
		if err != nil {
			utils.FailWithMsg(c, "Failed to create temp file")
			return
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		defer func() { _ = tmpFile.Close() }()

		zipWriter := zip.NewWriter(tmpFile)

		// fs.WalkDir 产出的 path 自带 "skill-sublinkpro/" 前缀，直接作为 zip 内路径，
		// 这样解压后即得到 skill-sublinkpro/ 目录，可直接放入 ~/.claude/skills/。
		// 与 backup.go 保持一致（其 zip 根部保留 db/ 与 template/）。
		walkErr := fs.WalkDir(skillFS, "skill-sublinkpro", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(path)
			if d.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			data, err := fs.ReadFile(skillFS, path)
			if err != nil {
				return err
			}
			_, err = writer.Write(data)
			return err
		})
		if walkErr != nil {
			utils.FailWithMsg(c, "Failed to archive skill: "+walkErr.Error())
			return
		}

		// 必须在发送前关闭 zipWriter，才能写入 zip 中央目录结构。
		if err := zipWriter.Close(); err != nil {
			utils.FailWithMsg(c, "Failed to close zip writer: "+err.Error())
			return
		}
		_ = tmpFile.Sync()

		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", "attachment; filename=skill-sublinkpro.zip")
		if _, err := tmpFile.Seek(0, 0); err != nil {
			utils.FailWithMsg(c, "Failed to seek temp file: "+err.Error())
			return
		}
		_, _ = io.Copy(c.Writer, tmpFile)
	}
}
