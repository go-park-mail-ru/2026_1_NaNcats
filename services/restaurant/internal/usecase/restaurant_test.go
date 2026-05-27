package usecase

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	s3Mocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRestaurantBrandUseCase_GetRestaurantBrandsByIDs(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantBrandRepository)

	defaultLogo := "http://s3.ru/default-brand.png"
	ids := []int64{1, 2}

	tests := []struct {
		name          string
		ids           []int64
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name: "Успешное получение по списку ID",
			ids:  ids,
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsByIDs(gomock.Any(), ids).
					Return([]domain.RestaurantBrand{
						{ID: 1, Name: "KFC", LogoURL: ""}, // Проверка подстановки дефолтной картинки
						{ID: 2, Name: "Burger King", LogoURL: "custom.png"},
					}, nil)
			},
			expectedLen: 2,
		},
		{
			name: "Ошибка репозитория",
			ids:  ids,
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsByIDs(gomock.Any(), ids).
					Return(nil, errors.New("db fail"))
			},
			expectedError: errors.New("db fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, nil, nil)
			res, err := uc.GetRestaurantBrandsByIDs(context.Background(), tt.ids)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, defaultLogo, res[0].LogoURL)
				}
			}
		})
	}
}

func TestRestaurantBrandUseCase_GetRestaurantBrandsList(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantBrandRepository)

	defaultLogo := "http://s3.ru/default-brand.png"
	limit, offset := 10, 0

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name: "Успешное получение списка",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), limit, offset).
					Return([]domain.RestaurantBrand{
						{ID: 1, Name: "Mac", LogoURL: ""},
					}, nil)
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), limit, offset).
					Return(nil, errors.New("connection reset"))
			},
			expectedError: errors.New("connection reset"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, nil, nil)
			res, err := uc.GetRestaurantBrandsList(context.Background(), limit, offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, defaultLogo, res[0].LogoURL)
				}
			}
		})
	}
}

func TestRestaurantBrandUseCase_GetRestaurantBrandsByCategoryName(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantBrandRepository)

	defaultLogo := "http://s3.ru/default-brand.png"
	category := "Fastfood"
	limit, offset := 10, 0

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name: "Успешное получение по категории",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsByCategoryName(gomock.Any(), category, limit, offset).
					Return([]domain.RestaurantBrand{
						{ID: 1, Name: "Teremok", LogoURL: ""},
					}, nil)
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetRestaurantBrandsByCategoryName(gomock.Any(), category, limit, offset).
					Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, nil, nil)
			res, err := uc.GetRestaurantBrandsByCategoryName(context.Background(), category, limit, offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, defaultLogo, res[0].LogoURL)
				}
			}
		})
	}
}

func TestRestaurantBrandUseCase_SearchRestaurantBrands(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantBrandRepository)

	defaultLogo := "http://s3.ru/default-brand.png"
	query := "Pizza"
	limit, offset := 10, 0

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name: "Успешный поиск",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					SearchRestaurantBrands(gomock.Any(), query, limit, offset).
					Return([]domain.RestaurantBrand{
						{ID: 1, Name: "Dodo Pizza", LogoURL: ""},
					}, nil)
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					SearchRestaurantBrands(gomock.Any(), query, limit, offset).
					Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, nil, nil)
			res, err := uc.SearchRestaurantBrands(context.Background(), query, limit, offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, defaultLogo, res[0].LogoURL)
				}
			}
		})
	}
}

func TestRestaurantBrandUseCase_GetRestaurantBrandByID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantBrandRepository)

	defaultLogo := "http://s3.ru/default-brand.png"
	brandID := int64(42)

	tests := []struct {
		name          string
		id            int64
		mockInit      mockInit
		expectedName  string
		expectedError string
	}{
		{
			name: "Успешное получение по ID",
			id:   brandID,
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), brandID).
					Return(domain.RestaurantBrand{ID: brandID, Name: "Subway", LogoURL: ""}, nil)
			},
			expectedName: "Subway",
		},
		{
			name:          "Ошибка: некорректный ID (ноль)",
			id:            0,
			mockInit:      func(m *mocks.MockRestaurantBrandRepository) {},
			expectedError: "invalid restaurant brand id",
		},
		{
			name:          "Ошибка: некорректный ID (отрицательный)",
			id:            -1,
			mockInit:      func(m *mocks.MockRestaurantBrandRepository) {},
			expectedError: "invalid restaurant brand id",
		},
		{
			name: "Ошибка: ресторан не найден",
			id:   brandID,
			mockInit: func(m *mocks.MockRestaurantBrandRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), brandID).
					Return(domain.RestaurantBrand{}, domain.ErrRestaurantNotFound)
			},
			expectedError: "restaurant not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, nil, nil)
			res, err := uc.GetRestaurantBrandByID(context.Background(), tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, res.Name)
				assert.Equal(t, defaultLogo, res.LogoURL)
			}
		})
	}
}

