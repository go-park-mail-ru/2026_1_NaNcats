package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermanentError(t *testing.T) {
	t.Run("С обёрнутой ошибкой", func(t *testing.T) {
		cause := stderrors.New("disk full")
		e := NewPermanentError(cause)

		assert.False(t, e.IsRetryable())
		assert.Equal(t, cause, e.Unwrap())
		assert.Equal(t, "permanent error: disk full", e.Error())
		assert.True(t, stderrors.Is(e, cause))
	})

	t.Run("Без обёрнутой ошибки", func(t *testing.T) {
		e := &PermanentError{}
		assert.Equal(t, "permanent error", e.Error())
		assert.Nil(t, e.Unwrap())
	})

	t.Run("С текстом сообщения", func(t *testing.T) {
		e := NewPermanentErrorWithMessage("bad payload")
		assert.False(t, e.IsRetryable())
		assert.Equal(t, "permanent error: bad payload", e.Error())
		assert.Equal(t, "bad payload", e.Unwrap().Error())
	})
}

func TestTemporaryError(t *testing.T) {
	t.Run("С обёрнутой ошибкой", func(t *testing.T) {
		cause := stderrors.New("timeout")
		e := NewTemporaryError(cause)

		assert.True(t, e.IsRetryable())
		assert.Equal(t, cause, e.Unwrap())
		assert.Equal(t, "temporary error: timeout", e.Error())
	})

	t.Run("Без обёрнутой ошибки", func(t *testing.T) {
		e := &TemporaryError{}
		assert.Equal(t, "temporary error", e.Error())
		assert.Nil(t, e.Unwrap())
	})

	t.Run("С текстом сообщения", func(t *testing.T) {
		e := NewTemporaryErrorWithMessage("network issue")
		assert.True(t, e.IsRetryable())
		assert.Equal(t, "temporary error: network issue", e.Error())
		assert.Equal(t, "network issue", e.Unwrap().Error())
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    bool
		comment string
	}{
		{
			name:    "nil — false",
			err:     nil,
			want:    false,
			comment: "nil не ретраится",
		},
		{
			name:    "PermanentError — false",
			err:     NewPermanentErrorWithMessage("permanent"),
			want:    false,
			comment: "permanent явно не ретраить",
		},
		{
			name:    "TemporaryError — true",
			err:     NewTemporaryErrorWithMessage("temp"),
			want:    true,
			comment: "temporary явно ретраить",
		},
		{
			name:    "Обычная ошибка — true (по умолчанию ретраим)",
			err:     stderrors.New("oops"),
			want:    true,
			comment: "не классифицированную ошибку считаем временной",
		},
		{
			name:    "Permanent внутри Wrap-цепочки — false",
			err:     stderrors.Join(stderrors.New("outer"), NewPermanentErrorWithMessage("inner")),
			want:    false,
			comment: "errors.As должен дотянуться через Join",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsRetryable(tt.err), tt.comment)
		})
	}
}
