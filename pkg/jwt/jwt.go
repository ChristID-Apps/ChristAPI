package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret []byte

func init() {
	s := os.Getenv("JWT_SECRET")
	if s != "" {
		secret = []byte(s)
	}
}

func Secret() []byte {
	return secret
}

func GenerateToken(userID int) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("JWT_SECRET is not configured")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}
