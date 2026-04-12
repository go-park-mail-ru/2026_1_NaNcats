package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/imageutil"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase UserUseCase
type UserUseCase interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetByID(ctx context.Context, userID int) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Check(ctx context.Context, userID int) (bool, error)
	UpdateProfile(ctx context.Context, userID int, name, email *string) error
	UpdateAvatar(ctx context.Context, userID int, file io.Reader) (string, error)
	DeleteAvatar(ctx context.Context, userID int) error
}

type userUseCase struct {
	userRepo    repository.UserRepository
	fileStorage repository.FileStorage
}

func NewUserUseCase(ur repository.UserRepository, fs repository.FileStorage) UserUseCase {
	return &userUseCase{
		userRepo:    ur,
		fileStorage: fs,
	}
}

// создаем юзера
func (u *userUseCase) Create(ctx context.Context, user domain.User) (int, error) {
	id, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// возвращает юзера по переданному userID
func (u *userUseCase) GetByID(ctx context.Context, userID int) (domain.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// возвращает юзера по переданной почте
func (u *userUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// проверяет существует ли юзер
func (u *userUseCase) Check(ctx context.Context, userID int) (bool, error) {
	isExists, err := u.userRepo.CheckUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

// обновляет поля юзера
func (u *userUseCase) UpdateProfile(ctx context.Context, userID int, name, email *string) error {
	return u.userRepo.UpdateProfile(ctx, userID, name, email)
}

func (u *userUseCase) UpdateAvatar(ctx context.Context, userID int, file io.Reader) (string, error) {
	user, err := u.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	webpData, err := imageutil.ConvertToWebp(file)
	if err != nil {
		return "", domain.ErrInvalidImageExt
	}

	filename := "avatars/" + uuid.New().String() + ".webp"

	newAvatarURL, err := u.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
	if err != nil {
		return "", err
	}

	err = u.userRepo.UpdateAvatarURL(ctx, userID, newAvatarURL)
	if err != nil {
		// если фотка загружена на S3, но по какой-то причине не обновился URL у юзера, то удаляем фотку
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), user.AvatarURL)
		}(user.AvatarURL)
		return "", err
	}

	if user.AvatarURL != "" {
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), user.AvatarURL)
		}(user.AvatarURL)
	}

	return newAvatarURL, nil
}

func (u *userUseCase) DeleteAvatar(ctx context.Context, userID int) error {
	user, err := u.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.AvatarURL == "" {
		return nil
	}

	err = u.userRepo.UpdateAvatarURL(ctx, userID, "")
	if err != nil {
		return err
	}

	go func(urlToDelete string) {
		_ = u.fileStorage.DeleteFile(context.Background(), user.AvatarURL)
	}(user.AvatarURL)

	return nil
}
