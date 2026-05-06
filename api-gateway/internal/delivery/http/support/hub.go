package support

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

//easyjson:json
type WsEvent struct {
	ID             int64  `json:"id"`
	TicketPublicID string `json:"ticket_public_id"`
	TicketID       int64  `json:"ticket_id"`
	AuthorRole     string `json:"author_role"`
	Text           string `json:"text"`
	CreatedAt      string `json:"created_at"`
}

type RedisHub struct {
	pool        *redis.Pool
	logger      logger.Logger
	connections map[string][]*websocket.Conn
	mu          sync.RWMutex
}

func NewRedisHub(pool *redis.Pool, l logger.Logger) *RedisHub {
	return &RedisHub{
		pool:        pool,
		logger:      l,
		connections: make(map[string][]*websocket.Conn),
	}
}

func (h *RedisHub) AddConnection(ctx context.Context, ticketPublicID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[ticketPublicID] = append(h.connections[ticketPublicID], conn)

	if len(h.connections[ticketPublicID]) == 1 {
		go h.subscribeToRedis(ctx, ticketPublicID)
	}
}

func (h *RedisHub) RemoveConnection(ticketPublicID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[ticketPublicID]
	for i, c := range conns {
		if c == conn {
			h.connections[ticketPublicID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *RedisHub) Publish(ticketPublicID string, event WsEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	channel := h.getChannelName(ticketPublicID)

	conn := h.pool.Get()
	defer conn.Close()

	_, err = conn.Do("PUBLISH", channel, payload)
	return err
}

func (h *RedisHub) subscribeToRedis(ctx context.Context, ticketPublicID string) {
	channel := h.getChannelName(ticketPublicID)

	conn := h.pool.Get()
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}
	if err := psc.Subscribe(channel); err != nil {
		h.logger.Error("failed to subscribe to redis", err)
		return
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
			return
		}
	}()

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			h.mu.RLock()
			localConns := h.connections[ticketPublicID]
			h.mu.RUnlock()

			if len(localConns) == 0 {
				psc.Unsubscribe(channel)
				return
			}

			for _, wsConn := range localConns {
				err := wsConn.WriteMessage(websocket.TextMessage, v.Data)
				if err != nil {
					h.logger.Warn("failed to write to local ws conn", logger.Err(err))
				}
			}

		case redis.Subscription:
			if v.Count == 0 {
				return
			}

		case error:
			h.logger.Info("redis pubsub connection closed or error", logger.String("error", v.Error()))
			return
		}
	}
}

func (h *RedisHub) getChannelName(ticketPublicID string) string {
	return "ticket_chat_" + ticketPublicID
}
