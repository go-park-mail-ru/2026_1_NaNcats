package grpcutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockStatusCoder struct {
	msg  string
	code codes.Code
}

func (m mockStatusCoder) Error() string          { return m.msg }
func (m mockStatusCoder) GRPCStatus() codes.Code { return m.code }

type mockStatusSlugger struct {
	msg  string
	slug string
	code codes.Code
}

func (m mockStatusSlugger) Error() string          { return m.msg }
func (m mockStatusSlugger) GRPCStatus() codes.Code { return m.code }
func (m mockStatusSlugger) Slug() string           { return m.slug }

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		name         string
		inputErr     error
		isNil        bool
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "Успех: nil ошибка",
			inputErr:     nil,
			isNil:        true,
			expectedCode: codes.OK,
		},
		{
			name:         "Ошибка уже является gRPC статусом",
			inputErr:     status.Error(codes.InvalidArgument, "already grpc error"),
			isNil:        false,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "already grpc error",
		},
		{
			name: "Ошибка реализует StatusCoder и Slugger",
			inputErr: mockStatusSlugger{
				msg:  "full error message text",
				slug: "USER_NOT_FOUND",
				code: codes.NotFound,
			},
			isNil:        false,
			expectedCode: codes.NotFound,
			expectedMsg:  "USER_NOT_FOUND",
		},
		{
			name: "Ошибка реализует только StatusCoder",
			inputErr: mockStatusCoder{
				msg:  "just a status error",
				code: codes.PermissionDenied,
			},
			isNil:        false,
			expectedCode: codes.PermissionDenied,
			expectedMsg:  "just a status error",
		},
		{
			name:         "Обычная Go ошибка",
			inputErr:     errors.New("standard go error"),
			isNil:        false,
			expectedCode: codes.Internal,
			expectedMsg:  "standard go error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resErr := ToGRPCError(tt.inputErr)

			if tt.isNil {
				assert.NoError(t, resErr)
				return
			}

			require.Error(t, resErr)

			st, ok := status.FromError(resErr)
			require.True(t, ok, "Ожидалась ошибка формата gRPC-статус")

			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}
