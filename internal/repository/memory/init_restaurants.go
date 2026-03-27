package memory

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// seedRestaurants генерирует стартовые данные (30 ресторанов) для in-memory базы
func seedRestaurants() map[int]domain.RestaurantBrand {
	restaurantsMap := make(map[int]domain.RestaurantBrand)

	mockData := []domain.RestaurantBrand{
		{
			Name:          "Вкусно - и точка",
			Description:   "Описание 1",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/vkusno_i_tochka.png",
		},
		{
			Name:          "Tutta La Vita",
			Description:   "Описание 2",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/tutta_la_vita.png",
		},
		{
			Name:          "Izumi",
			Description:   "Описание 3",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/izumi.png",
		},
		{
			Name:          "Папа Джонс",
			Description:   "Описание 4",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/papa_jhons.png",
		},
		{
			Name:          "Лепим и варим",
			Description:   "Описание 5",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/lepim_i_varim.png",
		},
		{
			Name:          "Анна",
			Description:   "Описание 6",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/anna.png",
		},
		{
			Name:          "Pinskiy GO",
			Description:   "Описание 7",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/pinsky_go.png",
		},
		{
			Name:          "Moro",
			Description:   "Описание 8",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/moro.png",
		},
		{
			Name:          "DiDi",
			Description:   "Описание 9",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/didi.png",
		},
		{
			Name:          "Калифорния Дайнер",
			Description:   "Описание 10",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/kaliforniya_dainer.png",
		},
		{
			Name:          "Ketch up",
			Description:   "Описание 11",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/ketch_up.png",
		},
		{
			Name:          "Villa Pasta",
			Description:   "Описание 12",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/villa_pasta.png",
		},
		{
			Name:          "Китайские Новости",
			Description:   "Описание 13",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/kitayskie_novosti.png",
		},
		{
			Name:          "Братья Караваевы",
			Description:   "Описание 14",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/bratya_karavaevy.png",
		},
		{
			Name:          "Империя Пиццы",
			Description:   "Описание 15",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/imperiya_pizzi.png",
		},
		{
			Name:          "Крошка Картошка",
			Description:   "Описание 16",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/kroshka_kartoshka.png",
		},
		{
			Name:          "Раменная Ru-Rik",
			Description:   "Описание 17",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/ramennaya_ru_rik.png",
		},
		{
			Name:          "El Chapo Burgers Tacos&Burritos",
			Description:   "Описание 18",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/el_Chapo_Burgers_Tacos&Burritos.png",
		},
		{
			Name:          "Subway",
			Description:   "Описание 19",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/subway.png",
		},
		{
			Name:          "Аист",
			Description:   "Описание 20",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/aist.png",
		},
		{
			Name:          "Ванлав",
			Description:   "Описание 21",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/vanlav.png",
		},
		{
			Name:          "Varvarka III",
			Description:   "Описание 22",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/Varvarka_III.png",
		},
		{
			Name:          "Техникум",
			Description:   "Описание 23",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/Technikum.png",
		},
		{
			Name:          "Ариум Grill",
			Description:   "Описание 24",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/arium_grill.png",
		},
		{
			Name:          "Честная Рыба",
			Description:   "Описание 25",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/chestnaya_ryba.png",
		},
		{
			Name:          "Eshak",
			Description:   "Описание 26",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/eshak.png",
		},
		{
			Name:          "Руки ВВерх!",
			Description:   "Описание 27",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/ryki_vverh.png",
		},
		{
			Name:          "Такахули",
			Description:   "Описание 28",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/takahyli.png",
		},
		{
			Name:          "Машрумс",
			Description:   "Описание 29",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/mashrums.png",
		},
		{
			Name:          "FoodBand",
			Description:   "Описание 30",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/FoodBand.png",
		},
	}

	for i, rest := range mockData {
		rest.ID = i + 1
		rest.CreatedAt = time.Now()
		rest.UpdatedAt = time.Now()

		restaurantsMap[rest.ID] = rest
	}

	return restaurantsMap
}
