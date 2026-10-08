package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"archive/zip"

	"sublink/config"
)

// Archive describes a generated SublinkPro backup archive.
type Archive struct {
	Path     string
	Name     string
	Size     int64
	Modified time.Time
}

type archiveFolder struct {
	sourcePath string
	zipName    string
}

// CreateArchiveFile creates a temporary backup ZIP containing db/ and template/.
// The caller owns Archive.Path and must remove it after use.
func CreateArchiveFile(ctx context.Context) (Archive, error) {
	tmpFile, err := os.CreateTemp("", "sublink-pro-backup-*.zip")
	if err != nil {
		return Archive{}, fmt.Errorf("创建备份临时文件失败: %w", err)
	}
	path := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(path)
	}

	templatePath := "template"
	if cwd, cwdErr := os.Getwd(); cwdErr == nil {
		templatePath = filepath.Join(cwd, "template")
	}
	folders := []archiveFolder{
		{sourcePath: config.GetDBPath(), zipName: "db"},
		{sourcePath: templatePath, zipName: "template"},
	}

	limited := &archiveSizeWriter{Writer: tmpFile, Max: MaxBackupArchiveBytes}
	if err := writeArchive(ctx, limited, folders); err != nil {
		cleanup()
		return Archive{}, err
	}
	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return Archive{}, fmt.Errorf("同步备份文件失败: %w", err)
	}
	info, err := tmpFile.Stat()
	if err != nil {
		cleanup()
		return Archive{}, fmt.Errorf("读取备份文件信息失败: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(path)
		return Archive{}, fmt.Errorf("关闭备份文件失败: %w", err)
	}

	now := time.Now()
	return Archive{
		Path:     path,
		Name:     "sublink-pro-backup-" + now.Format("20060102-150405.000000000") + ".zip",
		Size:     info.Size(),
		Modified: now,
	}, nil
}

type archiveSizeWriter struct {
	Writer io.Writer
	Max    int64
	Size   int64
}

func (writer *archiveSizeWriter) Write(data []byte) (int, error) {
	if writer.Size+int64(len(data)) > writer.Max {
		return 0, fmt.Errorf("备份压缩包超过 %d 字节限制", writer.Max)
	}
	written, err := writer.Writer.Write(data)
	writer.Size += int64(written)
	return written, err
}

func writeArchive(ctx context.Context, dst io.Writer, folders []archiveFolder) error {
	zipWriter := zip.NewWriter(dst)
	closed := false
	defer func() {
		if !closed {
			_ = zipWriter.Close()
		}
	}()

	for _, folder := range folders {
		if err := addFolder(ctx, zipWriter, folder); err != nil {
			return err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("关闭 ZIP 写入器失败: %w", err)
	}
	closed = true
	return nil
}

func addFolder(ctx context.Context, zipWriter *zip.Writer, folder archiveFolder) error {
	return filepath.Walk(folder.sourcePath, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("读取备份目录 %s 失败: %w", folder.zipName, walkErr)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("备份目录包含不允许的符号链接: %s", filePath)
		}
		if !info.IsDir() && info.Name() == "GeoLite2-City.mmdb" {
			return nil
		}
		relPath, err := filepath.Rel(folder.sourcePath, filePath)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if relPath == "." {
			header.Name = folder.zipName
		} else {
			header.Name = filepath.ToSlash(filepath.Join(folder.zipName, relPath))
		}
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}
