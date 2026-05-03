package s3

import (
	"context"
	"io"
)

//go:generate mockgen -destination=mocks/file_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3 FileStorage
type FileStorage interface {
	UploadFile(ctx context.Context, file io.Reader, filename, contentType string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}
