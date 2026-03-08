package memory

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type imageStorage struct {
	basePath string
}

func NewImageStorage() repository.ImageStorage {
	return &imageStorage{
		basePath: "./uploads",
	}
}

func (s *imageStorage) Download(ctx context.Context, filePath string) ([]byte, error) {
	fullFilePath := filepath.Join(s.basePath, filePath)
	cleanPath := filepath.Clean(fullFilePath)

	photo, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}

	return photo, nil
}
