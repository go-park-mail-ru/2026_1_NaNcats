package domain

type PaymentMethod struct {
	ID         int64
	UserID     int64
	ExternalID string
	CardType   string
	Last4      string
	IssuerName string
	IsDefault  bool
}
