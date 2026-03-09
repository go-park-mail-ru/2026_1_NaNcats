package usecase

import (
	"context"
	"strings"

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
		if strings.HasPrefix(filepath, "users") {
			photo, err = u.imageStorage.Download(ctx, "default/avatar.png")
			if err != nil {
				return nil, err
			}
			return photo, nil
		} else if strings.HasPrefix(filepath, "restaurants/banners") {
			photo, err = u.imageStorage.Download(ctx, "default/banner.png")
			if err != nil {
				return nil, err
			}
			return photo, nil
		} else if strings.HasPrefix(filepath, "restaurants/logos") {
			photo, err = u.imageStorage.Download(ctx, "default/logo.png")
			if err != nil {
				return nil, err
			}
			return photo, nil
		}
		return nil, err
	}

	return photo, err
}
