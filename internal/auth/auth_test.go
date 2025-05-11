package auth_test

import (
	"os"
	"testing"

	"calculator/internal/auth"
	"calculator/internal/database"
	"calculator/pkg/authutils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := database.DB.AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("failed to migrate User: %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	setupTestDB(t)
	login := "alice"
	password := "password123"
	if err := auth.Register(login, password); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	user, err := database.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("GetUserByLogin error: %v", err)
	}
	if user.Login != login {
		t.Errorf("user.Login = %q, want %q", user.Login, login)
	}
	if !authutils.CheckPasswordHash(password, user.GetPassword()) {
		t.Error("password hash does not match original password")
	}
}

func TestRegister_AlreadyExists(t *testing.T) {
	setupTestDB(t)
	login := "bob"
	password := "pass"
	if err := auth.Register(login, password); err != nil {
		t.Fatalf("first Register returned error: %v", err)
	}
	err := auth.Register(login, password)
	if err == nil {
		t.Fatal("expected error on registering existing user, got nil")
	}
	if err.Error() != "user already exists" {
		t.Errorf("error = %q, want %q", err.Error(), "user already exists")
	}
}

func TestLogin_Success(t *testing.T) {
	setupTestDB(t)
	login := "carol"
	password := "secret"
	hash, err := authutils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	user := database.User{Login: login}
	user.SetPassword(hash)
	if err := database.CreateUser(&user); err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	secret := "jwtsecret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")
	tokenStr, err := auth.Login(login, password)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("Parse JWT error: %v", err)
	}
	if !token.Valid {
		t.Fatal("token is invalid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims type: %T", token.Claims)
	}
	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		t.Fatalf("user_id claim not float64")
	}
	if uint(uidFloat) != user.ID {
		t.Errorf("user_id claim = %v, want %v", uidFloat, user.ID)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDB(t)
	login := "dave"
	password := "secret"
	hash, _ := authutils.HashPassword(password)
	user := database.User{Login: login}
	user.SetPassword(hash)
	database.CreateUser(&user)
	os.Setenv("JWT_SECRET", "jwtsecret")
	defer os.Unsetenv("JWT_SECRET")
	_, err := auth.Login(login, "wrong")
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("error = %q, want %q", err.Error(), "invalid credentials")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	setupTestDB(t)
	os.Setenv("JWT_SECRET", "jwtsecret")
	defer os.Unsetenv("JWT_SECRET")
	_, err := auth.Login("nonexistent", "pass")
	if err == nil {
		t.Fatal("expected error for invalid credentials, got nil")
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("error = %q, want %q", err.Error(), "invalid credentials")
	}
}
