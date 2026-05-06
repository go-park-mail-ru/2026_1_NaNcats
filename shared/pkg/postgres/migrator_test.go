package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		expectErr   bool
	}{
		{
			name:        "Ошибка: неверный URL базы данных",
			databaseURL: "invalid-url",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunMigrations(tt.databaseURL)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
