package handler

//go:generate easyjson $GOFILE

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
type UserProfileResponse struct {
	Name  string `json:"name" example:"Илья"`
	Email string `json:"email" example:"terminator2007@gmail.com"`
}

type userProfileHandler struct {
	userProfileUC usecase.UserProfileUseCase
	userUC        usecase.UserUseCase
	sessionUC     usecase.SessionUseCase
	logger        domain.Logger
}

func NewUserProfileHandler(upuc usecase.UserProfileUseCase, uuc usecase.UserUseCase, suc usecase.SessionUseCase, logger domain.Logger) *userProfileHandler {
	return &userProfileHandler{
		userProfileUC: upuc,
		userUC:        uuc,
		sessionUC:     suc,
		logger:        logger,
	}
}

func (h *userProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
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

	resp := UserProfileResponse{
		Name:  userProfile.Name,
		Email: userProfile.Email,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *userProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "unauthorized or missing context")
		return
	}

	curRequest := UserProfileUpdateRequest{}

	err := request.JSON(r, &curRequest, l)
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
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "profile uptade succeed"})
}
