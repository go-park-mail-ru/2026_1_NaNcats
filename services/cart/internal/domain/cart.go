package domain

import "time"

const (
	CartModeSolo   = "solo"
	CartModeShared = "shared"

	CartStatusActive = "active"
	CartStatusLocked = "locked"
)

type CartItem struct {
	DishID      int64
	Quantity    int32
	Name        string
	Price       int64
	ImageURL    string
	OwnerUserID *int64 // Если nil - позиция ничейная
}

type CartMember struct {
	UserID   int64
	JoinedAt time.Time
}

type Cart struct {
	ID                string
	AdminID           int64
	RestaurantBrandID int64
	Mode              string
	Status            string
	Items             []CartItem
	Members           []CartMember
	UpdatedAt         time.Time
}

// проверка прав
func (c *Cart) HasMember(userID int64) bool {
	for _, m := range c.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

// GetItem возвращает позицию блюда участника ownerID. В совместной корзине у
// одного блюда бывает по позиции на участника, поэтому нужен и владелец.
func (c *Cart) GetItem(dishID, ownerID int64) *CartItem {
	for i := range c.Items {
		if c.Items[i].DishID == dishID && c.Items[i].OwnerUserID != nil && *c.Items[i].OwnerUserID == ownerID {
			return &c.Items[i]
		}
	}
	return nil
}

type CartInvite struct {
	Token     string
	CartID    string
	ExpiresAt time.Time
}

type PaymentIntent struct {
	PayForAll    bool
	PayerMapping map[int64]int64 // Кого -> Кто
}
