package domain

type Address struct {
	ID             int64
	PublicID       string
	Location       Location
	Apartment      string
	Entrance       string
	Floor          string
	DoorCode       string
	CourierComment string
	Label          string
}
