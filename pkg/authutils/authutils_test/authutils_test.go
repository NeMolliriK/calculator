package authutils_test

import (
	"testing"
	"time"

	"calculator/pkg/authutils"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "myS3cret!"
	hash, err := authutils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Error("HashPassword should not return the plain password")
	}
	if !authutils.CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash returned false for correct password")
	}
	if authutils.CheckPasswordHash("wrongpass", hash) {
		t.Error("CheckPasswordHash returned true for incorrect password")
	}
}

func TestGenerateJWT(t *testing.T) {
	userID := uint(42)
	secret := "test-secret"
	tokenStr, err := authutils.GenerateJWT(userID, secret)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			t.Errorf("unexpected signing method: %v", token.Method)
		}
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("Parse token error: %v", err)
	}
	if !token.Valid {
		t.Fatal("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("invalid claims type: %T", token.Claims)
	}

	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		t.Fatalf("user_id claim is not float64: %T", claims["user_id"])
	}
	if uint(uidFloat) != userID {
		t.Errorf("user_id claim = %v, want %v", uidFloat, userID)
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp claim is not float64: %T", claims["exp"])
	}
	expTime := time.Unix(int64(expFloat), 0)
	if time.Now().After(expTime) {
		t.Errorf("exp time %v is in the past", expTime)
	}
	if expTime.After(time.Now().Add(25 * time.Hour)) {
		t.Errorf("exp time %v is too far in the future", expTime)
	}
}

func TestGenerateJWT_InvalidSignature(t *testing.T) {
	userID := uint(1)
	secret := "secret1"
	tokenStr, err := authutils.GenerateJWT(userID, secret)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}

	_, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("other-secret"), nil
	})
	if err == nil {
		t.Error("Parse did not return error for invalid signature")
	}
}
