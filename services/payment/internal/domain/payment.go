package domain

type PaymentMethod struct {
	ID          int64
	UserID      int64
	ExternalID  string
	First6      string
	Last4       string
	ExpiryMonth string
	ExpiryYear  string
	CardType    string
	IssuerName  string
	IsDefault   bool
}
