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

type WsManager struct {
	localConns sync.Map
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

func (m *WsManager) AddConnection(orderID string, conn *websocket.Conn) {
	if oldConn, ok := m.localConns.Load(conn); ok {
		err := oldConn.(*websocket.Conn).Close()
		if err != nil {
			m.logger.Warn("error while closing old connection", logger.Err(err))
		}
		m.localConns.Store(orderID, conn)
		m.logger.Info("WebSocket connected", logger.String("order_id", orderID))

		go m.readPump(orderID, conn)
	}
}

func (m *WsManager) RemoveConnection(orderID string) {
	if conn, ok := m.localConns.LoadAndDelete(orderID); ok {
		err := conn.(*websocket.Conn).Close()
		if err != nil {
			m.logger.Warn("error while closing old connection", logger.Err(err))
		}
		m.logger.Info("WebSocket disconnected", logger.String("order_id", orderID))
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
			time.Sleep(time.Second) // не спамим жи ес
		}
	}
}

func (m *WsManager) deliverToLocalSocket(event events.GatewayEvent) error {
	val, ok := m.localConns.Load(event.OrderID)
	if !ok {
		// сокета тут нет, это норм
		return nil
	}
	conn := val.(*websocket.Conn)

	msgBytes, err := easyjson.Marshal(event)
	if err != nil {
		m.logger.Error("easyjson failed", err, logger.Err(err))
		return err
	}

	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		m.logger.Error("failed to set write deadline", err, logger.Err(err))
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		m.logger.Warn("Failed to write to websocket", logger.String("error", err.Error()))
		m.RemoveConnection(event.OrderID)
		return err
	}

	if event.Status == "finished" || event.Status == "cancelled" || event.Status == "failed" {
		m.RemoveConnection(event.OrderID)
	}

	return nil
}

func (m *WsManager) readPump(orderID string, conn *websocket.Conn) {
	defer m.RemoveConnection(orderID)
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn.SetPongHandler(func(string) error {
		err := conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if err != nil {
			m.logger.Error("failed to set read deadline", err, logger.Err(err))
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
