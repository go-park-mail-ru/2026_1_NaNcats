package handler

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

type RegisterRequest struct {
	// DTO запроса на регистрацию
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	// DTO отправки сведений о пользователе
	ID        int       `json:"id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type authHandler struct {
	authUC usecase.AuthUseCase
}

func NewAuthHandler(auc usecase.AuthUseCase) *authHandler {
	return &authHandler{
		authUC: auc,
	}
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	// метод хендлера authHandler, нужен для обработки регистрации по запросу /register
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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
		Phone:        curRequest.Phone,
		Name:         curRequest.Name,
		Email:        curRequest.Email,
		PasswordHash: curRequest.Password,
	}

	// контекст нынешнего запроса, позволяет досрочно завершить бизнес-логику
	// если пользователь отключится/отменит загрузку запроса
	ctx := r.Context()

	// созданный юзер, id сессии
	createdUser, sessionID, expiresAt, err := h.authUC.Register(ctx, userToCreate)
	if err != nil {
		// добавить больше бизнес-ошибок (не только bad request)
		response.Error(w, http.StatusBadRequest, err.Error())
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",         // имя куки
		Value:    sessionID,            // значение - случайный идентификатор из usecase
		Expires:  expiresAt,            // срок жизни
		HttpOnly: true,                 // защита: JavaScript(фронт) не сможет прочитать эту куку
		Path:     "/",                  // кука будет отправляться на все эндпоинты сайта
		SameSite: http.SameSiteLaxMode, // защита от CSRF атак
		// Secure: true,				// если будем на https
	})

	// ответ, который отдаем юзеру
	resp := UserResponse{
		ID:        createdUser.ID,
		Phone:     createdUser.Phone,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
	}

	response.JSON(w, http.StatusCreated, resp)
}
