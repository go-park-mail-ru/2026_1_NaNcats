package repository

import (
	"context"
	"io"
)

type FileStorage interface {
	UploadFile(ctx context.Context, file io.Reader, filename, contentType string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}
