package repository

import (
	"context"
)

// контракт хранилища фотографий
type ImageStorage interface {
	// получает байты по URL-у картинки
	Download(ctx context.Context, filePath string) ([]byte, error)
}
