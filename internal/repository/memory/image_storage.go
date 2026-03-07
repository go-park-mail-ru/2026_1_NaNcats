package memory

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
)

type imageStorage struct {
	basePath string
}

func NewImageStorage() repository.ImageStorage {
	return &imageStorage{
		basePath: "./uploads",
	}
}

// Функция для генерации нового имени файла в нашей файловой системе
func generateFileName(originalName string) string {
	// Получаем расширение файла (например, .png или .jpg)
	ext := filepath.Ext(originalName)

	newID := uuid.New().String()

	// Склеиваем UUID и расширение
	return newID + ext
}

func (s *imageStorage) Upload(ctx context.Context, fileBytes []byte, originalName, folder string) (string, error) {
	// Определяем путь до папки
	dirPath := filepath.Join(s.basePath, folder)

	// права 0755 - чтение и выполнение для всех + запись для владельца
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", err
	}

	fileName := generateFileName(originalName)
	fullFilePath := filepath.Join(dirPath, fileName)

	// права 0644 - читают все, остальное только у сервера
	err = os.WriteFile(fullFilePath, fileBytes, 0644)
	if err != nil {
		return "", err
	}

	// возвращаем относительный путь до файла
	return filepath.Join(folder, fileName), nil
}

func (s *imageStorage) Download(ctx context.Context, filePath string) ([]byte, error) {
	fullFilePath := filepath.Join(s.basePath, filePath)
	cleanPath := filepath.Clean(fullFilePath)

	photo, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}

	return photo, nil
}
