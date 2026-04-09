package main

import (
	"context"
	"errors"
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
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/s3"
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

	// Читаем уровень логирования из переменной окружения
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // По умолчанию для прода
	}

	// "чистый" логгер из pkg
	rawLogger, err := logger.NewZapLogger(logLevel)
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
		appLogger.Fatal("database connection string is missing", errors.New("DATABASE_URL env var is empty"))
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}

	config.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)

	// Открываем соединение с БД
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		appLogger.Fatal("database pool creation failed", err)
	}
	defer pool.Close()

	// Проверяем соединение с БД
	if err := pool.Ping(ctx); err != nil {
		appLogger.Fatal("could not ping the database", err)
	}

	// Запускаем миграции
	err = postgres.RunMigrations(dbURL)
	if err != nil && err != migrate.ErrNoChange {
		appLogger.Fatal("failed to run migrations", err)
	}

	// S3
	keyID := os.Getenv("S3_KEY_ID")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucketName := "nancats-bucket"

	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := redisrepo.NewSessionRepo(redisPool)
	restaurantBrandRepo := postgres.NewRestaurantBrandRepo(pool)
	s3Repo, err := s3.NewS3Storage(ctx, keyID, secretKey, bucketName, "ru-central1")
	if err != nil {
		appLogger.Fatal("Failed to init S3", err)
	}

	// ttl сессии - 24 часа
	sessionTTL := 24 * time.Hour

	userUC := usecase.NewUserUseCase(userRepo, s3Repo)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, sessionTTL)
	authUC := usecase.NewAuthUseCase(userUC, sessionUC)
	restaurantBrandUC := usecase.NewRestaurantBrandUseCase(restaurantBrandRepo)
	userProfileUC := usecase.NewUserProfileUseCase(userUC)

	defaultAvatarURL := os.Getenv("DEFAULT_AVATAR_URL")
	if defaultAvatarURL == "" {
		appLogger.Warn("DEFAULT_AVATAR_URL пустой, фронтенд может упасть при запросе стандартного аватара", map[string]any{})
	}

	authHandler := handler.NewAuthHandler(authUC, userUC, appLogger)
	restaurantBrandHandler := handler.NewRestaurantBrandHandler(restaurantBrandUC, appLogger)
	userProfileHandler := handler.NewUserProfileHandler(userProfileUC, userUC, sessionUC, appLogger, defaultAvatarURL)

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

	mux.HandleFunc("GET /api/restaurants/brands", restaurantBrandHandler.GetRestaurantBrandsList)

	mux.Handle("GET /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.GetUserProfile)))
	mux.Handle("PATCH /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.UpdateProfile)))
	mux.Handle("POST /api/profile/avatar", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.UpdateAvatar)))
	mux.Handle("DELETE /api/profile/avatar", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.DeleteAvatar)))

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// применение глобальных мидлваров, применяются снизу вверх
	handler := corsMW.Handler(mux)
	handler = loggingMW.Handler(handler)
	handler = requestIDMW.Handler(handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler, // передаем обернутый роутер
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	appLogger.Info("starting server", map[string]any{
		"port":          port,
		"read_timeout":  "10s",
		"write_timeout": "10s",
	})

	err = server.ListenAndServe()
	if err != nil {
		appLogger.Fatal("server failed to start", err)
	}
}
