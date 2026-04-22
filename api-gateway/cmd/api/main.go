package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/auth"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/user"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"

	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	_ "github.com/go-park-mail-ru/2026_1_NaNcats/docs"
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

	// Читаем уровень логирования из переменной окружения
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	rawLogger, err := logger.NewZapLogger(logLevel)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting API Gateway...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	validate := validator.New()

	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50054"
	}
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50052"
	}

	appLogger.Info("Connecting to Auth Service...")
	authConn, err := grpc.NewClient(
		authServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptors.UnaryClientRequestID(), // Прокидываем TraceID
			interceptors.UnaryClientUserID(),    // Прокидываем UserID
		),
	)
	if err != nil {
		appLogger.Fatal("Failed to connect to Auth Service", err)
	}
	defer authConn.Close()

	appLogger.Info("Connecting to User Service...")
	userConn, err := grpc.NewClient(
		userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptors.UnaryClientRequestID(),
			interceptors.UnaryClientUserID(),
		),
	)
	if err != nil {
		appLogger.Fatal("Failed to connect to User Service", err)
	}
	defer userConn.Close()

	authGrpcClient := pbAuth.NewAuthServiceClient(authConn)
	authClient := authclient.NewAuthClient(authGrpcClient)

	userGrpcClient := pbUser.NewUserServiceClient(userConn)
	userClient := userclient.NewUserClient(userGrpcClient)

	authHandler := auth.NewAuthHandler(authClient, userClient, appLogger, validate)
	userProfileHandler := user.NewUserProfileHandler(userClient, appLogger)

	reqIDMW := middleware.NewRequestIDMiddleware()
	loggingMW := middleware.NewLoggingMiddleware(appLogger)
	corsMW := middleware.NewCORSMiddleware([]string{
		"http://localhost:3000",
		"http://localhost:2033",
		"https://localhost:2033",
	})
	authMW := middleware.NewAuthMiddleware(authClient, appLogger)
	csrfMW := middleware.NewCSRFMiddleware(authClient, appLogger)

	// создание собственного роутера
	mux := http.NewServeMux()

	// === AUTH ===
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/auth/me", authMW.RequireAuth(http.HandlerFunc(authHandler.GetMe)))
	mux.Handle("GET /api/csrf", authMW.RequireAuth(http.HandlerFunc(authHandler.GetCSRF)))

	// === PROFILE ===
	mux.Handle("GET /api/profile", authMW.RequireAuth(http.HandlerFunc(userProfileHandler.GetUserProfile)))
	mux.Handle("PATCH /api/profile", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(userProfileHandler.UpdateProfile))))
	mux.Handle("POST /api/profile/avatar", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(userProfileHandler.UpdateAvatar))))
	mux.Handle("DELETE /api/profile/avatar", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(userProfileHandler.DeleteAvatar))))

	// TODO: потом прикрутить остальные микросервисы
	/*
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
	*/

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// применение глобальных мидлваров, применяются снизу вверх
	handler := corsMW.Handler(mux)
	handler = loggingMW.Handler(handler)
	handler = reqIDMW.Handler(handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler, // передаем обернутый роутер
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		appLogger.Info("API Gateway is running", logger.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server failed to start", err)
		}
	}()

	// 10. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping API Gateway gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", err)
	}
	appLogger.Info("API Gateway stopped")
}
