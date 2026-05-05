package errutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestDomainError_Constructors(t *testing.T) {
	causeErr := errors.New("db connection failed")

	tests := []struct {
		name         string
		errCreator   func() domainError
		expectedMsg  string
		expectedSlug string
		expectedCode codes.Code
	}{
		{
			name: "Создание базовой ошибки (New)",
			errCreator: func() domainError {
				return New("NOT_FOUND", "user not found", codes.NotFound)
			},
			expectedMsg:  "[NOT_FOUND] user not found",
			expectedSlug: "NOT_FOUND",
			expectedCode: codes.NotFound,
		},
		{
			name: "Создание ошибки с оберткой (Wrap)",
			errCreator: func() domainError {
				return Wrap("DB_ERROR", "failed to query user", causeErr, codes.Internal)
			},
			expectedMsg:  "[DB_ERROR] failed to query user: db connection failed",
			expectedSlug: "DB_ERROR",
			expectedCode: codes.Internal,
		},
		{
			name: "Быстрое создание ошибки (Message)",
			errCreator: func() domainError {
				return Message("BAD_REQ", "invalid payload")
			},
			expectedMsg:  "[BAD_REQ] invalid payload",
			expectedSlug: "BAD_REQ",
			expectedCode: codes.Internal, // По умолчанию в Message зашит codes.Internal
		},
		{
			name: "Создание внутренней ошибки (Internal)",
			errCreator: func() domainError {
				return Internal("something went wrong", causeErr)
			},
			expectedMsg:  "[INTERNAL_SERVER_ERROR] something went wrong: db connection failed",
			expectedSlug: InternalSlug,
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.errCreator()

			// Assert
			require.NotNil(t, err)
			assert.Equal(t, tt.expectedMsg, err.Error())
			assert.Equal(t, tt.expectedSlug, err.Slug())
			assert.Equal(t, tt.expectedCode, err.GRPCStatus())
		})
	}
}
