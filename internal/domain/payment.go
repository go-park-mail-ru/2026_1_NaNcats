package domain

type PaymentMethod struct {
	ID         int    `json:"-"`
	UserID     int    `json:"-"`
	ExternalID string `json:"id"`
	CardType   string `json:"card_type" example:"Mir"`
	Last4      string `json:"last4" example:"6767"`
	IssuerName string `json:"issuer_name,omitempty" example:"Sber"`
	IsDefault  bool   `json:"is_default"`
}
