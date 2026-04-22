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

	addressHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/address"
	authHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/auth"
	cartHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/cart"
	orderHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/order"
	paymentHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/payment"
	restaurantHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/restaurant"
	userHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/user"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/addressclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"

	pbAddress "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	pbCart "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pbPayment "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	_ "github.com/go-park-mail-ru/2026_1_NaNcats/docs"
)

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

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

	services := map[string]string{
		"auth":       getEnvOrDefault("AUTH_SERVICE_ADDR", "localhost:50054"),
		"user":       getEnvOrDefault("USER_SERVICE_ADDR", "localhost:50052"),
		"restaurant": getEnvOrDefault("RESTAURANT_SERVICE_ADDR", "localhost:50053"),
		"cart":       getEnvOrDefault("CART_SERVICE_ADDR", "localhost:50055"),
		"address":    getEnvOrDefault("ADDRESS_SERVICE_ADDR", "localhost:50051"),
		"payment":    getEnvOrDefault("PAYMENT_SERVICE_ADDR", "localhost:50056"),
		"order":      getEnvOrDefault("ORDER_SERVICE_ADDR", "localhost:50057"),
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptors.UnaryClientRequestID(),
			interceptors.UnaryClientUserID(),
			interceptors.UnaryClientLogging(appLogger),
		),
	}
	appLogger.Info("Connecting to microservices...")

	authConn, err := grpc.NewClient(services["auth"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Auth Service", err)
	}
	defer authConn.Close()

	userConn, err := grpc.NewClient(services["user"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to User Service", err)
	}
	defer userConn.Close()

	restConn, err := grpc.NewClient(services["restaurant"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Restaurant Service", err)
	}
	defer restConn.Close()

	cartConn, err := grpc.NewClient(services["cart"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Cart Service", err)
	}
	defer cartConn.Close()

	addrConn, err := grpc.NewClient(services["address"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Address Service", err)
	}
	defer addrConn.Close()

	payConn, err := grpc.NewClient(services["payment"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Payment Service", err)
	}
	defer payConn.Close()

	orderConn, err := grpc.NewClient(services["order"], grpcOpts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to Order Service", err)
	}
	defer orderConn.Close()

	authClient := authclient.NewAuthClient(pbAuth.NewAuthServiceClient(authConn))
	userClient := userclient.NewUserClient(pbUser.NewUserServiceClient(userConn))
	restClient := restaurantclient.NewRestaurantClient(pbRestaurant.NewRestaurantServiceClient(restConn))
	cartClient := cartclient.NewCartClient(pbCart.NewCartServiceClient(cartConn))
	addrClient := addressclient.NewAddressClient(pbAddress.NewAddressServiceClient(addrConn))
	payClient := paymentclient.NewPaymentClient(pbPayment.NewPaymentServiceClient(payConn))
	orderClient := orderclient.NewOrderClient(pbOrder.NewOrderServiceClient(orderConn))

	authHandler := authHttp.NewAuthHandler(authClient, userClient, appLogger, validate)
	userProfileHandler := userHttp.NewUserProfileHandler(userClient, appLogger)
	restaurantHandler := restaurantHttp.NewRestaurantHandler(restClient, appLogger)
	cartHandler := cartHttp.NewCartHandler(cartClient, appLogger)
	addressHandler := addressHttp.NewAddressHandler(addrClient, appLogger)
	paymentHandler := paymentHttp.NewPaymentHandler(payClient, appLogger)
	orderHandler := orderHttp.NewOrderHandler(orderClient, appLogger)

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

	// === RESTAURANTS & DISHES ===
	mux.HandleFunc("GET /api/restaurants/brands", restaurantHandler.GetRestaurantBrandsList)
	mux.HandleFunc("GET /api/restaurants/brands/{id}", restaurantHandler.GetRestaurantBrandByID)
	mux.HandleFunc("GET /api/restaurants/brands/{id}/dishes", restaurantHandler.GetDishesByRestaurantBrandID)

	// === ADDRESSES ===
	mux.Handle("POST /api/profile/addresses", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(addressHandler.AddAddress))))
	mux.Handle("GET /api/profile/addresses", authMW.RequireAuth(http.HandlerFunc(addressHandler.GetAddresses)))
	mux.Handle("DELETE /api/profile/addresses/{id}", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(addressHandler.DeleteAddress))))
	mux.Handle("PATCH /api/profile/addresses/{id}", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(addressHandler.UpdateAddress))))

	// === PAYMENTS ===
	mux.Handle("POST /api/profile/cards/bind", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(paymentHandler.InitiateCardBinding))))
	mux.Handle("GET /api/profile/cards", authMW.RequireAuth(http.HandlerFunc(paymentHandler.GetUserCards)))
	mux.Handle("DELETE /api/profile/cards/{id}", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(paymentHandler.DeleteCard))))
	mux.Handle("PUT /api/profile/cards/{id}", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(paymentHandler.SetDefaultCard))))
	mux.HandleFunc("POST /api/webhooks/yookassa", paymentHandler.YookassaWebhook) // ВАЖНО: без мидлварей авторизации!

	// === CART ===
	mux.Handle("GET /api/cart", authMW.RequireAuth(http.HandlerFunc(cartHandler.GetCart)))
	mux.Handle("PUT /api/cart", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.UpdateCart))))

	// === ORDERS ===
	mux.Handle("POST /api/orders", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(orderHandler.CreateOrder))))
	mux.Handle("GET /api/profile/orders", authMW.RequireAuth(http.HandlerFunc(orderHandler.GetMyOrders)))

	// === SWAGGER ===
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	handler := corsMW.Handler(mux)
	handler = loggingMW.Handler(handler)
	handler = reqIDMW.Handler(handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		appLogger.Info("API Gateway is running", logger.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server failed to start", err)
		}
	}()

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
