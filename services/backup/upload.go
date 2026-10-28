package backup

import (
	"context"
	"os"
)

// CreateAndUpload builds a system backup archive and uploads it to the configured WebDAV target.
// The caller owns no temporary files; the local archive is removed after upload.
func CreateAndUpload(ctx context.Context, cfg Config) (RemoteFile, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return RemoteFile{}, err
	}
	archive, err := CreateArchiveFile(ctx)
	if err != nil {
		return RemoteFile{}, err
	}
	defer func() { _ = os.Remove(archive.Path) }()
	return client.Upload(ctx, archive)
}
