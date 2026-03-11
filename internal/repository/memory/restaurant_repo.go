package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
)

// структура брендов ресторанов на основе мап
type restaurantBrandRepo struct {
	mu               sync.RWMutex                         // защита от записи во время чтения из мапы
	restaurantBrands map[uuid.UUID]domain.RestaurantBrand // мапа ресторанов, ключ - sessionID
}

// функция-конструктор репозитория сессий
func NewRestaurantBrandRepo() repository.RestaurantBrandRepository {
	return &restaurantBrandRepo{
		restaurantBrands: seedRestaurants(),
	}
}

func (r *restaurantBrandRepo) GetRestaurantBrandsList(ctx context.Context, limit, offset int) []domain.RestaurantBrand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем слайс для ресторанных брендов той же длины, что и сама мапа
	restaurantBrandsSlice := make([]domain.RestaurantBrand, 0, len(r.restaurantBrands))

	// Переносим в слайс данные из мапы
	for _, currRestaurantBrand := range r.restaurantBrands {
		restaurantBrandsSlice = append(restaurantBrandsSlice, currRestaurantBrand)
	}

	sort.Slice(restaurantBrandsSlice, func(i, j int) bool {
		if restaurantBrandsSlice[i].PromotionTier != restaurantBrandsSlice[j].PromotionTier {
			// Сортируем по убыванию (от большего к меньшему) по уровню продвижения
			return restaurantBrandsSlice[i].PromotionTier > restaurantBrandsSlice[j].PromotionTier
		}

		// Второй параметр сортировки - название ресторана (по возрастанию)
		return restaurantBrandsSlice[i].Name < restaurantBrandsSlice[j].Name
	})

	total := len(restaurantBrandsSlice)

	if offset >= total {
		return []domain.RestaurantBrand{}
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return restaurantBrandsSlice[offset:end]
}
