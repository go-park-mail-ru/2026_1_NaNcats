package usecase

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/imageutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/dish_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase DishUseCase
//go:generate gowrap gen -i DishUseCase -t ../../../../shared/templates/tracing.tmpl -o dish_tracing_mw.go -v TracerName=restaurant-service
type DishUseCase interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error)
	GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error)
	CreateDish(ctx context.Context, d domain.Dish, image []byte, idemKey string) (domain.Dish, error)
	UpdateDish(ctx context.Context, d domain.Dish, newImage []byte, idemKey string) (domain.Dish, error)
	DeleteDish(ctx context.Context, id int64) error
}

type dishUseCase struct {
	dishRepo           repository.DishRepository
	defaultFoodLogoURL string
	fileStorage        s3.FileStorage
}

func NewDishUseCase(dr repository.DishRepository, dflurl string, s3 s3.FileStorage) DishUseCase {
	return &dishUseCase{
		dishRepo:           dr,
		defaultFoodLogoURL: dflurl,
		fileStorage:        s3,
	}
}

func (uc *dishUseCase) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("restaurant_brand.id", restaurantBrandID),
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	// Валидация входа
	if restaurantBrandID <= 0 {
		return nil, domain.ErrInvalidRestaurantBrandID
	}

	// Пагинация
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	dishes, err := uc.dishRepo.GetDishesByRestaurantBrandID(ctx, restaurantBrandID, limit, offset)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.returned_count", len(dishes)))

	for i, dish := range dishes {
		if dish.ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}

func (uc *dishUseCase) GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("dishes.requested_count", len(ids)))

	if len(ids) == 0 {
		return nil, nil
	}

	dishes, err := uc.dishRepo.GetDishesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.found_count", len(dishes)))

	for i := range dishes {
		if dishes[i].ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}

func (uc *dishUseCase) CreateDish(ctx context.Context, d domain.Dish, image []byte, idemKey string) (domain.Dish, error) {
	if len(image) > 0 {
		webpData, err := imageutil.ConvertToWebp(bytes.NewReader(image))
		if err != nil {
			return domain.Dish{}, domain.ErrInvalidImageExt
		}

		filename := fmt.Sprintf("foods/%s.webp", uuid.NewString())
		url, err := uc.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
		if err != nil {
			return domain.Dish{}, errutil.Internal("failed to upload food image", err)
		}
		d.ImageURL = url
	} else {
		d.ImageURL = uc.defaultFoodLogoURL
	}

	return uc.dishRepo.Create(ctx, d, idemKey)
}

func (uc *dishUseCase) UpdateDish(ctx context.Context, d domain.Dish, newImage []byte, idemKey string) (domain.Dish, error) {
	existing, err := uc.dishRepo.GetDishByID(ctx, d.ID)
	if err != nil {
		return domain.Dish{}, err
	}

	if len(newImage) > 0 {
		webpData, err := imageutil.ConvertToWebp(bytes.NewReader(newImage))
		if err == nil {
			filename := fmt.Sprintf("foods/%s.webp", uuid.NewString())
			newUrl, uploadErr := uc.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
			if uploadErr == nil {
				if existing.ImageURL != uc.defaultFoodLogoURL && existing.ImageURL != "" {
					go func() { _ = uc.fileStorage.DeleteFile(context.Background(), existing.ImageURL) }()
				}
				d.ImageURL = newUrl
			}
		}
	} else {
		d.ImageURL = existing.ImageURL
	}

	// Частичное обновление полей
	if d.Name == "" {
		d.Name = existing.Name
	}
	if d.Description == "" {
		d.Description = existing.Description
	}
	if d.Price <= 0 {
		d.Price = existing.Price
	}

	return uc.dishRepo.Update(ctx, d)
}

func (uc *dishUseCase) DeleteDish(ctx context.Context, id int64) error {
	return uc.dishRepo.Delete(ctx, id)
}
