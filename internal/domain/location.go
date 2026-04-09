package domain

type Location struct {
	ID          int     `json:"id"`
	AddressText string  `json:"address_text"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}
