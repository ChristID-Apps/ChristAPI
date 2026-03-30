package jwt

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var secret = []byte("secret-key")

func GenerateToken(userID int) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    return token.SignedString(secret)
}