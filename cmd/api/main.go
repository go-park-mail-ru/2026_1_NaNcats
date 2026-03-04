package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/memory"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
)

func main() {
	port := os.Getenv("PORT") // выделенный под сервер порт из окружения
	if port == "" {
		port = "8080"
	}

	userRepo := memory.NewUserRepo()

	sessionRepo := memory.NewSessionRepo()

	// ttl сессии - 30 дней
	sessionTTL := 30 * 24 * time.Hour

	sessionUC := usecase.NewSessionUseCase(sessionRepo, sessionTTL)

	authUC := usecase.NewAuthUseCase(userRepo, sessionUC)

	authHandler := handler.NewAuthHandler(authUC)

	http.HandleFunc("/api/register", authHandler.Register)

	// добавить роут api/login

	log.Printf("Server is starting on port %s...", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
