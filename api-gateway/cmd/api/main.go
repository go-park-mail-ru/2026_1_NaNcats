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
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	addressHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/address"
	authHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/auth"
	cartHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/cart"
	orderHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/order"
	paymentHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/payment"
	restaurantHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/restaurant"
	supportHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/support"
	userHttp "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/http/user"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	rabbitclient "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/addressclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"

	gatewayConfig "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"

	pbAddress "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	pbCart "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pbPayment "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	pbSupport "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	_ "github.com/go-park-mail-ru/2026_1_NaNcats/docs"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func mustInitConn(addr string, serviceName string, appLogger logger.Logger, opts []grpc.DialOption) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		appLogger.Fatal("Failed to connect to "+serviceName, err)
	}
	appLogger.Info("Connected to "+serviceName, logger.String("addr", addr))
	return conn
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
	cfg := gatewayConfig.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting API Gateway...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cleanup, err := metrics.InitMetrics(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("Failed to init metrics", err)
	}
	defer cleanup()

	cleanupTraces, err := metrics.InitTracing(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("Failed to init tracing", err)
	}
	defer cleanupTraces()

	validate := validator.New()

	redisPool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// TODO: вынести адрес в конфиг, пока так сойдет
			redisAddr := os.Getenv("REDIS_ADDR")
			if redisAddr == "" {
				redisAddr = "localhost:6379"
			}
			return redis.Dial("tcp", redisAddr)
		},
	}
	defer redisPool.Close()

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			interceptors.UnaryClientRequestID(),
			interceptors.UnaryClientUserID(),
			interceptors.UnaryClientLogging(appLogger),
		),
	}
	appLogger.Info("Connecting to microservices...")

	authConn := mustInitConn(cfg.GRPCClients.AuthAddr, "Auth Service", appLogger, grpcOpts)
	defer authConn.Close()

	userConn := mustInitConn(cfg.GRPCClients.UserAddr, "User Service", appLogger, grpcOpts)
	defer userConn.Close()

	restConn := mustInitConn(cfg.GRPCClients.RestaurantAddr, "Restaurant Service", appLogger, grpcOpts)
	defer restConn.Close()

	cartConn := mustInitConn(cfg.GRPCClients.CartAddr, "Cart Service", appLogger, grpcOpts)
	defer cartConn.Close()

	addrConn := mustInitConn(cfg.GRPCClients.AddressAddr, "Address Service", appLogger, grpcOpts)
	defer addrConn.Close()

	payConn := mustInitConn(cfg.GRPCClients.PaymentAddr, "Payment Service", appLogger, grpcOpts)
	defer payConn.Close()

	orderConn := mustInitConn(cfg.GRPCClients.OrderAddr, "Order Service", appLogger, grpcOpts)
	defer orderConn.Close()

	supportConn := mustInitConn(cfg.GRPCClients.SupportAddr, "Support Service", appLogger, grpcOpts)
	defer supportConn.Close()

	authClient := authclient.NewAuthClient(pbAuth.NewAuthServiceClient(authConn))
	userClient := userclient.NewUserClient(pbUser.NewUserServiceClient(userConn))
	restClient := restaurantclient.NewRestaurantClient(pbRestaurant.NewRestaurantServiceClient(restConn))
	cartClient := cartclient.NewCartClient(pbCart.NewCartServiceClient(cartConn))
	addrClient := addressclient.NewAddressClient(pbAddress.NewAddressServiceClient(addrConn))
	payClient := paymentclient.NewPaymentClient(pbPayment.NewPaymentServiceClient(payConn))
	orderClient := orderclient.NewOrderClient(pbOrder.NewOrderServiceClient(orderConn))
	supportClient := supportclient.NewSupportClient(pbSupport.NewSupportServiceClient(supportConn))
	rabbitClient, err := rabbitclient.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("failed to init RabbitMq client", err)
	}

	orderChannel := "gateway:order:events"
	wsManager := websocket.NewWsManager(redisPool, orderChannel, appLogger)

	go wsManager.RunPubSubListener(ctx)

	authHandler := authHttp.NewAuthHandler(authClient, userClient, appLogger, validate)
	userProfileHandler := userHttp.NewUserProfileHandler(userClient, appLogger)
	restaurantHandler := restaurantHttp.NewRestaurantHandler(restClient, appLogger)
	cartHandler := cartHttp.NewCartHandler(cartClient, appLogger)
	addressHandler := addressHttp.NewAddressHandler(addrClient, appLogger)
	paymentHandler := paymentHttp.NewPaymentHandler(payClient, appLogger)
	orderHandler := orderHttp.NewOrderHandler(orderClient, wsManager, appLogger)

	redisHub := supportHttp.NewRedisHub(redisPool, appLogger)
	supportHandler := supportHttp.NewSupportHandler(supportClient, redisHub, appLogger)

	reqIDMW := middleware.NewRequestIDMiddleware()
	loggingMW := middleware.NewLoggingMiddleware(appLogger)
	corsMW := middleware.NewCORSMiddleware(cfg.HTTP.AllowedOrigins)
	authMW := middleware.NewAuthMiddleware(authClient, appLogger)
	csrfMW := middleware.NewCSRFMiddleware(authClient, appLogger)

	gatewayConsumer := rabbitmq.NewGatewayConsumer(rabbitClient, wsManager, appLogger)
	if err := gatewayConsumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start Gateway RabbitMQ consumer", err)
	}

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
	// Базовые операции
	mux.Handle("GET /api/cart", authMW.RequireAuth(http.HandlerFunc(cartHandler.GetCart)))
	mux.Handle("DELETE /api/cart", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.ClearCart))))
	mux.Handle("POST /api/cart/lock", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.LockCart))))

	// Операции с товарами
	mux.Handle("POST /api/cart/items", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.AddItem))))
	mux.Handle("PUT /api/cart/items", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.UpdateQuantity))))
	mux.Handle("DELETE /api/cart/items", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.RemoveItem))))
	mux.Handle("PATCH /api/cart/items/owner", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.ReassignOwner))))

	// Управление совместной корзиной
	mux.Handle("POST /api/cart/invite", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.GenerateInvite))))
	mux.Handle("POST /api/cart/join", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.JoinCart))))
	mux.Handle("DELETE /api/cart/members", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.KickMember))))
	mux.Handle("POST /api/cart/close", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(cartHandler.CloseSharedCart))))

	// WebSockets
	mux.Handle("GET /api/ws/cart", authMW.RequireAuth(http.HandlerFunc(cartHandler.ConnectCartWS)))

	// === ORDERS ===
	mux.Handle("POST /api/orders", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(orderHandler.CreateOrder))))
	mux.Handle("GET /api/profile/orders", authMW.RequireAuth(http.HandlerFunc(orderHandler.GetMyOrders)))
	mux.Handle("GET /api/ws/orders/{id}", authMW.RequireAuth(http.HandlerFunc(orderHandler.TrackOrderWS)))
	mux.Handle("POST /api/orders/splits/{id}/pay", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(orderHandler.PayForFriend))))

	// === SUPPORT === (Пользовательская часть)
	mux.HandleFunc("GET /api/support/categories", supportHandler.GetCategories)
	mux.HandleFunc("POST /api/support/tickets", supportHandler.CreateTicket)
	mux.HandleFunc("GET /api/support/tickets", supportHandler.GetMyTickets)
	mux.HandleFunc("GET /api/support/tickets/{id}/events", supportHandler.GetTicketEvents)
	mux.HandleFunc("POST /api/support/tickets/{id}/rate", supportHandler.RateTicket)
	mux.HandleFunc("GET /api/support/tickets/{id}/chat", supportHandler.ConnectChat)

	// === SUPPORT === (Операторская часть / Админка)
	mux.Handle("GET /api/admin/support/tickets", authMW.RequireAuth(http.HandlerFunc(supportHandler.GetAssignedTickets)))
	mux.Handle("PATCH /api/admin/support/tickets/{id}/status", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(supportHandler.ChangeTicketStatus))))
	mux.Handle("POST /api/admin/support/tickets/{id}/reassign", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(supportHandler.ReassignTicket))))
	mux.Handle("PATCH /api/admin/support/agent/status", authMW.RequireAuth(csrfMW.Check(http.HandlerFunc(supportHandler.SetAgentStatus))))
	mux.Handle("GET /api/admin/support/templates", authMW.RequireAuth(http.HandlerFunc(supportHandler.GetTemplates)))

	// === SWAGGER ===
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	handler := corsMW.Handler(mux)
	handler = loggingMW.Handler(handler)
	handler = reqIDMW.Handler(handler)
	otelHandler := otelhttp.NewHandler(handler, "api-gateway")

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      otelHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		appLogger.Info("API Gateway is running", logger.String("port", cfg.HTTP.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server failed to start", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping API Gateway gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatal("Server forced to shutdown", err)
	}
	appLogger.Info("API Gateway stopped")
}
