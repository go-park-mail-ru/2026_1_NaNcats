package repository

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrPromocodeNotFound(t *testing.T) {
	t.Run("Сингл-инстанс, сравним через ==", func(t *testing.T) {
		assert.Equal(t, ErrPromocodeNotFound, ErrPromocodeNotFound)
		assert.True(t, stderrors.Is(ErrPromocodeNotFound, ErrPromocodeNotFound))
	})

	t.Run("Сообщение фиксированное", func(t *testing.T) {
		assert.Equal(t, "promocode not found", ErrPromocodeNotFound.Error())
	})
}

func TestRepositoryErrors(t *testing.T) {
	assert.Equal(t, "order state has changed or order not found", ErrStateChanged.Error())
	assert.Equal(t, "split not found by payment ID", ErrSplitNotFound.Error())
}
