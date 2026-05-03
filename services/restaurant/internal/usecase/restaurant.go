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
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase RestaurantBrandUseCase
//go:generate gowrap gen -i RestaurantBrandUseCase -t ../../../../shared/templates/tracing.tmpl -o restaurant_brand_tracing_mw.go -v TracerName=restaurant-service
type RestaurantBrandUseCase interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int64) (domain.RestaurantBrand, error)
	GetRestaurantBrandsByIDs(ctx context.Context, brandIDs []int64) ([]domain.RestaurantBrand, error)
	CreateRestaurantBrand(ctx context.Context, b domain.RestaurantBrand, image []byte, idemKey string) (domain.RestaurantBrand, error)
	UpdateRestaurantBrand(ctx context.Context, b domain.RestaurantBrand, newImage []byte, idemKey string) (domain.RestaurantBrand, error)
	DeleteRestaurantBrand(ctx context.Context, id int64) error
}

type restaurantBrandUseCase struct {
	restaurantBrandRepo      repository.RestaurantBrandRepository
	defaultRestaurantLogoURL string
	fileStorage              s3.FileStorage
}

func NewRestaurantBrandUseCase(rbr repository.RestaurantBrandRepository, drlurl string, s3 s3.FileStorage) RestaurantBrandUseCase {
	return &restaurantBrandUseCase{
		restaurantBrandRepo:      rbr,
		defaultRestaurantLogoURL: drlurl,
		fileStorage:              s3,
	}
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsByIDs(ctx context.Context, brandIDs []int64) ([]domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64Slice("brand.ids", brandIDs))

	brands, err := rb.restaurantBrandRepo.GetRestaurantBrandsByIDs(ctx, brandIDs)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("brands.count", len(brands)))

	for i, brand := range brands {
		if brand.LogoURL == "" {
			brands[i].LogoURL = rb.defaultRestaurantLogoURL
		}
	}

	return brands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	restaurantBrands, err := rb.restaurantBrandRepo.GetRestaurantBrandsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("brands.count", len(restaurantBrands)))

	for i, restaurantBrand := range restaurantBrands {
		if restaurantBrand.LogoURL == "" {
			restaurantBrands[i].LogoURL = rb.defaultRestaurantLogoURL
		}
	}

	return restaurantBrands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandByID(ctx context.Context, id int64) (domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("brand.id", id))

	if id <= 0 {
		return domain.RestaurantBrand{}, errutil.New("INVALID_RESTAURANT_BRAND_ID", "invalid restaurant brand id", codes.InvalidArgument)
	}

	restaurantBrand, err := rb.restaurantBrandRepo.GetByID(ctx, id)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}

	if restaurantBrand.LogoURL == "" {
		restaurantBrand.LogoURL = rb.defaultRestaurantLogoURL
	}

	return restaurantBrand, nil
}

func (uc *restaurantBrandUseCase) CreateRestaurantBrand(ctx context.Context, b domain.RestaurantBrand, image []byte, idemKey string) (domain.RestaurantBrand, error) {
	// Загрузка логотипа в S3
	if len(image) > 0 {
		webpData, err := imageutil.ConvertToWebp(bytes.NewReader(image))
		if err != nil {
			return domain.RestaurantBrand{}, domain.ErrInvalidImageExt
		}

		filename := fmt.Sprintf("restaurants/%s.webp", uuid.NewString())
		url, err := uc.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
		if err != nil {
			return domain.RestaurantBrand{}, errutil.Internal("failed to upload logo", err)
		}
		b.LogoURL = url
	} else {
		b.LogoURL = uc.defaultRestaurantLogoURL
	}

	// Сохранение в БД
	return uc.restaurantBrandRepo.Create(ctx, b, idemKey)
}

func (uc *restaurantBrandUseCase) UpdateRestaurantBrand(ctx context.Context, b domain.RestaurantBrand, newImage []byte, idemKey string) (domain.RestaurantBrand, error) {
	existing, err := uc.restaurantBrandRepo.GetByID(ctx, b.ID)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}

	// Если пришло новое изображение - меняем его в S3
	if len(newImage) > 0 {
		webpData, err := imageutil.ConvertToWebp(bytes.NewReader(newImage))
		if err == nil {
			filename := fmt.Sprintf("restaurants/%s.webp", uuid.NewString())
			newUrl, uploadErr := uc.fileStorage.UploadFile(ctx, webpData, filename, "image/webp")
			if uploadErr == nil {
				// Удаляем старое фото, если оно не дефолтное
				if existing.LogoURL != uc.defaultRestaurantLogoURL && existing.LogoURL != "" {
					go func() { _ = uc.fileStorage.DeleteFile(context.Background(), existing.LogoURL) }()
				}
				b.LogoURL = newUrl
			}
		}
	} else {
		b.LogoURL = existing.LogoURL
	}

	// Заполняем остальные пустые поля из существующей записи
	if b.Name == "" {
		b.Name = existing.Name
	}
	if b.Description == "" {
		b.Description = existing.Description
	}
	if b.PromotionTier == 0 {
		b.PromotionTier = existing.PromotionTier
	}

	return uc.restaurantBrandRepo.Update(ctx, b)
}

func (rb *restaurantBrandUseCase) DeleteRestaurantBrand(ctx context.Context, id int64) error {
	return rb.restaurantBrandRepo.Delete(ctx, id)
}