func createValidImageBytes() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func TestRestaurantBrandUseCase_CreateRestaurantBrand(t *testing.T) {
	type mockInit func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage)

	ownerID := int64(42)
	ctx := context.WithValue(context.Background(), common.UserIDKey, ownerID)
	defaultLogo := "http://s3.ru/default-logo.png"
	newLogoURL := "http://s3.ru/restaurants/new-logo.webp"
	validImage := createValidImageBytes()
	idemKey := "idem-create"

	brandInput := domain.RestaurantBrand{
		OwnerProfileID: ownerID,
		Name:           "Burger Heroes",
		Description:    "Tasty",
	}

	tests := []struct {
		name          string
		image         []byte
		mockInit      mockInit
		expectedError string
	}{
		{
			name:  "Успешное создание без фото (используется дефолтное)",
			image: nil,
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				expected := brandInput
				expected.LogoURL = defaultLogo
				expected.PromotionTier = 1
				r.EXPECT().Create(gomock.Any(), expected, idemKey).Return(expected, nil)
			},
		},
		{
			name:  "Успешное создание с фото",
			image: validImage,
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newLogoURL, nil)
				expected := brandInput
				expected.LogoURL = newLogoURL
				expected.PromotionTier = 1
				r.EXPECT().Create(gomock.Any(), expected, idemKey).Return(expected, nil)
			},
		},
		{
			name:          "Ошибка: невалидный формат изображения",
			image:         []byte("definitely-not-an-image"),
			mockInit:      func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {},
			expectedError: domain.ErrInvalidImageExt.Error(),
		},
		{
			name:  "Ошибка S3 при загрузке логотипа",
			image: validImage,
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return("", errors.New("s3 fail"))
			},
			expectedError: "failed to upload logo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			tt.mockInit(repo, fs)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, fs, nil)
			_, err := uc.CreateRestaurantBrand(ctx, brandInput, tt.image, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestaurantBrandUseCase_UpdateRestaurantBrand(t *testing.T) {
	type mockInit func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage)

	ownerID := int64(42)
	ctx := context.WithValue(context.Background(), common.UserIDKey, ownerID)
	defaultLogo := "http://s3.ru/default-logo.png"
	oldLogoURL := "http://s3.ru/restaurants/old.webp"
	newLogoURL := "http://s3.ru/restaurants/new.webp"
	brandID := int64(1)
	validImage := createValidImageBytes()

	existingBrand := domain.RestaurantBrand{
		ID:             brandID,
		OwnerProfileID: ownerID,
		Name:           "Old Name",
		Description:    "Old Desc",
		LogoURL:        oldLogoURL,
		PromotionTier:  1,
	}

	tests := []struct {
		name          string
		input         domain.RestaurantBrand
		newImage      []byte
		mockInit      mockInit
		expectedError string
	}{
		{
			name: "Успешное частичное обновление (без фото)",
			input: domain.RestaurantBrand{
				ID:   brandID,
				Name: "New Name",
			},
			newImage: nil,
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				r.EXPECT().GetByID(gomock.Any(), brandID).Return(existingBrand, nil)
				expected := existingBrand
				expected.Name = "New Name"
				r.EXPECT().Update(gomock.Any(), expected).Return(expected, nil)
			},
		},
		{
			name: "Успешное обновление с новым фото (старое удаляется)",
			input: domain.RestaurantBrand{
				ID: brandID,
			},
			newImage: validImage,
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				r.EXPECT().GetByID(gomock.Any(), brandID).Return(existingBrand, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newLogoURL, nil)
				fs.EXPECT().DeleteFile(gomock.Any(), oldLogoURL).Return(nil)

				expected := existingBrand
				expected.LogoURL = newLogoURL
				r.EXPECT().Update(gomock.Any(), expected).Return(expected, nil)
			},
		},
		{
			name:  "Ошибка: бренд не найден",
			input: domain.RestaurantBrand{ID: 404},
			mockInit: func(r *mocks.MockRestaurantBrandRepository, fs *s3Mocks.MockFileStorage) {
				r.EXPECT().GetByID(gomock.Any(), int64(404)).Return(domain.RestaurantBrand{}, domain.ErrRestaurantNotFound)
			},
			expectedError: domain.ErrRestaurantNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			tt.mockInit(repo, fs)

			uc := NewRestaurantBrandUseCase(repo, defaultLogo, fs, nil)
			_, err := uc.UpdateRestaurantBrand(ctx, tt.input, tt.newImage, "idem")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestRestaurantBrandUseCase_DeleteRestaurantBrand(t *testing.T) {
	type mockInit func(r *mocks.MockRestaurantBrandRepository)

	ownerID := int64(42)
	ctx := context.WithValue(context.Background(), common.UserIDKey, ownerID)
	brandID := int64(1)
	ownedBrand := domain.RestaurantBrand{ID: brandID, OwnerProfileID: ownerID}

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление",
			mockInit: func(r *mocks.MockRestaurantBrandRepository) {
				r.EXPECT().GetByID(gomock.Any(), brandID).Return(ownedBrand, nil)
				r.EXPECT().Delete(gomock.Any(), brandID).Return(nil)
			},
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(r *mocks.MockRestaurantBrandRepository) {
				r.EXPECT().GetByID(gomock.Any(), brandID).Return(ownedBrand, nil)
				r.EXPECT().Delete(gomock.Any(), brandID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRestaurantBrandRepository(ctrl)
			tt.mockInit(repo)

			uc := NewRestaurantBrandUseCase(repo, "", nil, nil)
			err := uc.DeleteRestaurantBrand(ctx, brandID)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
