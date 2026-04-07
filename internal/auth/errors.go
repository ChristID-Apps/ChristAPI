package auth

import "errors"

var (
    ErrUserNotFound    = errors.New("USER_NOT_FOUND")
    ErrInvalidPassword = errors.New("INVALID_PASSWORD")
)