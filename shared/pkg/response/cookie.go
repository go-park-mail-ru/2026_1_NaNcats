package response

import (
	"net/http"
	"time"
)

func SetCookie(w http.ResponseWriter, name, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,                 // имя куки
		Value:    value,                // значение - случайный идентификатор из usecase
		Expires:  expiresAt,            // срок жизни
		HttpOnly: true,                 // защита: JavaScript(фронт) не сможет прочитать эту куку
		Path:     "/",                  // кука будет отправляться на все эндпоинты сайтаx
		SameSite: http.SameSiteLaxMode, // защита от CSRF атак
		Secure: true,				
	})
}
