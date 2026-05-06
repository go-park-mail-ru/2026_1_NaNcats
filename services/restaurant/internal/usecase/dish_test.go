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
	s3Mocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDishUseCase_GetDishesByRestaurantBrandID(t *testing.T) {
	type mockInit func(m *mocks.MockDishRepository)

	defaultLogo := "http://s3.ru/default-dish.png"
	brandID := int64(1)

	tests := []struct {
		name          string
		brandID       int64
		limit         int
		offset        int
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name:    "Успешное получение блюд",
			brandID: brandID,
			limit:   10,
			offset:  0,
			mockInit: func(m *mocks.MockDishRepository) {
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), brandID, 10, 0).
					Return([]domain.Dish{
						{ID: 1, Name: "Dish 1", ImageURL: "http://s3.ru/1.png"},
						{ID: 2, Name: "Dish 2", ImageURL: ""}, // Проверка подстановки дефолтной картинки
					}, nil)
			},
			expectedLen: 2,
		},
		{
			name:          "Ошибка: невалидный ID бренда",
			brandID:       0,
			mockInit:      func(m *mocks.MockDishRepository) {},
			expectedError: domain.ErrInvalidRestaurantBrandID,
		},
		{
			name:    "Проверка лимитов пагинации (отрицательный лимит и смещение)",
			brandID: brandID,
			limit:   -1,
			offset:  -5,
			mockInit: func(m *mocks.MockDishRepository) {
				// Должно исправиться на 20 и 0
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), brandID, 20, 0).
					Return([]domain.Dish{}, nil)
			},
			expectedLen: 0,
		},
		{
			name:    "Проверка лимитов пагинации (превышение максимума)",
			brandID: brandID,
			limit:   500,
			offset:  10,
			mockInit: func(m *mocks.MockDishRepository) {
				// Должно ограничиться до 100
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), brandID, 100, 10).
					Return([]domain.Dish{}, nil)
			},
			expectedLen: 0,
		},
		{
			name:    "Ошибка репозитория",
			brandID: brandID,
			limit:   20,
			offset:  0,
			mockInit: func(m *mocks.MockDishRepository) {
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), brandID, 20, 0).
					Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDishRepository(ctrl)
			tt.mockInit(repo)

			uc := NewDishUseCase(repo, defaultLogo, nil)
			res, err := uc.GetDishesByRestaurantBrandID(context.Background(), tt.brandID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				// Проверка подстановки дефолтной картинки
				for _, d := range res {
					if d.Name == "Dish 2" {
						assert.Equal(t, defaultLogo, d.ImageURL)
					}
				}
			}
		})
	}
}

func TestDishUseCase_GetDishesByIDs(t *testing.T) {
	type mockInit func(m *mocks.MockDishRepository)

	defaultLogo := "http://s3.ru/default-dish.png"
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
			mockInit: func(m *mocks.MockDishRepository) {
				m.EXPECT().
					GetDishesByIDs(gomock.Any(), ids).
					Return([]domain.Dish{
						{ID: 1, Name: "Dish 1", ImageURL: ""},
						{ID: 2, Name: "Dish 2", ImageURL: "custom.png"},
					}, nil)
			},
			expectedLen: 2,
		},
		{
			name:     "Пустой список ID",
			ids:      []int64{},
			mockInit: func(m *mocks.MockDishRepository) {},
			// Ожидаем ранний возврат (nil, nil)
			expectedLen: 0,
		},
		{
			name: "Ошибка репозитория",
			ids:  ids,
			mockInit: func(m *mocks.MockDishRepository) {
				m.EXPECT().
					GetDishesByIDs(gomock.Any(), ids).
					Return(nil, errors.New("fatal db error"))
			},
			expectedError: errors.New("fatal db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDishRepository(ctrl)
			tt.mockInit(repo)

			uc := NewDishUseCase(repo, defaultLogo, nil)
			res, err := uc.GetDishesByIDs(context.Background(), tt.ids)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, defaultLogo, res[0].ImageURL)
				}
			}
		})
	}
}

func createTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func TestDishUseCase_CreateDish(t *testing.T) {
	type mockInit func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage)

	ctx := context.Background()
	defaultLogo := "http://s3.ru/default-food.png"
	newFoodURL := "http://s3.ru/foods/new-food.webp"
	validImage := createTestImage()
	idemKey := "idem-create-dish"

	dishInput := domain.Dish{
		RestaurantBrandID: 1,
		Name:              "Burger",
		Price:             500,
	}

	tests := []struct {
		name          string
		image         []byte
		mockInit      mockInit
		expectedError string
	}{
		{
			name:  "Успешное создание без фото (дефолтное)",
			image: nil,
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				expectedDish := dishInput
				expectedDish.ImageURL = defaultLogo
				dr.EXPECT().Create(gomock.Any(), expectedDish, idemKey).Return(expectedDish, nil)
			},
		},
		{
			name:  "Успешное создание с фото",
			image: validImage,
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newFoodURL, nil)
				expectedDish := dishInput
				expectedDish.ImageURL = newFoodURL
				dr.EXPECT().Create(gomock.Any(), expectedDish, idemKey).Return(expectedDish, nil)
			},
		},
		{
			name:  "Ошибка: невалидный формат изображения",
			image: []byte("not-an-image"),
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				// Упадет на imageutil.ConvertToWebp
			},
			expectedError: domain.ErrInvalidImageExt.Error(),
		},
		{
			name:  "Ошибка S3 при загрузке",
			image: validImage,
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("s3 fail"))
			},
			expectedError: "failed to upload food image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dr := mocks.NewMockDishRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			tt.mockInit(dr, fs)

			uc := NewDishUseCase(dr, defaultLogo, fs)
			_, err := uc.CreateDish(ctx, dishInput, tt.image, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDishUseCase_UpdateDish(t *testing.T) {
	type mockInit func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage)

	ctx := context.Background()
	defaultLogo := "http://s3.ru/default-food.png"
	oldURL := "http://s3.ru/foods/old.webp"
	newURL := "http://s3.ru/foods/new.webp"
	validImage := createTestImage()
	dishID := int64(10)

	existingDish := domain.Dish{
		ID:       dishID,
		Name:     "Old Name",
		ImageURL: oldURL,
		Price:    100,
	}

	tests := []struct {
		name          string
		input         domain.Dish
		newImage      []byte
		mockInit      mockInit
		expectedError string
	}{
		{
			name: "Успешное частичное обновление (только цена, старое фото остается)",
			input: domain.Dish{
				ID:    dishID,
				Price: 200,
			},
			newImage: nil,
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				dr.EXPECT().GetDishByID(gomock.Any(), dishID).Return(existingDish, nil)
				expected := existingDish
				expected.Price = 200
				dr.EXPECT().Update(gomock.Any(), expected).Return(expected, nil)
			},
		},
		{
			name: "Успешное обновление с новым фото (старое удаляется)",
			input: domain.Dish{
				ID:   dishID,
				Name: "New Name",
			},
			newImage: validImage,
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				dr.EXPECT().GetDishByID(gomock.Any(), dishID).Return(existingDish, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newURL, nil)
				fs.EXPECT().DeleteFile(gomock.Any(), oldURL).Return(nil)

				expected := existingDish
				expected.Name = "New Name"
				expected.ImageURL = newURL
				dr.EXPECT().Update(gomock.Any(), expected).Return(expected, nil)
			},
		},
		{
			name:  "Ошибка: блюдо не найдено",
			input: domain.Dish{ID: 404},
			mockInit: func(dr *mocks.MockDishRepository, fs *s3Mocks.MockFileStorage) {
				dr.EXPECT().GetDishByID(gomock.Any(), int64(404)).Return(domain.Dish{}, domain.ErrDishNotFound)
			},
			expectedError: domain.ErrDishNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dr := mocks.NewMockDishRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			tt.mockInit(dr, fs)

			uc := NewDishUseCase(dr, defaultLogo, fs)
			_, err := uc.UpdateDish(ctx, tt.input, tt.newImage, "idem")

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

func TestDishUseCase_DeleteDish(t *testing.T) {
	type mockInit func(dr *mocks.MockDishRepository)

	ctx := context.Background()
	dishID := int64(1)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление",
			mockInit: func(dr *mocks.MockDishRepository) {
				dr.EXPECT().Delete(gomock.Any(), dishID).Return(nil)
			},
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(dr *mocks.MockDishRepository) {
				dr.EXPECT().Delete(gomock.Any(), dishID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dr := mocks.NewMockDishRepository(ctrl)
			tt.mockInit(dr)

			uc := NewDishUseCase(dr, "", nil)
			err := uc.DeleteDish(ctx, dishID)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
