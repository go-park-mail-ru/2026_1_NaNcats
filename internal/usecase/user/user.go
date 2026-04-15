package user

import (
	"context"
	"html"
	"io"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/imageutil"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user UserUseCase
type UserUseCase interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetByID(ctx context.Context, userID int) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Check(ctx context.Context, userID int) (bool, error)
	UpdateProfile(ctx context.Context, userID int, name, email *string) error
	UpdateAvatar(ctx context.Context, userID int, file io.Reader) (string, error)
	DeleteAvatar(ctx context.Context, userID int) (string, error)
}

type userUseCase struct {
	userRepo         repository.UserRepository
	fileStorage      repository.FileStorage
	defaultAvatarURL string
}

func NewUserUseCase(ur repository.UserRepository, fs repository.FileStorage, daurl string) UserUseCase {
	return &userUseCase{
		userRepo:         ur,
		fileStorage:      fs,
		defaultAvatarURL: daurl,
	}
}

// создаем юзера
func (u *userUseCase) Create(ctx context.Context, user domain.User) (int, error) {
	user.Name = html.EscapeString(user.Name)
	user.Email = html.EscapeString(user.Email)
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

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}

// возвращает юзера по переданной почте
func (u *userUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	email = html.EscapeString(email)
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
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
	if name != nil {
		escapedName := html.EscapeString(*name)
		name = &escapedName
	}
	if email != nil {
		escapedEmail := html.EscapeString(*email)
		email = &escapedEmail
	}

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

	filename := "avatars/" + uuid.NewString() + ".webp"

	newAvatarURL, err := u.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
	if err != nil {
		return "", err
	}

	err = u.userRepo.UpdateAvatarURL(ctx, userID, newAvatarURL)
	if err != nil {
		// если фотка загружена на S3, но по какой-то причине не обновился URL у юзера, то удаляем фотку
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(newAvatarURL)
		return "", err
	}

	if user.AvatarURL != u.defaultAvatarURL && user.AvatarURL != "" {
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(user.AvatarURL)
	}

	return newAvatarURL, nil
}

func (u *userUseCase) DeleteAvatar(ctx context.Context, userID int) (string, error) {
	user, err := u.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.AvatarURL == u.defaultAvatarURL || user.AvatarURL == "" {
		return u.defaultAvatarURL, nil
	}

	urlToDelete := user.AvatarURL

	err = u.userRepo.UpdateAvatarURL(ctx, userID, "")
	if err != nil {
		return u.defaultAvatarURL, err
	}

	go func(urlToDelete string) {
		_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
	}(urlToDelete)

	return u.defaultAvatarURL, nil
}
