package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/redisrepo"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "github.com/go-park-mail-ru/2026_1_NaNcats/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title 		NaNcats Delivery API
// @version 	1.0
// @description API бэкенда для проекта "Delivery Club" от команды NaNcats
// @host		localhost:8080
// @BasePath	/api
func main() {
	// Пытаемся загрузить .env файл только для локальной разработки
	// В Docker переменные прокинутся сами через docker-compose
	_ = godotenv.Load()

	port := os.Getenv("PORT") // выделенный под сервер порт из окружения
	if port == "" {
		port = "8080"
	}

	// "чистый" логгер из pkg
	rawLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatalf("Connot start without logger: %v", err)
	}

	// Оборачиваем его в адаптер, который реализует domain.Logger
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)

	ctx := context.Background()

	redisAddr := flag.String("addr", "redis://user:@localhost:6379/0", "redis addr")

	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(*redisAddr)
		},
	}
	defer redisPool.Close()

	// Получаем URL из переменной окружения (которая прописана в docker-compose)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}

	config.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)

	// Открываем соединение с БД
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Проверяем соединение с БД
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Could not ping the database: %v\n", err)
	}

	// Запускаем миграции
	err = postgres.RunMigrations(dbURL)
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations applied successfully")

	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := redisrepo.NewSessionRepo(redisPool)
	restaurantBrandRepo := postgres.NewRestaurantBrandRepo(pool)

	// ttl сессии - 24 часа
	sessionTTL := 24 * time.Hour

	userUC := usecase.NewUserUseCase(userRepo)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, sessionTTL)
	authUC := usecase.NewAuthUseCase(userUC, sessionUC)
	restaurantBrandUC := usecase.NewRestaurantBrandUseCase(restaurantBrandRepo)
	userProfileUC := usecase.NewUserProfileUseCase(userUC)

	authHandler := handler.NewAuthHandler(authUC, userUC, appLogger)
	restaurantBrandHandler := handler.NewRestaurantBrandHandler(restaurantBrandUC, appLogger)
	userProfileHandler := handler.NewUserProfileHandler(userProfileUC, userUC, sessionUC, appLogger)

	fileServer := http.FileServer(http.Dir("./uploads"))

	authMW := middleware.NewAuthMiddleware(sessionUC, appLogger)
	corsMW := middleware.NewCORSMiddleware([]string{
		"http://localhost:2033",
	})
	requestIDMW := middleware.NewRequestIDMiddleware()
	loggingMW := middleware.NewLoggingMiddleware(appLogger)

	// создание собственного роутера
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	// ручка, которую дергаем для проверки авторизации по куки с миддлваром на авторизацию
	mux.Handle("GET /api/auth/me", authMW.RequireAuth(http.HandlerFunc(authHandler.GetMe)))

	mux.Handle("GET /api/images/", http.StripPrefix("/api/images", fileServer))

	mux.HandleFunc("GET /api/restaurants/brands", restaurantBrandHandler.GetRestaurantBrandsList)

	mux.Handle("GET /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.GetUserProfile)))
	mux.Handle("PATCH /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.UpdateProfile)))

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// применение глобальных мидлваров, применяются снизу вверх
	handler := corsMW.Handler(mux)
	handler = loggingMW.Handler(handler)
	handler = requestIDMW.Handler(handler)

	log.Printf("Server is starting on port %s...", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler, // передаем обернутый роутер
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
