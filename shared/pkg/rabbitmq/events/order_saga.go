package events

//go:generate easyjson $GOFILE

// Очереди, в которые будет идти запись
const (
	QueueCartCommands    = "saga.cart.commands"
	QueuePaymentCommands = "saga.payment.commands"
	QueueOrderReplies    = "saga.order.replies"  // сюда воркеры ответы пишут
	QueueGatewayEvents   = "saga.gateway.events" // order отсылает сюда инфу для WebSocket
)

// команды оркестратора
const (
	CommandLockCart      = "LOCK_CART"
	CommandUnlockCart    = "UNLOCK_CART"
	CommandClearCart     = "CLEAR_CART"
	CommandCreatePayment = "CREATE_PAYMENT"
)

const (
	StatusSuccess = "SUCCESS"
	StatusError   = "ERROR"
)

// Команды от OrderService
//
//easyjson:json
type SagaCommand struct {
	OrderID        string `json:"order_id"` // public_id
	UserID         int64  `json:"user_id"`
	CartID         string `json:"cart_id"`
	Action         string `json:"action"` // команда для выполнения
	IdempotencyKey string `json:"idempotency_key"`

	Amount          int64  `json:"amount,omitempty"`
	PaymentMethodID string `json:"payment_method_id,omitempty"`
}

// Ответы оркестратору
//
//easyjson:json
type SagaReply struct {
	OrderID      string `json:"order_id"` // public_id
	UserID       int64
	Step         string `json:"step"`   // какой шаг, т.е. сервис
	Status       string `json:"status"` // SUCCESS/ERROR
	ErrorMessage string `json:"error_message,omitempty"`

	PaymentURL string `json:"payment_url,omitempty"`
	PaymentID  string `json:"payment_id,omitempty"`
}

// Событие для пуша в WebSocket
//
//easyjson:json
type GatewayEvent struct {
	OrderID    string `json:"order_id,omitempty"`
	CartID     string `json:"cart_id,omitempty"`
	EventType  string `json:"event_type,omitempty"`
	Status     string `json:"status,omitempty"`
	PaymentURL string `json:"payment_url,omitempty"`
	Error      string `json:"error,omitempty"`
}
