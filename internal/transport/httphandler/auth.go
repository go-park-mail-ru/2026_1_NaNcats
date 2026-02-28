package httphandler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// здесь должна быть дальнейшая бизнес-логика
}
