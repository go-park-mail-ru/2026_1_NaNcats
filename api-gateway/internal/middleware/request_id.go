package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/google/uuid"
)

type RequestIDMiddleware struct {
}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// Мидлваря, которая добавляет id к каждому запросу
func (m *RequestIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")

		if reqID == "" {
			reqID = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), common.RequestIDKey, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
