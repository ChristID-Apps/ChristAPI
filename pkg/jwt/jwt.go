package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret []byte

func init() {
	s := os.Getenv("JWT_SECRET")
	if s != "" {
		secret = []byte(s)
	} else {
		secret = []byte("secret-key")
	}
}

func Secret() []byte {
	return secret
}

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}
