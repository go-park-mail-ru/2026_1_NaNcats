package request

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	// Вспомогательная структура для парсинга в тестах
	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string // Описание теста
		body        string // Входящий JSON
		wantErr     error  // Ожидаемая конкретная ошибка (для errors.Is)
		errContains string // Подстрока, которая должна быть в тексте ошибки (для динамических ошибок)
	}{
		{
			name:    "Успешный парсинг",
			body:    `{"name":"Ivan", "age":25}`,
			wantErr: nil,
		},
		{
			name:    "Ошибка: пустое тело",
			body:    "",
			wantErr: ErrEmptyBody,
		},
		{
			name:    "Ошибка: некорректный JSON (синтаксис)",
			body:    `{"name": "Ivan"`, // пропущена скобка
			wantErr: ErrInvalidJSON,
		},
		{
			name:        "Ошибка: неверный тип поля",
			body:        `{"name": 123}`, // в Name (string) суем число
			errContains: "incorrect JSON type",
		},
		{
			name:    "Ошибка: лишние данные после JSON",
			body:    `{"name":"Ivan"}{"age":20}`,
			wantErr: ErrNotOnlyJSONVal,
		},
		{
			name:    "Ошибка: мусор вместо JSON",
			body:    `not a json at all`,
			wantErr: ErrInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем имитацию запроса
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			nopLogger := mocks.NewNopLogger()

			var data testData
			err := JSON(req, &data, nopLogger)

			if tt.wantErr != nil {
				// Проверка по конкретной переменной (Sentinel error)
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.errContains != "" {
				// Проверка по тексту (для fmt.Errorf)
				assert.Error(t, err) // Убеждаемся, что ошибка вообще есть
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				// Успешный кейс
				assert.NoError(t, err)
			}
		})
	}
}

// Отдельный тест для проверки лимита размера,
// так как генерировать 1МБ в таблице неудобно
func TestJSON_MaxSize(t *testing.T) {
	// Создаем тело ровно 1МБ + 1 байт
	largeBody := `"` + strings.Repeat("a", 1024*1024+1) + `"`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))

	nopLogger := mocks.NewNopLogger()

	var data struct{}
	err := JSON(req, &data, nopLogger)

	assert.ErrorIs(t, err, ErrBodyTooLarge)
}
