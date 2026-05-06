package support

//go:generate easyjson $GOFILE

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/sanitizer"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mailru/easyjson"
	"github.com/microcosm-cc/bluemonday"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//easyjson:json
type WsMessage struct {
	Text           string `json:"text"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (m *WsMessage) Sanitize(p *bluemonday.Policy) {
	m.Text = p.Sanitize(m.Text)
}

// ConnectChat godoc
// @Summary 		Подключение к чату поддержки (WebSocket)
// @Description		Устанавливает WebSocket соединение для общения в чате поддержки по конкретному тикету. Ожидает HTTP GET для апгрейда протокола.
// @Tags			support
// @Param			id		path	string	true					"Public ID тикета"
// @Success			101		"Протокол успешно изменен на WebSocket"
// @Failure			400		"Неверный ID тикета или ошибка апгрейда соединения"
// @Router			/support/tickets/{id}/chat [get]
func (h *SupportHandler) ConnectChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	ticketPublicID := r.PathValue("id")
	if ticketPublicID == "" {
		http.Error(w, "invalid ticket id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Error("failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	h.redisHub.AddConnection(ctx, ticketPublicID, conn)
	defer h.redisHub.RemoveConnection(ticketPublicID, conn)

	var authorID *int64
	authorRole := "USER"

	uID, isAuth := middleware.GetUserID(ctx)
	if isAuth {
		authorID = &uID
	}

	l.Info("client connected to chat", logger.String("ticket_public_id", ticketPublicID))

	for {
		_, messageData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.Error("websocket error", err)
			}
			break
		}

		var wsMsg WsMessage
		if err := easyjson.Unmarshal(messageData, &wsMsg); err != nil {
			l.Warn("invalid ws message format", logger.Err(err))
			continue
		}

		wsMsg.Sanitize(sanitizer.Policy)

		if wsMsg.IdempotencyKey == "" {
			wsMsg.IdempotencyKey = uuid.New().String()
		}

		input := supportclient.SendMessageInput{
			TicketPublicID: ticketPublicID,
			AuthorID:       authorID,
			AuthorRole:     authorRole,
			Message:        wsMsg.Text,
		}

		msgID, err := h.supportClient.SendMessage(ctx, input, wsMsg.IdempotencyKey)
		if err != nil {
			l.Error("failed to save message via grpc", err)
			conn.WriteJSON(map[string]interface{}{"error": "failed to send message"})
			continue
		}

		event := WsEvent{
			ID:             msgID,
			TicketPublicID: ticketPublicID,
			AuthorRole:     authorRole,
			Text:           wsMsg.Text,
			CreatedAt:      time.Now().Format(time.RFC3339),
		}

		if err := h.redisHub.Publish(ticketPublicID, event); err != nil {
			l.Error("failed to publish to redis", err)
		}
	}
}
