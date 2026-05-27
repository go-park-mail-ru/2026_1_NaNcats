package domain

// Локальная доменная модель бренда для слоя аналитики
type RestaurantBrand struct {
	ID             int64
	OwnerProfileID int64
	Name           string
}
