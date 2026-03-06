package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/memory"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"

	_ "github.com/go-park-mail-ru/2026_1_NaNcats/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title 		NaNcats Delivery API
// @version 	1.0
// @description API бэкенда для проекта "Delivery Club" от команды NaNcats
// @host		localhost:8080
// @BasePath	/api
func main() {
	port := os.Getenv("PORT") // выделенный под сервер порт из окружения
	if port == "" {
		port = "8080"
	}

	userRepo := memory.NewUserRepo()

	sessionRepo := memory.NewSessionRepo()

	// ttl сессии - 24 часа
	sessionTTL := 24 * time.Hour

	sessionUC := usecase.NewSessionUseCase(sessionRepo, sessionTTL)

	authUC := usecase.NewAuthUseCase(userRepo, sessionUC)

	authHandler := handler.NewAuthHandler(authUC)

	http.HandleFunc("POST /api/auth/register", authHandler.Register)
	http.HandleFunc("POST /api/auth/login", authHandler.Login)
	http.HandleFunc("GET /api/auth/me", authHandler.GetMe) // ручка, которую дергаем для проверки авторизации по куки

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Server is starting on port %s...", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
