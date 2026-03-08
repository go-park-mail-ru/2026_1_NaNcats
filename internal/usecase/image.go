package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type ImageUseCase interface {
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

func (u *imageUseCase) GetImage(ctx context.Context, filepath string) ([]byte, error) {
	photo, err := u.imageStorage.Download(ctx, filepath)
	if err != nil {
		return nil, err
	}

	return photo, err
}
