package api

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"strings"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// DownloadSkill 返回一个处理器，将嵌入的 skills/sublinkpro/ 目录即时打包为 zip 并下发。
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

		// 将 sublinkpro 保留为解压目录，使目录名与 SKILL.md 中的 name 一致。
		walkErr := fs.WalkDir(skillFS, "skills/sublinkpro", func(path string, d fs.DirEntry, err error) error {
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
			header.Name = strings.TrimPrefix(path, "skills/")
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
