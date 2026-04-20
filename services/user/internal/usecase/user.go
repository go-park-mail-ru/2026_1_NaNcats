package usecase

import (
	"bytes"
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/imageutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user UserUseCase
type UserUseCase interface {
	Create(ctx context.Context, user domain.User, idempotencyKey string) (int64, error)
	GetByID(ctx context.Context, userID int64) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Check(ctx context.Context, userID int64) (bool, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error
	UpdateAvatar(ctx context.Context, userID int64, imageData []byte, idempotencyKey string) (string, error)
	DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error)
}

type userUseCase struct {
	userRepo         repository.UserRepository
	fileStorage      s3.FileStorage
	defaultAvatarURL string
}

func NewUserUseCase(ur repository.UserRepository, fs s3.FileStorage, daurl string) UserUseCase {
	return &userUseCase{
		userRepo:         ur,
		fileStorage:      fs,
		defaultAvatarURL: daurl,
	}
}

// создаем юзера
func (u *userUseCase) Create(ctx context.Context, user domain.User, idempotencyKey string) (int64, error) {
	id, err := u.userRepo.CreateUser(ctx, user, idempotencyKey)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return 0, err
		}
		return 0, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create user in db", err, codes.Internal)
	}

	return id, nil
}

// возвращает юзера по переданному userID
func (u *userUseCase) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get user from db", err, codes.Internal)
	}

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}

// возвращает юзера по переданной почте
func (u *userUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get user by email from db", err, codes.Internal)
	}

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}

// проверяет существует ли юзер
func (u *userUseCase) Check(ctx context.Context, userID int64) (bool, error) {
	isExists, err := u.userRepo.CheckUserByID(ctx, userID)
	if err != nil {
		return false, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to check user existence in db", err, codes.Internal)
	}

	return isExists, nil
}

// обновляет поля юзера
func (u *userUseCase) UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error {
	err := u.userRepo.UpdateProfile(ctx, userID, name, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return err
		}
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return err
		}
		return errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to update user profile", err, codes.Internal)
	}

	return nil
}

func (u *userUseCase) UpdateAvatar(ctx context.Context, userID int64, imageData []byte, idempotencyKey string) (string, error) {
	user, err := u.GetByID(ctx, userID)
	if err != nil {
		// Ошибка уже обернута в errutil внутри GetByID
		return "", err
	}

	imageReader := bytes.NewReader(imageData)

	webpData, err := imageutil.ConvertToWebp(imageReader)
	if err != nil {
		return "", domain.ErrInvalidImageExt
	}

	filename := "avatars/" + uuid.NewString() + ".webp"

	newAvatarURL, err := u.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
	if err != nil {
		return "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to upload to S3", err, codes.Internal)
	}

	err = u.userRepo.UpdateAvatarURL(ctx, userID, newAvatarURL)
	if err != nil {
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(newAvatarURL)
		return "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to update avatar in db", err, codes.Internal)
	}

	if user.AvatarURL != u.defaultAvatarURL && user.AvatarURL != "" {
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(user.AvatarURL)
	}

	return newAvatarURL, nil
}

func (u *userUseCase) DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error) {
	user, err := u.GetByID(ctx, userID)
	if err != nil {
		// Ошибка уже обернута в errutil внутри GetByID
		return "", err
	}

	if user.AvatarURL == u.defaultAvatarURL || user.AvatarURL == "" {
		return u.defaultAvatarURL, nil
	}

	urlToDelete := user.AvatarURL

	err = u.userRepo.UpdateAvatarURL(ctx, userID, "")
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", err
		}
		return "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to reset avatar in db", err, codes.Internal)
	}

	go func(urlToDelete string) {
		_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
	}(urlToDelete)

	return u.defaultAvatarURL, nil
}
