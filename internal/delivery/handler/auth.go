package handler

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
)

// DTO запроса на регистрацию
//
//easyjson:json
type RegisterRequest struct {
	// Имя пользователя
	Name string `json:"name" example:"Иван"`
	// Email пользователя
	Email string `json:"email" example:"user@mail.ru"`
	// Пароль в открытом виде
	Password string `json:"password" example:"qwerty12345"`
}

// DTO отправки сведений о пользователе при регистрации
//
//easyjson:json
type RegisterResponse struct {
	// Имя для отображения в интерфейсе
	Name string `json:"name" example:"Иван"`
	// Email пользователя
	Email string `json:"email" example:"user@mail.ru"`
	// Время создания аккаунта по стандарту RFC 3339
	CreatedAt time.Time `json:"created_at" example:"2006-01-02T15:04:05Z07:00"`
}

// LoginRequest - DTO для входящего запроса на авторизацию
//
//easyjson:json
type LoginRequest struct {
	// Email пользователя
	Login string `json:"login" example:"user@mail.ru"`
	// Пароль в открытом виде
	Password string `json:"password" example:"qwerty12345"`
}

// LoginResponse - DTO для ответа при успешном входе
//
//easyjson:json
type LoginResponse struct {
	// Имя для отображения в интерфейсе
	Name string `json:"name" example:"Иван"`
	// URL аватарки пользователя
	AvatarURL string `json:"avatar_url" example:"users/avatars/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
}

// структура хендлера авторизации
type authHandler struct {
	authUC usecase.AuthUseCase
	userUC usecase.UserUseCase
	logger domain.Logger
}

// функция-конструтор хендлера
func NewAuthHandler(auc usecase.AuthUseCase, uuc usecase.UserUseCase, logger domain.Logger) *authHandler {
	return &authHandler{
		authUC: auc,
		userUC: uuc,
		logger: logger,
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
	// контекст нынешнего запроса, позволяет досрочно завершить бизнес-логику
	// если пользователь отключится/отменит загрузку запроса
	ctx := r.Context()

	l := h.logger.WithContext(ctx)

	// объект DTO запроса
	curRequest := RegisterRequest{}

	// заполняем объект DTO запроса данными из запроса
	err := request.JSON(r, &curRequest, l)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// структура, в которую кладем данные создаваемого юзера из запроса
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
			l.Info("registration validation failed", map[string]any{
				"email": curRequest.Email,
				"error": err.Error(),
			})
			response.Error(w, http.StatusBadRequest, err.Error())

		// Ошибка конфликта (409 Conflict)
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			l.Info("registration conflict: email already exists", map[string]any{
				"email": curRequest.Email,
			})
			response.Error(w, http.StatusConflict, err.Error())

		// Системные ошибки (500 Internal Server Error)
		default:
			l.Error("registration failed unexpectedly", err, map[string]any{
				"email": curRequest.Email,
			})
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// успех
	l.Info("user registered successfully", map[string]any{
		"user_id": createdUser.ID,
		"email":   createdUser.Email,
	})

	response.SetCookie(w, "session_id", createdSession.ID.String(), createdSession.ExpiresAt)

	// ответ, который отдаем юзеру
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

	err := request.JSON(r, &curRequest, l)
	if err != nil {
		l.Info("failed to decode login request", map[string]any{"error": err.Error()})
		response.Error(w, http.StatusBadRequest, err.Error())
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
		// Ошибка авторизации (401)
		case errors.Is(err, domain.ErrInvalidCredentials):
			l.Info("login failed: invalid credentials", map[string]any{"email": curRequest.Login})
			response.Error(w, http.StatusUnauthorized, "Invalid email or password")

		// Системные ошибки (500)
		default:
			l.Error("login failed unexpectedly", err, map[string]any{"email": curRequest.Login})
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Успещный успех
	l.Info("user logged in successfully", map[string]any{
		"user_id": loggedUser.ID,
		"email":   loggedUser.Email,
	})

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
		// Если куки нет, пользователь и так "вышел". Возвращаем 200
		l.Info("logout: no session cookie found, user already logged out", nil)
		response.JSON(w, http.StatusOK, nil)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		// Токен кривой. Сессию в базе искать нет смысла,
		// но мы логируем это, чтобы видеть странную активность
		l.Warn("logout: invalid session token format", map[string]any{
			"token_value": cookie.Value,
		})
	} else {
		// Токен валидный, пробуем удалить из хранилища
		err = h.authUC.Logout(ctx, sessionID)
		if err != nil {
			// Даже если сессия не найдена в базе, просто логируем это как Info
			l.Info("logout: session not found in database or already expired", map[string]any{
				"session_id": sessionID.String(),
			})
		} else {
			l.Info("logout: session successfully removed from database", map[string]any{
				"session_id": sessionID.String(),
			})
		}
	}

	response.SetCookie(w, "session_id", "", time.Unix(0, 0))

	l.Info("logout: cookie cleared in browser", nil)
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

	if errors.Is(err, middleware.ErrNoUserIDInContext) {
		l.Error("auth context contract broken: auth middleware missed userID in route", err, map[string]any{
			"user_id": userID,
		})
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	loggedUser, err := h.userUC.GetByID(ctx, userID)
	if err != nil {
		// Если мидлваря пропустила сессию, значит юзер в базе точно должен быть
		// Если его нет - это системная проблема или критический сбой
		l.Error("get profile failed", err, map[string]any{
			"user_id": userID,
		})
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	l.Info("profile retrieved successfully", map[string]any{
		"user_id": loggedUser.ID,
	})

	resp := LoginResponse{
		Name: loggedUser.Name,
	}

	response.JSON(w, http.StatusOK, resp)
}
