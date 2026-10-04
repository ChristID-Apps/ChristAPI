package uploadimage

import (
	"bytes"
	"errors"
	"testing"
)

func TestValidateContent(t *testing.T) {
	pngHeader := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	jpegHeader := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	webpHeader := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' '}
	tests := []struct {
		name      string
		filename  string
		size      int64
		content   []byte
		wantExt   string
		wantError error
	}{
		{name: "matching PNG", filename: "image.PNG", content: pngHeader, wantExt: ".png"},
		{name: "matching JPEG", filename: "image.jpeg", content: jpegHeader, wantExt: ".jpeg"},
		{name: "matching WebP", filename: "image.webp", content: webpHeader, wantExt: ".webp"},
		{name: "mismatched content", filename: "image.jpg", content: pngHeader, wantError: ErrUnsupported},
		{name: "unsupported extension", filename: "image.svg", content: pngHeader, wantError: ErrUnsupported},
		{name: "too large", filename: "image.png", size: MaxSize + 1, content: pngHeader, wantError: ErrTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateContent(test.filename, test.size, bytes.NewReader(test.content))
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
			if got != test.wantExt {
				t.Fatalf("extension = %q, want %q", got, test.wantExt)
			}
		})
	}
}
