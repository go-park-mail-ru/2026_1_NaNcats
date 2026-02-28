package httphandler

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

type RegisterRequest struct {
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// здесь должна быть дальнейшая бизнес-логика
}
