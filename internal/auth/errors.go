package auth

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("email is already registered")
	ErrInvalidPassword   = errors.New("invalid email or password")
	ErrWrongProvider     = errors.New("account uses another login provider")
	ErrAccountInactive   = errors.New("account is inactive")
	ErrAccountPending    = errors.New("account is awaiting admin approval")
	ErrOTPDeliveryFailed = errors.New("OTP email delivery failed")
)
