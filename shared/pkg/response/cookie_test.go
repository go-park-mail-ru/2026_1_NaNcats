package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetCookie(t *testing.T) {
	// Подготовка данных
	rec := httptest.NewRecorder()
	cookieName := "session_id"
	cookieValue := "test-uuid-12345"
	// Округляем время до секунды, так как в HTTP заголовках точность до секунд
	expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	SetCookie(rec, cookieName, cookieValue, expiresAt)

	// Получение результата из рекордера
	// rec.Result() возвращает объект *http.Response, у которого есть удобный метод Cookies()
	resp := rec.Result()
	cookies := resp.Cookies()

	assert.Len(t, cookies, 1, "Должна быть установлена ровно одна кука")

	cookie := cookies[0]

	assert.Equal(t, cookieName, cookie.Name)
	assert.Equal(t, cookieValue, cookie.Value)

	// Проверяем время истечения.
	// Используем .UTC(), так как в HTTP куках время всегда передается в формате UTC
	assert.True(t, expiresAt.Equal(cookie.Expires.UTC()), "Время истечения должно совпадать")

	// Проверяем параметры безопасности
	assert.True(t, cookie.HttpOnly, "Кука должна иметь флаг HttpOnly")
	assert.Equal(t, "/", cookie.Path, "Путь должен быть корневым (/)")
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite, "SameSite должен быть Lax")
}
