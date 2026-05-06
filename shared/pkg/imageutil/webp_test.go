package imageutil

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyPNG(t *testing.T) io.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	require.NoError(t, err, "failed to encode dummy png")

	return buf
}

func TestConvertToWebp(t *testing.T) {
	tests := []struct {
		name      string
		inputPrep func(t *testing.T) io.Reader
		expectErr bool
	}{
		{
			name: "Успешная конвертация PNG в WebP",
			inputPrep: func(t *testing.T) io.Reader {
				return createDummyPNG(t)
			},
			expectErr: false,
		},
		{
			name: "Ошибка: переданы невалидные данные (не картинка)",
			inputPrep: func(t *testing.T) io.Reader {
				return bytes.NewBufferString("this is definitely not an image")
			},
			expectErr: true,
		},
		{
			name: "Ошибка: передан пустой Reader",
			inputPrep: func(t *testing.T) io.Reader {
				return bytes.NewBuffer(nil)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := tt.inputPrep(t)

			result, err := ConvertToWebp(reader)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "failed to decode image")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Greater(t, result.Len(), 0, "Ожидается, что результирующий буфер не будет пустым")
			}
		})
	}
}
