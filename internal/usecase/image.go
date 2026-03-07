package usecase

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type ImageUseCase interface {
	UploadImage(ctx context.Context, fileBytes []byte, originalName, folder string) (string, error)
	GetImage(ctx context.Context, filePath string) ([]byte, error)
}

type imageUseCase struct {
	imageStorage repository.ImageStorage
}

func NewImageUseCase(is repository.ImageStorage) ImageUseCase {
	return &imageUseCase{
		imageStorage: is,
	}
}

func (u *imageUseCase) UploadImage(ctx context.Context, fileBytes []byte, originalName, folder string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		return "", domain.ErrInvalidImageExt
	}

	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		return "", domain.ErrInvalidImageExt
	}

	photoURL, err := u.imageStorage.Upload(ctx, fileBytes, originalName, folder)
	if err != nil {
		return "", err
	}
	return photoURL, nil
}

func (u *imageUseCase) GetImage(ctx context.Context, filepath string) ([]byte, error) {
	photo, err := u.imageStorage.Download(ctx, filepath)
	if err != nil {
		return nil, err
	}

	return photo, err
}
