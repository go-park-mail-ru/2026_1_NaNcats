package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestS3Storage_UploadFile(t *testing.T) {
	type mockInit func(m *mocks.MockS3ClientAPI)

	tests := []struct {
		name        string
		filename    string
		contentType string
		fileContent []byte
		mockInit    mockInit
		expectedURL string
		expectError bool
	}{
		{
			name:        "Успешная загрузка файла",
			filename:    "avatar.png",
			contentType: "image/png",
			fileContent: []byte("fake-image-bytes"),
			mockInit: func(m *mocks.MockS3ClientAPI) {
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						assert.Equal(t, "my-test-bucket", *params.Bucket)
						assert.Equal(t, "avatar.png", *params.Key)
						assert.Equal(t, "image/png", *params.ContentType)
						return &s3.PutObjectOutput{}, nil
					})
			},
			expectedURL: "https://my-test-bucket.storage.yandexcloud.net/avatar.png",
			expectError: false,
		},
		{
			name:        "Ошибка клиента S3 при загрузке",
			filename:    "error.png",
			contentType: "image/png",
			fileContent: []byte("fake-image-bytes"),
			mockInit: func(m *mocks.MockS3ClientAPI) {
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("s3 internal error"))
			},
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3 := mocks.NewMockS3ClientAPI(ctrl)
			tt.mockInit(mockS3)

			storage := &s3Storage{
				client: mockS3,
				bucket: "my-test-bucket",
				region: "ru-central1",
			}

			reader := bytes.NewReader(tt.fileContent)

			url, err := storage.UploadFile(context.Background(), reader, tt.filename, tt.contentType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestS3Storage_DeleteFile(t *testing.T) {
	type mockInit func(m *mocks.MockS3ClientAPI)

	tests := []struct {
		name        string
		fileURL     string
		mockInit    mockInit
		expectError bool
	}{
		{
			name:    "Успешное удаление файла",
			fileURL: "https://my-test-bucket.storage.yandexcloud.net/images/avatar.png",
			mockInit: func(m *mocks.MockS3ClientAPI) {
				m.EXPECT().DeleteObject(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						assert.Equal(t, "my-test-bucket", *params.Bucket)
						assert.Equal(t, "images/avatar.png", *params.Key)
						return &s3.DeleteObjectOutput{}, nil
					})
			},
			expectError: false,
		},
		{
			name:        "Ошибка: невалидный URL",
			fileURL:     "://bad-url",
			mockInit:    func(m *mocks.MockS3ClientAPI) {}, // Вызовов к S3 быть не должно
			expectError: true,
		},
		{
			name:        "Ошибка: пустой ключ (файл не указан)",
			fileURL:     "https://my-test-bucket.storage.yandexcloud.net/",
			mockInit:    func(m *mocks.MockS3ClientAPI) {}, // Вызовов к S3 быть не должно
			expectError: true,
		},
		{
			name:    "Ошибка клиента S3 при удалении",
			fileURL: "https://my-test-bucket.storage.yandexcloud.net/fail.png",
			mockInit: func(m *mocks.MockS3ClientAPI) {
				m.EXPECT().DeleteObject(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("s3 delete error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3 := mocks.NewMockS3ClientAPI(ctrl)
			tt.mockInit(mockS3)

			storage := &s3Storage{
				client: mockS3,
				bucket: "my-test-bucket",
				region: "ru-central1",
			}

			err := storage.DeleteFile(context.Background(), tt.fileURL)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
