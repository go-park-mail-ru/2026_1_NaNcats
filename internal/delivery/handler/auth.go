package handler

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
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
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// здесь должна быть дальнейшая бизнес-логика
}
