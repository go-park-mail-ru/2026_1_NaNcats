package handler

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
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
	userProfileUC    usecase.UserProfileUseCase
	userUC           usecase.UserUseCase
	sessionUC        usecase.SessionUseCase
	logger           domain.Logger
	defaultAvatarURL string
}

func NewUserProfileHandler(upuc usecase.UserProfileUseCase, uuc usecase.UserUseCase, suc usecase.SessionUseCase, logger domain.Logger, defaultAvatarURL string) *userProfileHandler {
	return &userProfileHandler{
		userProfileUC:    upuc,
		userUC:           uuc,
		sessionUC:        suc,
		logger:           logger,
		defaultAvatarURL: defaultAvatarURL,
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

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	userProfile, err := h.userProfileUC.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "server error while parsing query")
		return
	}

	avatar := userProfile.AvatarURL
	if avatar == "" {
		avatar = h.defaultAvatarURL
	}

	resp := UserProfileResponse{
		Name:      userProfile.Name,
		Email:     userProfile.Email,
		AvatarURL: avatar,
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

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	curRequest := UserProfileUpdateRequest{}

	err = request.JSON(r, &curRequest)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.userUC.UpdateProfile(ctx, userID, curRequest.Name, curRequest.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			response.Error(w, http.StatusConflict, "email already in use")
		case errors.Is(err, domain.ErrEmptyDBQuery):
			response.Error(w, http.StatusBadRequest, "no data to update")
		default:
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "profile uptade succeed"})
}

// UpdateAvatar godoc
// @Summary 		Обновление аватара пользователя
// @Description		Загружает и обновляет аватар текущего авторизованного пользователя. Принимает multipart/form-data с полем 'avatar'.
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

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
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

	if fileHeader.Size > (5 << 20) {
		response.Error(w, http.StatusBadRequest, "file size larger than 5MB limit")
		return
	}

	newAvatarURL, err := h.userUC.UpdateAvatar(ctx, userID, file)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidImageExt) {
			response.Error(w, http.StatusBadRequest, "unsupported image format (only JPEG/PNG allowed)")
			return
		}
		h.logger.Error("failed to update avatar", err, map[string]any{})
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
func (h *userProfileHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	err = h.userUC.DeleteAvatar(ctx, userID)
	if err != nil {
		h.logger.Error("failed to delete avatar", err, map[string]any{})
		response.Error(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	response.JSON(w, http.StatusOK, UpdateAvatarResponse{
		Message:   "avatar deleted successfully",
		AvatarURL: h.defaultAvatarURL,
	})
}
