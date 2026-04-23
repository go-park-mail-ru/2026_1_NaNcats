package events

// Очереди, в которые будет идти запись
const (
	QueueCartCommands    = "saga.cart.commands"
	QueuePaymentCommands = "saga.payment.commands"
	QueueOrderReplies    = "saga.order.replies"  // сюда воркеры ответы пишут
	QueueGatewayEvents   = "saga.gateway.events" // order отсылает сюда инфу для WebSocket
)

// команды оркестратора
const (
	CommandLockCart     = "LOCK_CART"
	CommandUnlockCart   = "UNLOCK_CART"
	CommandClearCart    = "CLEAR_CART"
	CommandCreatPayment = "CREATE_PAYMENT"
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
	Step         string `json:"step"`     // какой шаг, т.е. сервис
	Status       string `json:"status"`   // SUCCESS/ERROR
	ErrorMessage string `json:"error_message,omitempty"`

	PaymentURL string `json:"payment_url,omitempty"`
	PaymentID  string `json:"payment_id,omitempty"`
}

// Событие для пуша в WebSocket
//
//easyjson:json
type GatewayEvent struct {
	OrderID    string `json:"order_id"`
	Status     string `json:"status"` // TODO: продумать и проработать статусы, для выполнения продуктового требования к РК3
	PaymentURL string `json:"payment_url,omitempty"`
	Error      string `json:"error,omitempty"`
}
