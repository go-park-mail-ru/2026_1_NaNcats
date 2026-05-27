package errors

import (
	"errors"
	"fmt"
)

// Интерфейс для ошибок, которые поддерживают классификацию на возможность повторной обработки
type Retryable interface {
	error
	IsRetryable() bool
}

// Неустранимая (фатальная) ошибка
// Сообщения с ней не должны возвращаться в очередь и отправляются в DLQ
type PermanentError struct {
	cause error
}

// Оборачивает ошибку как фатальную
func NewPermanentError(err error) *PermanentError {
	return &PermanentError{cause: err}
}

// Новая фатальная ошибка с текстовым сообщением
func NewPermanentErrorWithMessage(msg string) *PermanentError {
	return &PermanentError{cause: errors.New(msg)}
}

func (e *PermanentError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("permanent error: %v", e.cause)
	}
	return "permanent error"
}

func (e *PermanentError) Unwrap() error {
	return e.cause
}

func (e *PermanentError) IsRetryable() bool {
	return false
}

// Временная ошибка, сообщения возвращаются в очередь
type TemporaryError struct {
	cause error
}

// Оборачивает ошибку как временную
func NewTemporaryError(err error) *TemporaryError {
	return &TemporaryError{cause: err}
}

// Новая временная ошибка с текстовым сообщением
func NewTemporaryErrorWithMessage(msg string) *TemporaryError {
	return &TemporaryError{cause: errors.New(msg)}
}

func (e *TemporaryError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("temporary error: %v", e.cause)
	}
	return "temporary error"
}

func (e *TemporaryError) Unwrap() error {
	return e.cause
}

func (e *TemporaryError) IsRetryable() bool {
	return true
}

// IsRetryable проверяет, можно ли повторить попытку обработки сообщения
// errors.As проверяет всю цепочку обернутых ошибок
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var retryableErr Retryable
	if errors.As(err, &retryableErr) {
		return retryableErr.IsRetryable()
	}

	return true
}
