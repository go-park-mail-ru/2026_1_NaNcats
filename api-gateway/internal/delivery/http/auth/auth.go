package auth

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/validatorutil"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

//easyjson:json
type RegisterRequest struct {
	Name     string `json:"name" example:"Иван"`
	Email    string `json:"email" example:"user@mail.ru"`
	Password string `json:"password" example:"qwerty12345" validate:"min=8,max=128"`
}

func (r *RegisterRequest) Sanitize(p *bluemonday.Policy) {
	r.Name = p.Sanitize(r.Name)
}

//easyjson:json
type RegisterResponse struct {
	Name      string    `json:"name" example:"Иван"`
	Email     string    `json:"email" example:"user@mail.ru"`
	CreatedAt time.Time `json:"created_at" example:"2006-01-02T15:04:05Z07:00"`
	CSRFToken string    `json:"csrf_token"`
}

//easyjson:json
type LoginRequest struct {
	Login    string `json:"login" example:"user@mail.ru"`
	Password string `json:"password" example:"qwerty12345" validate:"min=8,max=128"`
}

//easyjson:json
type LoginResponse struct {
	// PublicID — публичный UUID пользователя. Внутренний числовой id наружу
	// не отдаём: его утечка — это раскрытие внутренних идентификаторов.
	PublicID    string `json:"public_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name        string `json:"name" example:"Иван"`
	Email       string `json:"email" example:"ivan@example.com"`
	AvatarURL   string `json:"avatar_url" example:"users/avatars/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
	CSRFToken   string `json:"csrf_token,omitempty"`
	StreakWeeks int32  `json:"streak_weeks,omitempty"`
}

//easyjson:json
type CSRFResponse struct {
	CSRFToken string `json:"csrf_token"`
	Message   string `json:"message,omitempty"`
}

type AuthHandler struct {
	authClient authclient.AuthClient
	userClient userclient.UserClient
	logger     logger.Logger
	validate   *validator.Validate
}

func NewAuthHandler(ac authclient.AuthClient, uc userclient.UserClient, l logger.Logger, v *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authClient: ac,
		userClient: uc,
		logger:     l,
		validate:   v,
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
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var reqDTO RegisterRequest
	if err := request.JSON(r, &reqDTO); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validate.Struct(reqDTO); err != nil {
		errMsg := validatorutil.FormatValidationError(err)
		response.Error(w, http.StatusBadRequest, errMsg)
		return
	}

	userResp, err := h.userClient.CreateUser(ctx, reqDTO.Name, reqDTO.Email, reqDTO.Password, idemKey)
	if err != nil {
		if errors.Is(err, userclient.ErrEmailAlreadyExists) {
			response.Error(w, http.StatusConflict, "Email already exists")
			return
		}
		l.Error("failed to create user", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	session, err := h.authClient.IssueSession(ctx, userResp, "user", r.UserAgent())
	if err != nil {
		l.Error("failed to issue session", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	csrfToken, err := h.authClient.SetCSRF(ctx, session.Id)
	if err != nil {
		l.Error("failed to save csrf", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.SetCookie(w, "session_id", session.Id, session.ExpiresAt.AsTime())

	response.JSON(w, http.StatusCreated, RegisterResponse{
		Name:      reqDTO.Name,
		Email:     reqDTO.Email,
		CreatedAt: time.Now(),
		CSRFToken: csrfToken,
	})
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
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	var reqDTO LoginRequest
	if err := request.JSON(r, &reqDTO); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.validate.Struct(reqDTO); err != nil {
		errMsg := validatorutil.FormatValidationError(err)
		response.Error(w, http.StatusBadRequest, errMsg)
		return
	}

	session, err := h.authClient.Login(ctx, reqDTO.Login, reqDTO.Password, r.UserAgent())
	if err != nil {
		if errors.Is(err, authclient.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		l.Error("login failed", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	userResp, err := h.userClient.GetByID(ctx, session.UserId)
	if err != nil {
		l.Error("failed to fetch user profile after login", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	csrfToken, err := h.authClient.SetCSRF(ctx, session.Id)
	if err != nil {
		l.Error("failed to save csrf", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.SetCookie(w, "session_id", session.Id, session.ExpiresAt.AsTime())

	resp := LoginResponse{
		PublicID:  userResp.PublicId,
		Name:      userResp.Name,
		Email:     userResp.Email,
		AvatarURL: userResp.AvatarUrl,
		CSRFToken: csrfToken,
	}
	if _, profile, perr := h.userClient.GetUserProfile(ctx, session.UserId); perr == nil && profile != nil {
		resp.StreakWeeks = profile.StreakCount
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
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		_ = h.authClient.Logout(r.Context(), cookie.Value)
	}

	response.SetCookie(w, "session_id", "", time.Unix(0, 0))
	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
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
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userResp, err := h.userClient.GetByID(ctx, userID)
	if err != nil {
		l.Error("failed to fetch profile for GetMe", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := LoginResponse{
		PublicID:  userResp.PublicId,
		Name:      userResp.Name,
		Email:     userResp.Email,
		AvatarURL: userResp.AvatarUrl,
	}
	if _, profile, perr := h.userClient.GetUserProfile(ctx, userID); perr == nil && profile != nil {
		resp.StreakWeeks = profile.StreakCount
	}
	response.JSON(w, http.StatusOK, resp)
}

// GetCSRF godoc
// @Summary         Получение CSRF токена
// @Description     Возвращает текущий CSRF токен пользователя на основе его сессии. Если сессия не найдена, возвращает соответствующее сообщение.
// @Tags            auth
// @Produce         json
// @Success         200     {object}  CSRFResponse           "Успешное получение токена или сообщение об отсутствии сессии"
// @Failure         500     {object}  response.ErrorResponse "Внутренняя ошибка сервера"
// @Router          /csrf [get]
func (h *AuthHandler) GetCSRF(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		response.JSON(w, http.StatusOK, CSRFResponse{Message: "no session"})
		return
	}

	csrfToken, err := h.authClient.GetCSRF(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, authclient.ErrSessionNotFound) {
			response.Error(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, CSRFResponse{CSRFToken: csrfToken})
}
