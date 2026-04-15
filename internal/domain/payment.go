package domain

type PaymentMethod struct {
	ID         int
	UserID     int
	ExternalID string
	CardType   string
	Last4      string
	IssuerName string
	IsDefault  bool
}
