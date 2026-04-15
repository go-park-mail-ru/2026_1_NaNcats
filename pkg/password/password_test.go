package password

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		params      *Params
		expectedErr error
	}{
		{
			name:        "Успешное хэширование с дефолтными параметрами",
			password:    "correct_password",
			params:      nil,
			expectedErr: nil,
		},
		{
			name:     "Успешное хэширование с кастомными параметрами",
			password: "another_correct_password",
			params: &Params{
				Memory:      16 * 1024,
				Iterations:  1,
				Parallelism: 1,
				SaltLength:  16,
				KeyLength:   32,
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка, пароль слишком короткий",
			password:    "short",
			params:      nil,
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "Ошикба: пароль слишком длинный",
			password:    strings.Repeat("a", MaxPasswordLength+1),
			params:      nil,
			expectedErr: ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password, tt.params)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	correctPassword := "ochen_secured_parol_123"
	hash, err := HashPassword(correctPassword, nil)
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		encodedHash string
		expectOk    bool
		expectErr   error
	}{
		{
			name:        "Успешная проверка",
			password:    correctPassword,
			encodedHash: hash,
			expectOk:    true,
			expectErr:   nil,
		},
		{
			name:        "Ошибка: неверный пароль",
			password:    "wrong_password",
			encodedHash: hash,
			expectOk:    false,
			expectErr:   ErrWrongPassword,
		},
		{
			name:        "Ошибка: пароль слишком длинный",
			password:    strings.Repeat("a", MaxPasswordLength+1),
			encodedHash: hash,
			expectOk:    false,
			expectErr:   ErrPasswordTooLong,
		},
		{
			name:        "Ошибка: неверный формат хеша (мало частей)",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$m=65536,t=3,p=2$badhash",
			expectOk:    false,
			expectErr:   ErrInvalidHash,
		},
		{
			name:        "Ошибка: несовместимая версия",
			password:    correctPassword,
			encodedHash: "$argon2id$v=18$m=65536,t=3,p=2$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2g",
			expectOk:    false,
			expectErr:   ErrIncompatibleVersion,
		},
		{
			name:        "Ошибка: битый base64 в соли",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$m=65536,t=3,p=2$!@#$%^?&*()$aGFzaGhhc2hoYXNoaGFzaGhhc2g",
			expectOk:    false,
			expectErr:   nil,
		},
		{
			name:        "Ошибка: битый формат параметров (m=, t=, p=)",
			password:    correctPassword,
			encodedHash: "$argon2id$v=19$not_params$c2FsdA$aGFzaA",
			expectOk:    false,
			expectErr:   nil,
		},
		{
			name:        "Ошибка: битая версия (не число)",
			password:    correctPassword,
			encodedHash: "$argon2id$v=not_int$m=65536,t=3,p=2$c2FsdA$aGFzaA",
			expectOk:    false,
			expectErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifyPassword(tt.password, tt.encodedHash)

			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
				assert.False(t, ok)
			} else if tt.expectOk {
				assert.NoError(t, err)
				assert.True(t, ok)
			} else {
				assert.Error(t, err)
				assert.False(t, ok)
			}
		})
	}
}
