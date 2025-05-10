package authutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateJWT(userID uint, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateJWT(userID uint, secret string) (string, error) {
	return generateJWT(userID, secret)
}
