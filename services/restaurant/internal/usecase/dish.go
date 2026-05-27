package usecase

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/imageutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
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
	SearchDishes(ctx context.Context, query string, limit int) ([]domain.Dish, error)
	SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int) ([]domain.Dish, error)
	CreateDish(ctx context.Context, d domain.Dish, image []byte, idemKey string) (domain.Dish, error)
	UpdateDish(ctx context.Context, d domain.Dish, newImage []byte, idemKey string) (domain.Dish, error)
	DeleteDish(ctx context.Context, id int64) error
}

type dishUseCase struct {
	dishRepo            repository.DishRepository
	defaultFoodLogoURL  string
	restaurantBrandRepo repository.RestaurantBrandRepository
	fileStorage         s3.FileStorage
	logger              logger.Logger
}

func NewDishUseCase(dr repository.DishRepository, rbr repository.RestaurantBrandRepository, dflurl string, s3 s3.FileStorage, l logger.Logger) DishUseCase {
	return &dishUseCase{
		dishRepo:            dr,
		restaurantBrandRepo: rbr,
		defaultFoodLogoURL:  dflurl,
		fileStorage:         s3,
		logger:              l,
	}
}

func (uc *dishUseCase) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("restaurant_brand.id", restaurantBrandID),
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	if restaurantBrandID <= 0 {
		return nil, domain.ErrInvalidRestaurantBrandID
	}

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
	uc.applyDefaultLogos(dishes)
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
	uc.applyDefaultLogos(dishes)
	return dishes, nil
}

func (uc *dishUseCase) SearchDishes(ctx context.Context, query string, limit int) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("search.query", query), attribute.Int("limit", limit))

	if limit <= 0 {
		limit = 20
	}

	dishes, err := uc.dishRepo.SearchDishes(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.found_count", len(dishes)))
	uc.applyDefaultLogos(dishes)
	return dishes, nil
}

func (uc *dishUseCase) SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("restaurant_brand.id", brandID),
		attribute.String("search.query", query),
		attribute.Int("limit", limit),
	)

	if limit <= 0 {
		limit = 20
	}

	dishes, err := uc.dishRepo.SearchDishesByBrand(ctx, brandID, query, limit)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.found_count", len(dishes)))
	uc.applyDefaultLogos(dishes)
	return dishes, nil
}

func (uc *dishUseCase) CreateDish(ctx context.Context, d domain.Dish, image []byte, idemKey string) (domain.Dish, error) {
	ownerID, ok := common.GetUserID(ctx)
	if !ok {
		return domain.Dish{}, domain.ErrUnauthorized
	}

	brand, err := uc.restaurantBrandRepo.GetByID(ctx, d.RestaurantBrandID)
	if err != nil {
		return domain.Dish{}, err
	}
	if brand.OwnerProfileID != ownerID {
		return domain.Dish{}, domain.ErrPermissionDenied
	}

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
	ownerID, ok := common.GetUserID(ctx)
	if !ok {
		return domain.Dish{}, domain.ErrUnauthorized
	}

	existing, err := uc.dishRepo.GetDishByID(ctx, d.ID)
	if err != nil {
		return domain.Dish{}, err
	}

	brand, err := uc.restaurantBrandRepo.GetByID(ctx, existing.RestaurantBrandID)
	if err != nil {
		return domain.Dish{}, err
	}
	if brand.OwnerProfileID != ownerID {
		return domain.Dish{}, domain.ErrPermissionDenied
	}

	d.ImageURL = existing.ImageURL

	if len(newImage) > 0 {
		webpData, err := imageutil.ConvertToWebp(bytes.NewReader(newImage))
		if err == nil {
			filename := fmt.Sprintf("foods/%s.webp", uuid.NewString())
			newUrl, uploadErr := uc.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
			if uploadErr == nil {
				if existing.ImageURL != uc.defaultFoodLogoURL && existing.ImageURL != "" {
					go func(oldUrl string) {
						delCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						if err := uc.fileStorage.DeleteFile(delCtx, oldUrl); err != nil {
							uc.logger.Error("failed to delete old food image", err, logger.Any("old_url", oldUrl), logger.Err(err))
						}
					}(existing.ImageURL)
				}
				d.ImageURL = newUrl
			} else {
				uc.logger.Error("failed to upload new food image", uploadErr, logger.Err(uploadErr))
				d.ImageURL = existing.ImageURL
			}
		}
	} else {
		d.ImageURL = existing.ImageURL
	}

	if d.Name == "" {
		d.Name = existing.Name
	}
	if d.Description == "" {
		d.Description = existing.Description
	}
	if d.Price <= 0 {
		d.Price = existing.Price
	}

	d.RestaurantBrandID = existing.RestaurantBrandID

	return uc.dishRepo.Update(ctx, d)
}

func (uc *dishUseCase) DeleteDish(ctx context.Context, id int64) error {
	ownerID, ok := common.GetUserID(ctx)
	if !ok {
		return domain.ErrUnauthorized
	}

	existing, err := uc.dishRepo.GetDishByID(ctx, id)
	if err != nil {
		return err
	}

	brand, err := uc.restaurantBrandRepo.GetByID(ctx, existing.RestaurantBrandID)
	if err != nil {
		return err
	}
	if brand.OwnerProfileID != ownerID {
		return domain.ErrPermissionDenied
	}

	return uc.dishRepo.Delete(ctx, id)
}

func (uc *dishUseCase) applyDefaultLogos(dishes []domain.Dish) {
	for i := range dishes {
		if dishes[i].ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}
}
