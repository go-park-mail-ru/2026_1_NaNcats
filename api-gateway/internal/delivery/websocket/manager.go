package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/mailru/easyjson"
)

// Оборачивает оригинальный websocket.Conn
type SafeConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewSafeConnection(conn *websocket.Conn) *SafeConnection {
	return &SafeConnection{
		conn: conn,
	}
}

// Безопасно записывает сообщение в сокет, захватывая мьютекс
func (s *SafeConnection) WriteMessage(messageType int, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteMessage(messageType, data)
}

// Безопасно отправляет управляющий фрейм в сокет
func (s *SafeConnection) WriteControl(controlType int, data []byte, deadline time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteControl(controlType, data, deadline)
}

// Безопасно устанавливает таймаут записи
func (s *SafeConnection) SetWriteDeadline(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.SetWriteDeadline(t)
}

// Закрывает сетевое соединение сокета, не используем мьютекс,
// Close() - потокобезопасный
func (s *SafeConnection) Close() error {
	return s.conn.Close()
}

type Room struct {
	mu      sync.RWMutex
	clients map[int64]*SafeConnection
}

type WsManager struct {
	orderRooms sync.Map
	cartRooms  sync.Map
	connSeq    atomic.Int64
	redisPool  *redis.Pool
	channel    string
	logger     logger.Logger
}

func NewWsManager(rp *redis.Pool, channel string, l logger.Logger) *WsManager {
	return &WsManager{
		redisPool: rp,
		channel:   channel,
		logger:    l,
	}
}

// AddOrderConnection подключает зрителя к комнате заказа. Комната допускает
// несколько соединений одновременно: организатор и участники совместного
// заказа смотрят его параллельно, поэтому новое подключение не вытесняет уже открытые
func (m *WsManager) AddOrderConnection(orderID string, conn *websocket.Conn) {
	val, _ := m.orderRooms.LoadOrStore(orderID, &Room{clients: make(map[int64]*SafeConnection)})
	room := val.(*Room)

	connID := m.connSeq.Add(1)

	safeConn := NewSafeConnection(conn)

	room.mu.Lock()
	room.clients[connID] = safeConn
	room.mu.Unlock()

	// Новому зрителю сразу отдаём последнее событие заказа из кэша.
	rc := m.redisPool.Get()
	cachedMsg, err := redis.Bytes(rc.Do("GET", "ws_cache:order:"+orderID))
	rc.Close()
	if err == nil && len(cachedMsg) > 0 {
		_ = conn.WriteMessage(websocket.TextMessage, cachedMsg)
	}

	m.logger.Info("Order WebSocket connected", logger.String("order_id", orderID))

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := safeConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

	// Для чтения используем сырое соединение, так как чтение не конкурирует с записью
	go m.readOrderPump(orderID, connID, conn)
}

func (m *WsManager) RemoveOrderConnection(orderID string, connID int64) {
	val, ok := m.orderRooms.Load(orderID)
	if !ok {
		return
	}
	room := val.(*Room)

	room.mu.Lock()
	if conn, exists := room.clients[connID]; exists {
		_ = conn.Close()
		delete(room.clients, connID)
		m.logger.Info("Order WebSocket disconnected", logger.String("order_id", orderID))
	}
	isEmpty := len(room.clients) == 0
	room.mu.Unlock()

	if isEmpty {
		m.orderRooms.Delete(orderID)
	}
}

// Закрывает все соединения комнаты заказа и удаляет её, нужен
// при терминальном статусе заказа, когда обновлять больше нечего
func (m *WsManager) closeOrderRoom(orderID string) {
	val, ok := m.orderRooms.LoadAndDelete(orderID)
	if !ok {
		return
	}
	room := val.(*Room)

	room.mu.Lock()
	for _, conn := range room.clients {
		_ = conn.Close()
	}
	room.clients = make(map[int64]*SafeConnection)
	room.mu.Unlock()
}

