package domain

type Address struct {
	ID             int      `json:"-"`
	PublicID       string   `json:"id"`
	Location       Location `json:"location"`
	Apartment      string   `json:"apartment"`
	Entrance       string   `json:"entrance"`
	Floor          string   `json:"floor"`
	DoorCode       string   `json:"door_code"`
	CourierComment string   `json:"courier_comment"`
	Label          string   `json:"label"`
}
