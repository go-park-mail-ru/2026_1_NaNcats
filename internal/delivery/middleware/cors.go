package middleware

import "net/http"

type CORSMiddleware struct {
	allowedOrigins []string // мапа разрешенных источников
}

func NewCORSMiddleware(origins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: origins,
	}
}

func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		isAllowed := false // флаг, есть ли origin среди разрешенных
		for _, o := range m.allowedOrigins {
			if origin == o {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			// сам источник
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// разрешаем браузеру отправлять и получать куки
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// заголовки, которые фронтенду разрешено присылать
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			// методы, которые разрешаем присылать фронтенду
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		}

		// обработка preflight запросов - т.е. проверки браузера на cors
		if r.Method == http.MethodOptions {
			// принято отвечать на preflight запросы 204 No Content
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// передаем запрос дальше
		next.ServeHTTP(w, r)
	})
}
