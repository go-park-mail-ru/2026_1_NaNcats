package yookassa

//go:generate easyjson $GOFILE

import "time"

// Request модели

//easyjson:json
type CreatePaymentMethodRequest struct {
	Type         string                            `json:"type"`
	Confirmation *PaymentMethodRequestConfirmation `json:"confirmation,omitempty"`
}

//easyjson:json
type CreatePaymentRequest struct {
	Amount            CreatePaymentRequestAmount        `json:"amount"`
	Confirmation      *CreatePaymentRequestConfirmation `json:"confirmation,omitempty"`
	Capture           bool                              `json:"capture"`
	SavePaymentMethod bool                              `json:"save_payment_method"`
	Description       string                            `json:"description,omitempty"`
}

//easyjson:json
type PaymentMethodRequestConfirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url"`
}

//easyjson:json
type CreatePaymentRequestAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

//easyjson:json
type CreatePaymentRequestConfirmation struct {
	Type string `json:"type"`
}

// Response модели

//easyjson:json
type PaymentResponse struct {
	ID            string                        `json:"id"`
	Status        string                        `json:"status"`
	Amount        PaymentResponseAmount         `json:"amount"`
	Recipient     PaymentResponseRecipient      `json:"recipient"`
	PaymentMethod *PaymentResponsePaymentMethod `json:"payment_method,omitempty"`
	CreatedAt     time.Time                     `json:"created_at"`
	Confirmation  *PaymentResponseConfirmation  `json:"confirmation,omitempty"`
	Test          bool                          `json:"test"`
	Paid          bool                          `json:"paid"`
	Refundable    bool                          `json:"refundable"`
}

//easyjson:json
type PaymentMethodResponse struct {
	Type         string                             `json:"type"`
	ID           string                             `json:"id"`
	Saved        bool                               `json:"saved"`
	Status       string                             `json:"status"`
	Title        string                             `json:"title,omitempty"`
	Card         *PaymentMethodResponseCard         `json:"card,omitempty"`
	Holder       PaymentMethodResponseHolder        `json:"holder"`
	Confirmation *PaymentMethodResponseConfirmation `json:"confirmation,omitempty"`
}

//easyjson:json
type PaymentResponseAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

//easyjson:json
type PaymentResponseRecipient struct {
	AccountID string `json:"account_id"`
	GatewayID string `json:"gateway_id"`
}

//easyjson:json
type PaymentResponsePaymentMethod struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Saved  bool   `json:"saved"`
	Status string `json:"status"`
}

//easyjson:json
type PaymentResponseConfirmation struct {
	Type              string `json:"type"`
	ConfirmationToken string `json:"confirmation_token"`
}

//easyjson:json
type PaymentMethodResponseCard struct {
	First6      string `json:"first6"`
	Last4       string `json:"last4"`
	ExpiryYear  string `json:"expiry_year"`
	ExpiryMonth string `json:"expiry_month"`
	CardType    string `json:"card_type"`
	IssuerName  string `json:"issuer_name"`
}

//easyjson:json
type PaymentMethodResponseHolder struct {
	AccountID string `json:"account_id"`
}

//easyjson:json
type PaymentMethodResponseConfirmation struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

//easyjson:json
type WebhookNotification struct {
	Type   string                     `json:"type"`
	Event  string                     `json:"event"`
	Object WebhookPaymentMethodObject `json:"object"`
}

//easyjson:json
type WebhookPaymentMethodObject struct {
	ID     string                     `json:"id"`
	Status string                     `json:"status"`
	Saved  bool                       `json:"saved"`
	Type   string                     `json:"type"`
	Card   *PaymentMethodResponseCard `json:"card,omitempty"`
}
