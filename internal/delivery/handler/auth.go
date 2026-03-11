package handler

import (
	"errors"
	"log"
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
type RegisterRequest struct {
	// Имя пользователя
	Name string `json:"name" example:"Иван"`
	// Email пользователя
	Email string `json:"email" example:"user@mail.ru"`
	// Пароль в открытом виде
	Password string `json:"password" example:"qwerty12345"`
}

// DTO отправки сведений о пользователе при регистрации
type RegisterResponse struct {
	// Уникальный ID пользователя в системе
	ID uuid.UUID `json:"id" example:"1"`
	// Имя для отображения в интерфейсе
	Name string `json:"name" example:"Иван"`
	// Email пользователя
	Email string `json:"email" example:"user@mail.ru"`
	// Время создания аккаунта по стандарту RFC 3339
	CreatedAt time.Time `json:"created_at" example:"2006-01-02T15:04:05Z07:00"`
}

// LoginRequest - DTO для входящего запроса на авторизацию
type LoginRequest struct {
	// Email пользователя
	Login string `json:"login" example:"user@mail.ru"`
	// Пароль в открытом виде
	Password string `json:"password" example:"qwerty12345"`
}

// LoginResponse - DTO для ответа при успешном входе
type LoginResponse struct {
	// Уникальный ID пользователя в системе
	ID uuid.UUID `json:"id" example:"1"`
	// Имя для отображения в интерфейсе
	Name string `json:"name" example:"Иван"`
	// URL аватарки пользователя
	AvatarURL string `json:"avatar_url" example:"users/avatars/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
}

// структура хендлера авторизации
type authHandler struct {
	authUC usecase.AuthUseCase
}

// функция-конструтор хендлера
func NewAuthHandler(auc usecase.AuthUseCase) *authHandler {
	return &authHandler{
		authUC: auc,
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
	// объект DTO запроса
	curRequest := RegisterRequest{}

	// заполняем объект DTO запроса данными из запроса
	err := request.JSON(r, &curRequest)
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

	// контекст нынешнего запроса, позволяет досрочно завершить бизнес-логику
	// если пользователь отключится/отменит загрузку запроса
	ctx := r.Context()

	createdUser, createdSession, err := h.authUC.Register(ctx, userToCreate)
	if err != nil {
		switch {
		// Клиентские ошибки (400 Bad Request)
		case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidPassword):
			response.Error(w, http.StatusBadRequest, err.Error())

		// Ошибка конфликта (409 Conflict)
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			response.Error(w, http.StatusConflict, err.Error())

		// Системные ошибки (500 Internal Server Error)
		default:
			// Мы пишем подробности в лог, но клиенту отдаем общую фразу
			log.Printf("[ERROR] Register failed: %v", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.SetCookie(w, "session_id", createdSession.ID.String(), createdSession.ExpiresAt)

	// ответ, который отдаем юзеру
	resp := RegisterResponse{
		ID:        createdUser.ID,
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
	curRequest := LoginRequest{}

	err := request.JSON(r, &curRequest)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userToLogin := domain.User{
		Email:        curRequest.Login,
		PasswordHash: curRequest.Password,
	}

	ctx := r.Context()

	loggedUser, createdSession, err := h.authUC.Login(ctx, userToLogin)
	if err != nil {
		switch {
		// Ошибка авторизации (401)
		case errors.Is(err, domain.ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "Invalid email or password")

		// Системные ошибки (500)
		default:
			log.Printf("[ERROR] Login failed: %v", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.SetCookie(w, "session_id", createdSession.ID.String(), createdSession.ExpiresAt)

	resp := LoginResponse{
		ID:   loggedUser.ID,
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
// @Failure			401		{object}  response.ErrorResponse	"Кука не найдена"
// @Router			/auth/logout [post]
func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// Если куки нет, пользователь и так "вышел". Возвращаем 200
		response.JSON(w, http.StatusOK, nil)
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err == nil {
		// Если UUID валидный, пытаемся удалить сессию в бизнес-логике.
		ctx := r.Context()
		err = h.authUC.Logout(ctx, sessionID)
		if err != nil {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
	}

	// нулевое время в эпохе Unix
	// в любом случае принудительно удаляем куку у пользователя
	response.SetCookie(w, "session_id", "", time.Unix(0, 0))
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

	// берем userID из контекста, который нам пришел из мидлвара AuthMiddleware
	// Value возвращает any. Используем утверждение типа, чтобы Go знал что это uuid
	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	// если там не int или nil
	if !ok {
		log.Printf("[ERROR] GetMe called without AuthMiddleware")
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	loggedUser, err := h.authUC.GetProfile(ctx, userID)
	if err != nil {
		// Если мидлваря пропустила сессию, значит юзер в базе точно должен быть
		// Если его нет - это системная проблема или критический сбой
		log.Printf("[ERROR] GetProfile failed for user %v: %v", userID, err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := LoginResponse{
		ID:   loggedUser.ID,
		Name: loggedUser.Name,
	}

	response.JSON(w, http.StatusOK, resp)
}
