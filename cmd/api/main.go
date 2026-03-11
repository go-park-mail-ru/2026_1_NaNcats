package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
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
	restaurantBrandRepo := memory.NewRestaurantBrandRepo()

	// ttl сессии - 24 часа
	sessionTTL := 24 * time.Hour

	sessionUC := usecase.NewSessionUseCase(sessionRepo, sessionTTL)
	authUC := usecase.NewAuthUseCase(userRepo, sessionUC)
	restaurantBrandUC := usecase.NewRestaurantBrandUseCase(restaurantBrandRepo)

	authHandler := handler.NewAuthHandler(authUC)
	restaurantBrandHandler := handler.NewRestaurantBrandHandler(restaurantBrandUC)

	fileServer := http.FileServer(http.Dir("./uploads"))

	authMW := middleware.NewAuthMiddleware(sessionUC)
	corsMW := middleware.NewCORSMiddleware([]string{
		"http://localhost:2033",
	})

	// создание собственного роутера
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	// ручка, которую дергаем для проверки авторизации по куки с миддлваром на авторизацию
	mux.Handle("GET /api/auth/me", authMW.RequireAuth(http.HandlerFunc(authHandler.GetMe)))

	mux.Handle("GET /api/images/", http.StripPrefix("/api/images", fileServer))

	mux.HandleFunc("GET /api/restaurants/brands", restaurantBrandHandler.GetRestaurantBrandsList)

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// применение глобальных мидлваров
	siteHandler := corsMW.Handler(mux)

	log.Printf("Server is starting on port %s...", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      siteHandler, // передаем обернутый роутер
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
