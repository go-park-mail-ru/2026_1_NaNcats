package domain

type Address struct {
	ID             int
	PublicID       string
	Location       Location
	Apartment      string
	Entrance       string
	Floor          string
	DoorCode       string
	CourierComment string
	Label          string
}
