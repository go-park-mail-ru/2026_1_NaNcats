package repository

import (
	"context"
)

// контракт хранилища фотографий
type ImageStorage interface {
	// сохраняет байты и возвращает полный URL файла
	Upload(ctx context.Context, fileBytes []byte, originalName, folder string) (string, error)
	// получает байты по URL-у картинки
	Download(ctx context.Context, filePath string) ([]byte, error)
}
