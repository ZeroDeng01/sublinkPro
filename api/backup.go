package api

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sublink/config"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func Backup(c *gin.Context) {
	// 创建临时文件用于存储压缩包
	tmpFile, err := os.CreateTemp("", "backup-*.zip")
	if err != nil {
		utils.FailWithMsg(c, "Failed to create temp file")
		return
	}
	defer os.Remove(tmpFile.Name()) // 确保函数退出时删除临时文件
	defer tmpFile.Close()           // 确保函数退出时关闭临时文件

	// 创建zip写入器
	zipWriter := zip.NewWriter(tmpFile)
	// defer zipWriter.Close() // 不在这里 defer，我们需要在发送文件前手动 Close

	// 获取配置的数据库目录路径
	dbPath := config.GetDBPath()

	// 获取模板目录路径（基于当前工作目录）
	templatePath := "template"
	if cwd, err := os.Getwd(); err == nil {
		templatePath = filepath.Join(cwd, "template")
	}

	// 备份目录配置：源路径 -> zip 内的目录名
	type backupFolder struct {
		sourcePath string // 实际文件系统路径
		zipName    string // zip 中的目录名称
	}
	folders := []backupFolder{
		{sourcePath: dbPath, zipName: "db"},
		{sourcePath: templatePath, zipName: "template"},
	}

	// 遍历文件夹并添加到zip文件中
	for _, folder := range folders {
		baseDir := folder.sourcePath
		zipPrefix := folder.zipName

		// filepath.Walk 会遍历所有子文件和子目录
		walkErr := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 计算相对路径
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}

			// 创建zip中的文件头
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// 构建 zip 中的路径：zipPrefix + 相对路径
			if relPath == "." {
				header.Name = zipPrefix
			} else {
				header.Name = filepath.ToSlash(filepath.Join(zipPrefix, relPath))
			}

			// 如果是目录，需要以/结尾
			if info.IsDir() {
				header.Name += "/"
			} else {
				// 设置压缩方法
				header.Method = zip.Deflate
			}

			// 创建zip中的文件
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			// 如果是文件，写入文件内容
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err // **[修正]** 只返回 error
				}

				// **[关键修正]** 不要使用 defer file.Close()！
				// 立即复制并关闭文件，防止文件句柄泄露
				_, err = io.Copy(writer, file)
				file.Close() // <--- 立即关闭

				if err != nil {
					return err // **[修正]** 只返回 error
				}
			}
			return nil
		})

		// 在 Walk 循环结束后，统一检查错误
		if walkErr != nil {
			utils.FailWithMsg(c, "备份目录 '"+zipPrefix+"' 失败: "+walkErr.Error())
			return
		}
	}

	// **[关键修正]** 必须在发送文件 *之前* 关闭 zipWriter，
	// 这样才能将 zip 的中央目录结构写入文件
	err = zipWriter.Close()
	if err != nil {
		utils.FailWithMsg(c, "Failed to close zip writer: "+err.Error())
		return
	}

	// 确保临时文件已完全写入（可选，但在某些系统上更安全）
	err = tmpFile.Sync()
	if err != nil {
		utils.FailWithMsg(c, "Failed to sync temp file: "+err.Error())
		return
	}

	// **[修正]** 在所有操作都成功后，再设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=sublink-pro-backup.zip")

	// 将文件指针重置到文件开头
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		utils.FailWithMsg(c, "Failed to seek temp file: "+err.Error())
		return
	}

	// 将文件内容发送给客户端
	_, err = io.Copy(c.Writer, tmpFile)
	if err != nil {
		// 此时可能已经发送了部分响应，JSON 可能无效，但尽力而为
		// utils.FailWithMsg(c, "Failed to send file to client: "+err.Error())
		// Since headers might be sent, we can't cleanly send JSON. But let's leave it or log it.
		// The original code tried to send JSON. I'll keep it consistent.
		utils.FailWithMsg(c, "Failed to send file to client: "+err.Error())
		return
	}
}
