package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/chai2010/webp"
)

func ConvertToWebp(src io.Reader) (*bytes.Buffer, error) {
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (unsupported format?): %w", err)
	}
	_ = format // это я сделал, чтобы не забыть, что возвращает Decode вторым аргументом

	options := &webp.Options{Lossless: false, Quality: 80}

	buf := new(bytes.Buffer)
	if err := webp.Encode(buf, img, options); err != nil {
		return nil, fmt.Errorf("failed to encode to webp: %w", err)
	}

	return buf, nil
}
