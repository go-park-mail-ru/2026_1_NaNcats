package handler

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/validatorutil"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

//easyjson:json
type RegisterRequest struct {
	Name     string `json:"name" example:"Иван"`
	Email    string `json:"email" example:"user@mail.ru"`
	Password string `json:"password" example:"qwerty12345" validate:"min=8,max=128"`
}

//easyjson:json
type RegisterResponse struct {
	Name      string    `json:"name" example:"Иван"`
	Email     string    `json:"email" example:"user@mail.ru"`
	CreatedAt time.Time `json:"created_at" example:"2006-01-02T15:04:05Z07:00"`
}

//easyjson:json
type LoginRequest struct {
	Login    string `json:"login" example:"user@mail.ru"`
	Password string `json:"password" example:"qwerty12345" validate:"min=8,max=128"`
}

//easyjson:json
type LoginResponse struct {
	Name      string `json:"name" example:"Иван"`
	AvatarURL string `json:"avatar_url" example:"users/avatars/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
}

type authHandler struct {
	authUC   usecase.AuthUseCase
	userUC   usecase.UserUseCase
	logger   domain.Logger
	validate *validator.Validate
}

func NewAuthHandler(auc usecase.AuthUseCase, uuc usecase.UserUseCase, logger domain.Logger, v *validator.Validate) *authHandler {
	return &authHandler{
		authUC:   auc,
		userUC:   uuc,
		logger:   logger,
		validate: v,
	}
}

// Register godoc
// @Summary 		Регистрация пользователя
// @Description		Проверяет данные, создает нового пользователя и устанавливает сессионную куку
// @Tags			auth
// @Accept			json
// @Produce			json
// @Param			input	body	  RegisterRequest	true	"Данные для регистрации"
// @Success			201		{object}  RegisterResponse			"Успешная регистрация"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка валидации (email/пароль)"
// @Failure			409		{object}  response.ErrorResponse	"Пользователь с такой почтой уже существует"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/auth/register [post]
func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	curRequest := RegisterRequest{}
	err := request.JSON(r, &curRequest)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.validate.Struct(curRequest); err != nil {
		errMsg := validatorutil.FormatValidationError(err)
		l.Warn("registration validation failed", domain.String("error", errMsg))
		response.Error(w, http.StatusBadRequest, errMsg)
		return
	}

	userToCreate := domain.User{
		Name:         curRequest.Name,
		Email:        curRequest.Email,
		PasswordHash: curRequest.Password,
	}

	userAgent := r.UserAgent()

	createdUser, createdSession, err := h.authUC.Register(ctx, userToCreate, userAgent)
	if err != nil {
		switch {
		// Клиентские ошибки (400 Bad Request)
		case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidPassword):
			l.Warn("registration business validation failed", domain.String("email", curRequest.Email), domain.String("error", err.Error()))
			response.Error(w, http.StatusBadRequest, err.Error())

		// Ошибка конфликта (409 Conflict)
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			l.Info("registration conflict: email already exists", domain.String("email", curRequest.Email))
			response.Error(w, http.StatusConflict, err.Error())

		// Системные ошибки (500 Internal Server Error)
		default:
			l.Error("registration failed unexpectedly", err, domain.String("email", curRequest.Email))
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	l.Info("user registered successfully", domain.Int("user_id", createdUser.ID), domain.String("email", createdUser.Email))

	response.SetCookie(w, "session_id", createdSession.ID.String(), createdSession.ExpiresAt)

	resp := RegisterResponse{
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
	}

	response.JSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary 		Авторизация пользователя
// @Description		Проверяет учетные данные и устанавливает сессионную куку
// @Tags			auth
// @Accept			json
// @Produce			json
// @Param			input	body	  LoginRequest	true	"Данные для входа"
// @Success			200		{object}  LoginResponse			"Успешный вход"
// @Failure			400		{object}  response.ErrorResponse	"Неверный формат JSON"
// @Failure			401		{object}  response.ErrorResponse	"Неверный логин или пароль"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/auth/login [post]
func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	curRequest := LoginRequest{}
	err := request.JSON(r, &curRequest)
	if err != nil {
		l.Warn("failed to decode login request", domain.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.validate.Struct(curRequest); err != nil {
		errMsg := validatorutil.FormatValidationError(err)
		l.Warn("login validation failed", domain.String("error", errMsg))
		response.Error(w, http.StatusBadRequest, errMsg)
		return
	}

	userToLogin := domain.User{
		Email:        curRequest.Login,
		PasswordHash: curRequest.Password,
	}

	userAgent := r.UserAgent()

	loggedUser, createdSession, err := h.authUC.Login(ctx, userToLogin, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			l.Info("login failed: invalid credentials", domain.String("email", curRequest.Login))
			response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		default:
			l.Error("login failed unexpectedly", err, domain.String("email", curRequest.Login))
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	l.Info("user logged in successfully", domain.Int("user_id", loggedUser.ID), domain.String("email", loggedUser.Email))

	response.SetCookie(w, "session_id", createdSession.ID.String(), createdSession.ExpiresAt)

	resp := LoginResponse{
		Name: loggedUser.Name,
	}

	response.JSON(w, http.StatusOK, resp)
}

// Logout godoc
// @Summary 		Выход из текущей
// @Description		Удаляет информацию о текущей сессии и принудительно протухает куку с сессией
// @Tags			auth
// @Accept			json
// @Produce			json
// @Success			200		"Успешный выход"
// @Router			/auth/logout [post]
func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	cookie, err := r.Cookie("session_id")
	if err != nil {
		l.Debug("logout: no session cookie found, user already logged out")
		response.JSON(w, http.StatusOK, nil)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		l.Warn("logout: invalid session token format", domain.String("token_value", cookie.Value))
	} else {
		err = h.authUC.Logout(ctx, sessionID)
		if err != nil {
			l.Debug("logout: session not found in database or already expired", domain.String("session_id", sessionID.String()))
		}
	}

	response.SetCookie(w, "session_id", "", time.Unix(0, 0))

	l.Debug("logout: session cleared")
	response.JSON(w, http.StatusOK, nil)
}

// GetMe godoc
// @Summary 		Проверка текущей сессии
// @Description		Возвращает данные профиля пользователя, если сессионная кука валидна
// @Tags			auth
// @Accept			json
// @Produce			json
// @Success			200		{object}  LoginResponse				"Успешный вход и создание сессии"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка"
// @Router			/auth/me [get]
func (h *authHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("auth middleware missed userID in context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	loggedUser, err := h.userUC.GetByID(ctx, userID)
	if err != nil {
		l.Error("get profile failed for authenticated user", err, domain.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	l.Debug("profile retrieved successfully", domain.Int("user_id", loggedUser.ID))

	avatarURL := loggedUser.AvatarURL
	if avatarURL == "" {
		avatarURL = os.Getenv("DEFAULT_AVATAR_URL")
	}

	resp := LoginResponse{
		Name:      loggedUser.Name,
		AvatarURL: avatarURL,
	}

	response.JSON(w, http.StatusOK, resp)
}
