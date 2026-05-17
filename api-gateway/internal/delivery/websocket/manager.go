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

type Room struct {
	mu      sync.RWMutex
	clients map[int64]*websocket.Conn
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
// заказа смотрят его параллельно, поэтому новое подключение не вытесняет уже
// открытые.
func (m *WsManager) AddOrderConnection(orderID string, conn *websocket.Conn) {
	val, _ := m.orderRooms.LoadOrStore(orderID, &Room{clients: make(map[int64]*websocket.Conn)})
	room := val.(*Room)

	connID := m.connSeq.Add(1)

	room.mu.Lock()
	room.clients[connID] = conn
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
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

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

// closeOrderRoom закрывает все соединения комнаты заказа и удаляет её. Нужен
// при терминальном статусе заказа, когда обновлять больше нечего.
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
	room.clients = make(map[int64]*websocket.Conn)
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
		clients: make(map[int64]*websocket.Conn),
	})
	room := val.(*Room)

	room.mu.Lock()
	if oldConn, exists := room.clients[userID]; exists {
		_ = oldConn.Close()
	}
	room.clients[userID] = conn
	room.mu.Unlock()

	m.logger.Info("Joined cart room", logger.String("cart_id", cartID), logger.Int("user_id", int(userID)))

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}()

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
		psc.Unsubscribe()
		conn.Close()
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
			for userID, conn := range room.clients {
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
			room.mu.RUnlock()

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