func (m *WsManager) readOrderPump(orderID string, connID int64, conn *websocket.Conn) {
	defer m.RemoveOrderConnection(orderID, connID)
	conn.SetReadLimit(512)

	conn.SetPongHandler(func(string) error {
		err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			m.logger.Error("failed to set read deadline for order", err, logger.Err(err))
			return err
		}
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (m *WsManager) AddCartConnection(cartID string, userID int64, conn *websocket.Conn) {
	// Достаем существующую комнату или создаем новую
	val, _ := m.cartRooms.LoadOrStore(cartID, &Room{
		clients: make(map[int64]*SafeConnection),
	})
	room := val.(*Room)

	// Оборачиваем сырое соединение в потокобезопасную структуру
	safeConn := NewSafeConnection(conn)

	room.mu.Lock()
	if oldConn, exists := room.clients[userID]; exists {
		_ = oldConn.Close()
	}
	room.clients[userID] = safeConn
	room.mu.Unlock()

	m.logger.Info("Joined cart room", logger.String("cart_id", cartID), logger.Int("user_id", int(userID)))

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := safeConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

	// Для чтения используем сырое соединение, т.к. чтение/запись в gorilla/websocket могут происходить параллельно
	go m.readCartPump(cartID, userID, conn)
}

func (m *WsManager) RemoveCartConnection(cartID string, userID int64) {
	val, ok := m.cartRooms.Load(cartID)
	if !ok {
		return
	}
	room := val.(*Room)

	room.mu.Lock()
	if conn, exists := room.clients[userID]; exists {
		_ = conn.Close()
		delete(room.clients, userID)
		m.logger.Info("Left cart room", logger.String("cart_id", cartID), logger.Int("user_id", int(userID)))
	}

	// Проверяем, не опустела ли комната
	isEmpty := len(room.clients) == 0
	room.mu.Unlock()

	// Если в комнате никого не осталось, чистим
	if isEmpty {
		m.cartRooms.Delete(cartID)
		m.logger.Info("Cart room deleted (empty)", logger.String("cart_id", cartID))
	}
}

func (m *WsManager) readCartPump(cartID string, userID int64, conn *websocket.Conn) {
	defer m.RemoveCartConnection(cartID, userID)
	conn.SetReadLimit(512)

	conn.SetPongHandler(func(string) error {
		err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			m.logger.Error("failed to set read deadline for cart", err, logger.Err(err))
			return err
		}
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (m *WsManager) BroadcastToRedis(event events.GatewayEvent) error {
	conn := m.redisPool.Get()
	defer conn.Close()

	data, err := easyjson.Marshal(event)
	if err != nil {
		m.logger.Error("easyjson failed", err, logger.Err(err))
		return err
	}

	if event.OrderID != "" {
		_, _ = conn.Do("SETEX", "ws_cache:order:"+event.OrderID, 300, data)
	}

	_, err = conn.Do("PUBLISH", m.channel, data)
	return err
}

func (m *WsManager) RunPubSubListener(ctx context.Context) {
	conn := m.redisPool.Get()
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}

	if err := psc.Subscribe(m.channel); err != nil {
		m.logger.Error("Failed to subscribe to Redis Pub/Sub", err)
		return
	}

	go func() {
		<-ctx.Done()
		_ = psc.Unsubscribe()
		_ = conn.Close()
	}()

	m.logger.Info("Started listening to Redis Pub/Sub", logger.String("channel", m.channel))

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			var event events.GatewayEvent
			if err := json.Unmarshal(v.Data, &event); err != nil {
				m.logger.Error("Failed to unmarshal pubsub msg", err)
				continue
			}
			_ = m.deliverToLocalSocket(event)
		case error:
			m.logger.Error("Redis PubSub error", v)
			time.Sleep(time.Second)
		}
	}
}

func (m *WsManager) deliverToLocalSocket(event events.GatewayEvent) error {
	msgBytes, err := easyjson.Marshal(event)
	if err != nil {
		m.logger.Error("easyjson failed", err, logger.Err(err))
		return err
	}

	if event.CartID != "" {
		if val, ok := m.cartRooms.Load(event.CartID); ok {
			room := val.(*Room)

			room.mu.RLock()
			activeConns := make(map[int64]*SafeConnection, len(room.clients))
			for userID, conn := range room.clients {
				activeConns[userID] = conn
			}
			room.mu.RUnlock()

			for userID, conn := range activeConns {
				err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					m.logger.Warn("Failed to set write deadline for cart socket", logger.Int("user_id", int(userID)))
					continue
				}

				err = conn.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					m.logger.Warn("Failed to write to cart websocket", logger.Int("user_id", int(userID)), logger.Err(err))
				}
			}

			if event.EventType == "SharedCartClosed" {
				m.cartRooms.Delete(event.CartID)
				m.logger.Info("Cart room force deleted", logger.String("cart_id", event.CartID))
			}
		}
	}

	if event.OrderID != "" {
		if val, ok := m.orderRooms.Load(event.OrderID); ok {
			room := val.(*Room)

			room.mu.RLock()
			for connID, conn := range room.clients {
				if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
					m.logger.Warn("Failed to set write deadline for order socket", logger.Int("conn_id", int(connID)))
					continue
				}
				if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					m.logger.Warn("Failed to write to order websocket", logger.Err(err))
				}
			}
			room.mu.RUnlock()

			if event.Status == "finished" || event.Status == "cancelled" || event.Status == "failed" {
				m.closeOrderRoom(event.OrderID)
			}
		}
	}

	return nil
}
