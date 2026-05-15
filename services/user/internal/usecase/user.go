package usecase

import (
	"bytes"
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/imageutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	passUtil "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/password"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase UserUseCase
//go:generate gowrap gen -i UserUseCase -t ../../../../shared/templates/tracing.tmpl -o user_tracing_mw.go -v TracerName=user-service
type UserUseCase interface {
	Create(ctx context.Context, user domain.User, password, idempotencyKey string) (int64, error)
	GetByID(ctx context.Context, userID int64) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Check(ctx context.Context, userID int64) (bool, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error
	UpdateAvatar(ctx context.Context, userID int64, imageData []byte, idempotencyKey string) (string, error)
	DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	UpdateRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) error
	GetUsersByIDs(ctx context.Context, userIDs []int64) (map[int64]domain.User, error)
	GetByPublicID(ctx context.Context, publicID string) (domain.User, error)
}

//go:generate mockgen -destination=mocks/message_publisher_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase MessagePublisher
type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, data any) error
}

type userUseCase struct {
	userRepo         repository.UserRepository
	fileStorage      s3.FileStorage
	defaultAvatarURL string
	rabbitPublisher  MessagePublisher
	logger           logger.Logger
}

func NewUserUseCase(ur repository.UserRepository, fs s3.FileStorage, daurl string, mp MessagePublisher, l logger.Logger) UserUseCase {
	return &userUseCase{
		userRepo:         ur,
		fileStorage:      fs,
		defaultAvatarURL: daurl,
		rabbitPublisher:  mp,
		logger:           l,
	}
}

func (u *userUseCase) Create(ctx context.Context, user domain.User, password, idempotencyKey string) (int64, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("idempotency_key", idempotencyKey))

	hashedPassword, err := passUtil.HashPassword(password, passUtil.DefaultParams)
	if err != nil {
		return 0, errutil.Internal("failed to hash password", err)
	}
	user.PasswordHash = hashedPassword

	id, err := u.userRepo.CreateUser(ctx, user, idempotencyKey)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return 0, err
		}
		return 0, errutil.Internal("failed to create user in db", err)
	}

	span.SetAttributes(attribute.Int64("user.id", id))
	return id, nil
}

func (u *userUseCase) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, errutil.Internal("failed to get user from db", err)
	}

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}

func (u *userUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, errutil.Internal("failed to get user by email from db", err)
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", user.ID))

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}

func (u *userUseCase) Check(ctx context.Context, userID int64) (bool, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	isExists, err := u.userRepo.CheckUserByID(ctx, userID)
	if err != nil {
		return false, errutil.Internal("failed to check user existence in db", err)
	}

	return isExists, nil
}

func (u *userUseCase) UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("idempotency_key", idempotencyKey),
		attribute.Bool("update.name_present", name != nil),
		attribute.Bool("update.email_present", email != nil),
	)

	err := u.userRepo.UpdateProfile(ctx, userID, name, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrEmailAlreadyExists) {
			return err
		}
		return errutil.Internal("failed to update user profile", err)
	}

	return nil
}

func (u *userUseCase) UpdateAvatar(ctx context.Context, userID int64, imageData []byte, idempotencyKey string) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.Int("image.input_size_bytes", len(imageData)),
	)

	user, err := u.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	imageReader := bytes.NewReader(imageData)

	span.AddEvent("converting_image_to_webp")
	webpData, err := imageutil.ConvertToWebp(imageReader)
	if err != nil {
		return "", domain.ErrInvalidImageExt
	}
	span.SetAttributes(attribute.Int("image.webp_size_bytes", webpData.Len()))

	filename := "avatars/" + uuid.NewString() + ".webp"
	span.SetAttributes(attribute.String("s3.target_filename", filename))

	newAvatarURL, err := u.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
	if err != nil {
		return "", errutil.Internal("failed to upload to S3 storage", err)
	}

	err = u.userRepo.UpdateAvatarURL(ctx, userID, newAvatarURL)
	if err != nil {
		span.AddEvent("db_update_failed_cleaning_up_s3")
		// Если не удалось обновить БД, удаляем файл, который уже успел улететь в S3
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(newAvatarURL)
		return "", errutil.Internal("failed to update avatar path in database", err)
	}

	// Удаление старого аватара
	if user.AvatarURL != u.defaultAvatarURL && user.AvatarURL != "" {
		span.AddEvent("cleaning_up_old_avatar")
		go func(urlToDelete string) {
			_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
		}(user.AvatarURL)
	}

	return newAvatarURL, nil
}

func (u *userUseCase) DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	user, err := u.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.AvatarURL == u.defaultAvatarURL || user.AvatarURL == "" {
		span.AddEvent("avatar_is_already_default")
		return u.defaultAvatarURL, nil
	}

	urlToDelete := user.AvatarURL
	err = u.userRepo.UpdateAvatarURL(ctx, userID, "")
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", err
		}
		return "", errutil.Internal("failed to reset avatar in database", err)
	}

	span.AddEvent("scheduling_old_avatar_deletion")
	go func(urlToDelete string) {
		_ = u.fileStorage.DeleteFile(context.Background(), urlToDelete)
	}(urlToDelete)

	return u.defaultAvatarURL, nil
}

func (u *userUseCase) UpdateRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("admin.target_user_id", userID),
		attribute.String("admin.new_role", newRole),
	)

	validRoles := map[string]bool{
		domain.RoleClient: true, domain.RoleCourier: true,
		domain.RoleOwner: true, domain.RoleAdmin: true, domain.RoleSupport: true,
	}
	if !validRoles[newRole] {
		return domain.ErrInvalidInput
	}

	oldRole, shouldNotify, err := u.userRepo.UpdateUserRole(ctx, userID, newRole, idempotencyKey)
	if err != nil {
		return err
	}

	// Если это был повторный запрос с тем же ключом, shouldNotify будет false
	if !shouldNotify {
		return nil
	}

	if oldRole != newRole {
		event := events.UserRoleChangedEvent{
			UserID:  userID,
			OldRole: oldRole,
			NewRole: newRole,
		}
		err = u.rabbitPublisher.PublishJSON(ctx, events.QueueUserEvents, event)
		if err != nil {
			// Логируем, но ошибку не возвращаем, так как БД уже обновилась
			u.logger.Error("failed to publish role change event", err)
		}
	}

	return nil
}

func (u *userUseCase) GetUsersByIDs(ctx context.Context, userIDs []int64) (map[int64]domain.User, error) {
	if len(userIDs) == 0 {
		return make(map[int64]domain.User), nil
	}

	users, err := u.userRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, errutil.Internal("failed to fetch users batch", err)
	}

	result := make(map[int64]domain.User, len(users))
	for _, user := range users {
		if user.AvatarURL == "" {
			user.AvatarURL = u.defaultAvatarURL
		}
		result[user.ID] = user
	}

	return result, nil
}

func (u *userUseCase) GetByPublicID(ctx context.Context, publicID string) (domain.User, error) {
	user, err := u.userRepo.GetUserByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, errutil.Internal("failed to get user by public id", err)
	}

	if user.AvatarURL == "" {
		user.AvatarURL = u.defaultAvatarURL
	}

	return user, nil
}
