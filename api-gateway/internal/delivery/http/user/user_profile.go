package user

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	auth "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth"
	user "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

//easyjson:json
type UserProfileUpdateRequest struct {
	Name  *string `json:"name" example:"Андрей"`
	Email *string `json:"email" example:"new_mail@gmail.com"`
}

//easyjson:json
type UpdateAvatarResponse struct {
	Message   string `json:"message"`
	AvatarURL string `json:"avatar_url"`
}

//easyjson:json
type UserProfileResponse struct {
	Name      string `json:"name" example:"Илья"`
	Email     string `json:"email" example:"terminator2007@gmail.com"`
	AvatarURL string `json:"avatar_url" example:"users/avatars/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
}

type userProfileHandler struct {
	userProfileUC user.UserProfileUseCase
	userUC        user.UserUseCase
	sessionUC     auth.SessionUseCase
	logger        logger.Logger
}

func NewUserProfileHandler(upuc user.UserProfileUseCase, uuc user.UserUseCase, suc auth.SessionUseCase, logger logger.Logger) *userProfileHandler {
	return &userProfileHandler{
		userProfileUC: upuc,
		userUC:        uuc,
		sessionUC:     suc,
		logger:        logger,
	}
}

// GetUserProfile godoc
// @Summary 		Получение профиля пользователя
// @Description		Возвращает данные профиля (имя и email) текущего авторизованного пользователя
// @Tags			profile
// @Accept			json
// @Produce			json
// @Success			200		{object}  UserProfileResponse		"Успешное получение данных профиля"
// @Failure			404		{object}  response.ErrorResponse	"Пользователь не найден"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile [get]
func (h *userProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	userProfile, err := h.userProfileUC.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			l.Warn("user profile not found", logger.Int("user_id", userID))
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		l.Error("failed to fetch user profile", err, logger.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "server error while parsing query")
		return
	}

	resp := UserProfileResponse{
		Name:      userProfile.Name,
		Email:     userProfile.Email,
		AvatarURL: userProfile.AvatarURL,
	}

	response.JSON(w, http.StatusOK, resp)
}

// UpdateProfile godoc
// @Summary 		Обновление данных профиля
// @Description		Частично обновляет данные профиля текущего пользователя (имя и/или email)
// @Tags			profile
// @Accept			json
// @Produce			json
// @Param			input	body	  UserProfileUpdateRequest	true	"Данные для обновления профиля"
// @Success			200		{object}  map[string]string			"Профиль успешно обновлен"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка валидации JSON или нет данных для обновления"
// @Failure			409		{object}  response.ErrorResponse	"Указанный email уже используется другим пользователем"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile [patch]
func (h *userProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	curRequest := UserProfileUpdateRequest{}
	err = request.JSON(r, &curRequest)
	if err != nil {
		l.Warn("failed to decode update profile request", logger.Int("user_id", userID), logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.userUC.UpdateProfile(ctx, userID, curRequest.Name, curRequest.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			l.Info("profile update conflict: email already in use", logger.Int("user_id", userID))
			response.Error(w, http.StatusConflict, "email already in use")
		case errors.Is(err, domain.ErrEmptyDBQuery):
			l.Warn("profile update failed: no data to update", logger.Int("user_id", userID))
			response.Error(w, http.StatusBadRequest, "no data to update")
		default:
			l.Error("failed to update profile", err, logger.Int("user_id", userID))
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	l.Info("profile updated successfully", logger.Int("user_id", userID))
	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "profile uptade succeed"})
}

// UpdateAvatar godoc
// @Summary 		Обновление аватара пользователя
// @Description		Загружает и обновляет аватар текущего авторизованного пользователя. Принимает multipart/form-data with полем 'avatar'.
// @Tags			profile
// @Accept			multipart/form-data
// @Produce			json
// @Param			avatar	formData  file						true	"Файл аватара (WEBP/JPG/JPEG/PNG, до 5МБ)"
// @Success			200		{object}  UpdateAvatarResponse		"Аватар успешно обновлен"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка запроса (файл слишком большой, неверный формат или отсутствует)"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/avatar [post]
func (h *userProfileHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		l.Warn("avatar upload failed: file too large", logger.Int("user_id", userID))
		response.Error(w, http.StatusBadRequest, "file is too large (max 5MB)")
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		l.Warn("avatar upload failed: missing avatar field", logger.Int("user_id", userID))
		response.Error(w, http.StatusBadRequest, "failed to get 'avatar' field from form")
		return
	}
	defer file.Close()

	if fileHeader.Size > (5 << 20) {
		l.Warn("avatar upload failed: size limit exceeded", logger.Int("user_id", userID), logger.Int("size", int(fileHeader.Size)))
		response.Error(w, http.StatusBadRequest, "file size larger than 5MB limit")
		return
	}

	newAvatarURL, err := h.userUC.UpdateAvatar(ctx, userID, file)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidImageExt) {
			l.Warn("avatar upload failed: invalid format", logger.Int("user_id", userID))
			response.Error(w, http.StatusBadRequest, "unsupported image format (only JPEG/PNG allowed)")
			return
		}
		l.Error("failed to update avatar in storage", err, logger.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	l.Info("avatar updated successfully", logger.Int("user_id", userID), logger.String("url", newAvatarURL))
	response.JSON(w, http.StatusOK, UpdateAvatarResponse{
		Message:   "avatar updated successfully",
		AvatarURL: newAvatarURL,
	})
}

// DeleteAvatar godoc
// @Summary 		Удаление аватара пользователя
// @Description		Удаляет аватар текущего авторизованного пользователя и устанавливает аватар по умолчанию.
// @Tags			profile
// @Produce			json
// @Success			200		{object}  UpdateAvatarResponse		"Аватар успешно удален"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/avatar [delete]
func (h *userProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	newAvatarURL, err := h.userUC.DeleteAvatar(ctx, userID)
	if err != nil {
		l.Error("failed to delete avatar", err, logger.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	l.Info("avatar deleted successfully", logger.Int("user_id", userID))
	response.JSON(w, http.StatusOK, UpdateAvatarResponse{
		Message:   "avatar deleted successfully",
		AvatarURL: newAvatarURL,
	})
}
