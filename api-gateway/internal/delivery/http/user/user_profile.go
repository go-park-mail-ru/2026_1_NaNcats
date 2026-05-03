package user

//go:generate easyjson $GOFILE

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
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
	AvatarURL string `json:"avatar_url" example:"users/avatars/fjaun99f.png"`
}

type UserProfileHandler struct {
	userClient userclient.UserClient
	logger     logger.Logger
}

func NewUserProfileHandler(uc userclient.UserClient, l logger.Logger) *UserProfileHandler {
	return &UserProfileHandler{
		userClient: uc,
		logger:     l,
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
func (h *UserProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, _, err := h.userClient.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, userclient.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		l.Error("failed to fetch user profile", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, UserProfileResponse{
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarUrl,
	})
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
func (h *UserProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var reqDTO UserProfileUpdateRequest
	if err := request.JSON(r, &reqDTO); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.userClient.UpdateProfile(ctx, userID, reqDTO.Name, reqDTO.Email, idemKey)
	if err != nil {
		switch {
		case errors.Is(err, userclient.ErrEmailAlreadyExists):
			response.Error(w, http.StatusConflict, "email already in use")
		case errors.Is(err, userclient.ErrInvalidArgument):
			response.Error(w, http.StatusBadRequest, "no data to update or invalid data")
		default:
			l.Error("failed to update profile", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "profile update succeed"})
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
func (h *UserProfileHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "file is too large (max 5MB)")
		return
	}

	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to get 'avatar' field from form")
		return
	}
	defer file.Close()

	if fileHeader.Size > (5 << 20) {
		response.Error(w, http.StatusBadRequest, "file size larger than 5MB limit")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		l.Error("failed to read avatar file", err)
		response.Error(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	newAvatarURL, err := h.userClient.UpdateAvatar(ctx, userID, fileBytes, idemKey)
	if err != nil {
		if errors.Is(err, userclient.ErrInvalidArgument) {
			response.Error(w, http.StatusBadRequest, "unsupported image format")
			return
		}
		l.Error("failed to update avatar via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

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
func (h *UserProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	newAvatarURL, err := h.userClient.DeleteAvatar(ctx, userID, idemKey)
	if err != nil {
		l.Error("failed to delete avatar", err)
		response.Error(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	response.JSON(w, http.StatusOK, UpdateAvatarResponse{
		Message:   "avatar deleted successfully",
		AvatarURL: newAvatarURL,
	})
}

// easyjson:json
type UpdateRoleRequest struct {
	UserID  int64  `json:"user_id"`
	NewRole string `json:"new_role"`
}

// AdminUpdateRole godoc
// @Summary 		Смена роли пользователя (админ)
// @Tags			admin
// @Accept			json
// @Produce			json
// @Param			Idempotency-Key header string true "Ключ"
// @Param			input	body	  UpdateRoleRequest	true	"Данные"
// @Router			/admin/users/role [post]
func (h *UserProfileHandler) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req UpdateRoleRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.userClient.UpdateRole(ctx, req.UserID, req.NewRole, idemKey)
	if err != nil {
		l.Error("failed to update user role via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "role updated successfully"})
}
