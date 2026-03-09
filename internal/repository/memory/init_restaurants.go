package memory

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/google/uuid"
)

// seedRestaurants генерирует стартовые данные (30 ресторанов) для in-memory базы
func seedRestaurants() map[uuid.UUID]domain.RestaurantBrand {
	restaurantsMap := make(map[uuid.UUID]domain.RestaurantBrand)

	// Здесь заполняешь свои данные.
	// LogoURL и BannerURL — это те новые поля, которые ты добавил в domain.RestaurantBrand.
	// PromotionTier задавай от 1 до 5 (чтобы протестировать твою сортировку).
	mockData := []domain.RestaurantBrand{
		{
			Name:          "Вкусно - и точка",
			Description:   "Описание 1",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/vkusno_i_tochka.png",
			BannerURL:     "",
		},
		{
			Name:          "Tutta La Vita",
			Description:   "Описание 2",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/tutta_la_vita.png",
			BannerURL:     "",
		},
		{
			Name:          "Izumi",
			Description:   "Описание 3",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/izumi.png",
			BannerURL:     "",
		},
		{
			Name:          "Папа Джонс",
			Description:   "Описание 4",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/papa_jhons.png",
			BannerURL:     "",
		},
		{
			Name:          "Лепим и варим",
			Description:   "Описание 5",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/lepim_i_varim.png",
			BannerURL:     "",
		},
		{
			Name:          "Анна",
			Description:   "Описание 6",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/anna.png",
			BannerURL:     "",
		},
		{
			Name:          "Pinskiy GO",
			Description:   "Описание 7",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/pinsky_go.png",
			BannerURL:     "",
		},
		{
			Name:          "Moro",
			Description:   "Описание 8",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/moro.png",
			BannerURL:     "",
		},
		{
			Name:          "DiDi",
			Description:   "Описание 9",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/didi.png",
			BannerURL:     "",
		},
		{
			Name:          "Калифорния Дайнер",
			Description:   "Описание 10",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/kaliforniya_dainer.png",
			BannerURL:     "",
		},
		{
			Name:          "Ketch_up",
			Description:   "Описание 11",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/ketch_up.png",
			BannerURL:     "",
		},
		{
			Name:          "Villa Pasta",
			Description:   "Описание 12",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/villa_pasta.png",
			BannerURL:     "",
		},
		{
			Name:          "Китайские Новости",
			Description:   "Описание 13",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/kitayskie_novosti.png",
			BannerURL:     "",
		},
		{
			Name:          "Братья Караваевы",
			Description:   "Описание 14",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/bratya_karavaevy.png",
			BannerURL:     "",
		},
		{
			Name:          "Империя Пиццы",
			Description:   "Описание 15",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/imperiya_pizzi.png",
			BannerURL:     "",
		},
		{
			Name:          "Крошка Картошка",
			Description:   "Описание 16",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/kroshka_kartoshka.png",
			BannerURL:     "",
		},
		{
			Name:          "Раменная Ru-Rik",
			Description:   "Описание 17",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/ramennaya_ru_rik.png",
			BannerURL:     "",
		},
		{
			Name:          "El Chapo Burgers Tacos&Burritos",
			Description:   "Описание 18",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/el_Chapo_Burgers_Tacos&Burritos.png",
			BannerURL:     "",
		},
		{
			Name:          "Subway",
			Description:   "Описание 19",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/subway.png",
			BannerURL:     "",
		},
		{
			Name:          "Аист",
			Description:   "Описание 20",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/aist.png",
			BannerURL:     "",
		},
		{
			Name:          "Ванлав",
			Description:   "Описание 21",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/vanlav.png",
			BannerURL:     "",
		},
		{
			Name:          "Varvarka III",
			Description:   "Описание 22",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/Varvarka_III.png",
			BannerURL:     "",
		},
		{
			Name:          "Техникум",
			Description:   "Описание 23",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/Technikum.png",
			BannerURL:     "",
		},
		{
			Name:          "Ариум Grill",
			Description:   "Описание 24",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/arium_grill.png",
			BannerURL:     "",
		},
		{
			Name:          "Честная Рыба",
			Description:   "Описание 25",
			PromotionTier: 4,
			LogoURL:       "restaurants/logos/chestnaya_ryba.png",
			BannerURL:     "",
		},
		{
			Name:          "Eshak",
			Description:   "Описание 26",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/eshak.png",
			BannerURL:     "",
		},
		{
			Name:          "Руки ВВерх!",
			Description:   "Описание 27",
			PromotionTier: 1,
			LogoURL:       "restaurants/logos/ryki_vverh.png",
			BannerURL:     "",
		},
		{
			Name:          "Такахули",
			Description:   "Описание 28",
			PromotionTier: 5,
			LogoURL:       "restaurants/logos/takahyli.png",
			BannerURL:     "",
		},
		{
			Name:          "Машрумс",
			Description:   "Описание 29",
			PromotionTier: 3,
			LogoURL:       "restaurants/logos/mashrums.png",
			BannerURL:     "",
		},
		{
			Name:          "FoodBand",
			Description:   "Описание 30",
			PromotionTier: 2,
			LogoURL:       "restaurants/logos/FoodBand.png",
			BannerURL:     "",
		},
	}

	for _, rest := range mockData {
		rest.ID = uuid.New()
		rest.CreatedAt = time.Now()
		rest.UpdatedAt = time.Now()

		restaurantsMap[rest.ID] = rest
	}

	return restaurantsMap
}
