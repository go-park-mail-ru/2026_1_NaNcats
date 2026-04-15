package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres"
	addressPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/address"
	cartPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/cart"
	orderPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/order"
	paymentPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/payment"
	restaurantPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/restaurant"
	userPG "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres/user"
	paymentCache "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/redisrepo/payment"
	sessionCache "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/redisrepo/session"

	addressUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/address"
	authUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth"
	cartUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/cart"
	orderUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/order"
	paymentUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/payment"
	restaurantUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/restaurant"
	userUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user"

	addressHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/address"
	authHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/auth"
	cartHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/cart"
	orderHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/order"
	paymentHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/payment"
	restaurantHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/restaurant"
	userHandler "github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/handler/user"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/s3"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"github.com/go-playground/validator/v10"
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
	validate := validator.New()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisAddr := flag.String("addr", "redis://user:@localhost:6379/0", "redis addr")
		flag.Parse()
		redisURL = *redisAddr
	}

	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL)
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
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	bucketName := "nancats-bucket"

	// ЮКасса
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	yookassaSecretKey := os.Getenv("YOOKASSA_SECRET_KEY")
	returnURL := os.Getenv("YOOKASSA_RETURN_URL")
	yookassaClient := yookassa.NewClient(shopID, yookassaSecretKey)

	userRepo := userPG.NewUserRepo(pool)
	clientProfileRepo := userPG.NewClientProfileRepo(pool)
	sessionRepo := sessionCache.NewSessionRepo(redisPool)
	restaurantBrandRepo := restaurantPG.NewRestaurantBrandRepo(pool)
	paymentRepo := paymentPG.NewPaymentRepo(pool)
	paymentCacheRepo := paymentCache.NewPaymentCacheRepo(redisPool)
	addressRepo := addressPG.NewAddressRepo(pool)
	orderRepo := orderPG.NewOrderRepo(pool)
	dishRepo := restaurantPG.NewDishRepo(pool)
	cartRepo := cartPG.NewCartRepo(pool)
	s3Repo, err := s3.NewS3Storage(ctx, keyID, s3SecretKey, bucketName, "ru-central1")
	if err != nil {
		appLogger.Fatal("Failed to init S3", err)
	}

	// ttl сессии - 24 часа
	sessionTTL := 24 * time.Hour

	defaultAvatarURL := os.Getenv("DEFAULT_AVATAR_URL")
	defaultRestaurantLogoURL := os.Getenv("DEFAULT_RESTAURANT_LOGO_URL")
	defaultFoodLogoURL := os.Getenv("DEFAULT_FOOD_LOGO_URL")
	if defaultAvatarURL == "" || defaultRestaurantLogoURL == "" || defaultFoodLogoURL == "" {
		appLogger.Fatal("One of the default avatars is null", domain.ErrNoDefaultPhotoURL)
	}

	userUC := userUsecase.NewUserUseCase(userRepo, s3Repo, defaultAvatarURL)
	clientProfileUC := userUsecase.NewClientProfileUseCase(clientProfileRepo)
	sessionUC := authUsecase.NewSessionUseCase(sessionRepo, sessionTTL)
	authUC := authUsecase.NewAuthUseCase(userUC, sessionUC, clientProfileUC)
	restaurantBrandUC := restaurantUsecase.NewRestaurantBrandUseCase(restaurantBrandRepo, defaultRestaurantLogoURL)
	userProfileUC := userUsecase.NewUserProfileUseCase(userUC)
	cartUC := cartUsecase.NewCartUseCase(cartRepo, dishRepo, defaultFoodLogoURL)
	dishUC := restaurantUsecase.NewDishUseCase(dishRepo, defaultFoodLogoURL)
	orderUC := orderUsecase.NewOrderUseCase(orderRepo, addressRepo, cartUC, yookassaClient, defaultRestaurantLogoURL)
	paymentUC := paymentUsecase.NewPaymentUseCase(paymentRepo, paymentCacheRepo, orderRepo, yookassaClient, returnURL)
	addressUC := addressUsecase.NewAddressUseCase(addressRepo)

	if os.Getenv("DEFAULT_AVATAR_URL") == "" {
		appLogger.Warn("DEFAULT_AVATAR_URL пустой, фронтенд может упасть при запросе стандартного аватара")
	}

	authHandler := authHandler.NewAuthHandler(authUC, userUC, appLogger, validate)
	restaurantBrandHandler := restaurantHandler.NewRestaurantBrandHandler(restaurantBrandUC, appLogger)
	userProfileHandler := userHandler.NewUserProfileHandler(userProfileUC, userUC, sessionUC, appLogger)
	paymentHandler := paymentHandler.NewPaymentHandler(paymentUC, appLogger)
	addressHandler := addressHandler.NewAddressHandler(addressUC, appLogger)
	orderHandler := orderHandler.NewOrderHandler(orderUC, appLogger)
	dishHandler := restaurantHandler.NewDishHandler(dishUC, appLogger)
	cartHandler := cartHandler.NewCartHandler(cartUC, appLogger)

	authMW := middleware.NewAuthMiddleware(sessionUC, appLogger)
	csrfMid := middleware.NewCSRFMiddleware(sessionUC, appLogger)
	corsMW := middleware.NewCORSMiddleware([]string{
		"http://localhost:2033",
		"https://localhost:2033",
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
	mux.HandleFunc("GET /api/restaurants/brands/{id}/dishes", dishHandler.GetDishesByRestaurantBrandID)
	mux.HandleFunc("GET /api/restaurants/brands/{id}", restaurantBrandHandler.GetRestaurantBrandByID)

	mux.Handle("GET /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.GetUserProfile)))
	mux.Handle("PATCH /api/profile", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(userProfileHandler.UpdateProfile))))
	mux.Handle("POST /api/profile/avatar", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(userProfileHandler.UpdateAvatar))))
	mux.Handle("DELETE /api/profile/avatar", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(userProfileHandler.DeleteAvatar))))

	mux.Handle("POST /api/profile/cards/bind", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(paymentHandler.InitiateCardBinding))))
	mux.Handle("GET /api/profile/cards", authMW.RequireAuth(http.HandlerFunc(paymentHandler.GetUserCards)))
	mux.Handle("DELETE /api/profile/cards/{id}", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(paymentHandler.DeleteCard))))
	mux.Handle("PUT /api/profile/cards/{id}", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(paymentHandler.SetDefaultCard))))

	mux.Handle("POST /api/profile/addresses", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(addressHandler.AddAddress))))
	mux.Handle("GET /api/profile/addresses", authMW.RequireAuth(http.HandlerFunc(addressHandler.GetAddresses)))
	mux.Handle("DELETE /api/profile/addresses/{id}", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(addressHandler.DeleteAddress))))
	mux.Handle("PATCH /api/profile/addresses/{id}", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(addressHandler.UpdateAddress))))

	mux.Handle("POST /api/orders", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(orderHandler.CreateOrder))))
	mux.Handle("GET /api/profile/orders", authMW.RequireAuth(http.HandlerFunc(orderHandler.GetMyOrders)))

	mux.Handle("POST /api/webhooks/yookassa", http.HandlerFunc(paymentHandler.YookassaWebhook))

	mux.Handle("GET /api/cart", authMW.RequireAuth(http.HandlerFunc(cartHandler.GetCart)))
	mux.Handle("PUT /api/cart", authMW.RequireAuth(csrfMid.Check(http.HandlerFunc(cartHandler.UpdateCart))))

	mux.Handle("GET /api/csrf", authMW.RequireAuth(http.HandlerFunc(authHandler.GetCSRF)))

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

	appLogger.Info("starting server",
		domain.String("port", port),
		domain.String("read_timeout", "10s"),
		domain.String("write_timeout", "10s"),
	)

	err = server.ListenAndServe()
	if err != nil {
		appLogger.Fatal("server failed to start", err)
	}
}
