package uploadimage

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const MaxSize = 5 * 1024 * 1024

var (
	ErrTooLarge    = errors.New("image exceeds maximum size")
	ErrUnsupported = errors.New("unsupported image content or extension")
)

var allowedTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func Validate(file *multipart.FileHeader) (string, error) {
	if file.Size > MaxSize {
		return "", ErrTooLarge
	}
	opened, err := file.Open()
	if err != nil {
		return "", err
	}
	defer opened.Close()
	return validateContent(file.Filename, file.Size, opened)
}

func validateContent(filename string, size int64, content io.Reader) (string, error) {
	if size > MaxSize {
		return "", ErrTooLarge
	}
	extension := strings.ToLower(filepath.Ext(filename))
	expectedType, ok := allowedTypes[extension]
	if !ok {
		return "", ErrUnsupported
	}

	var header [512]byte
	read, err := io.ReadFull(content, header[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	actualType, _, err := mime.ParseMediaType(http.DetectContentType(header[:read]))
	if err != nil || actualType != expectedType {
		return "", ErrUnsupported
	}
	return extension, nil
}
