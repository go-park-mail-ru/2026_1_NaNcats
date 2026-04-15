package csrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	t.Run("Успешная генерация уникальных токенов", func(t *testing.T) {
		token1, err := GenerateToken()
		assert.NoError(t, err)
		assert.Len(t, token1, 64)

		token2, _ := GenerateToken()
		assert.NotEqual(t, token1, token2)
	})
}
