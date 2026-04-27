package websocket

import (
	"context"
	"encoding/json"
	"sync"
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
	orderConns sync.Map
	cartRooms  sync.Map
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

func (m *WsManager) AddOrderConnection(orderID string, conn *websocket.Conn) {
	if oldConn, ok := m.orderConns.LoadAndDelete(orderID); ok {
		err := oldConn.(*websocket.Conn).Close()
		if err != nil {
			m.logger.Warn("error while closing old order connection", logger.Err(err))
		}
	}

	m.orderConns.Store(orderID, conn)
	m.logger.Info("Order WebSocket connected", logger.String("order_id", orderID))

	go m.readOrderPump(orderID, conn)
}

func (m *WsManager) RemoveOrderConnection(orderID string) {
	if conn, ok := m.orderConns.LoadAndDelete(orderID); ok {
		err := conn.(*websocket.Conn).Close()
		if err != nil {
			m.logger.Warn("error while closing old order connection", logger.Err(err))
		}
		m.logger.Info("Order WebSocket disconnected", logger.String("order_id", orderID))
	}
}

func (m *WsManager) readOrderPump(orderID string, conn *websocket.Conn) {
	defer m.RemoveOrderConnection(orderID)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

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
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

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
		if val, ok := m.orderConns.Load(event.OrderID); ok {
			conn := val.(*websocket.Conn)

			err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				m.logger.Error("failed to set write deadline for order socket", err)
				return err
			}

			err = conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				m.logger.Warn("Failed to write to order websocket", logger.Err(err))
				m.RemoveOrderConnection(event.OrderID)
				return err
			}

			if event.Status == "finished" || event.Status == "cancelled" || event.Status == "failed" {
				m.RemoveOrderConnection(event.OrderID)
			}
		}
	}

	return nil
}
