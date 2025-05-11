package auth

import (
	"calculator/internal/database"
	"calculator/pkg/authutils"
	"errors"
	"os"
)

func Register(login, password string) error {
	if _, err := database.GetUserByLogin(login); err == nil {
		return errors.New("user already exists")
	}
	hash, err := authutils.HashPassword(password)
	if err != nil {
		return err
	}
	user := database.User{Login: login}
	user.SetPassword(hash)
	return database.CreateUser(&user)
}

func Login(login, password string) (string, error) {
	user, err := database.GetUserByLogin(login)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if !authutils.CheckPasswordHash(password, user.GetPassword()) {
		return "", errors.New("invalid credentials")
	}
	return authutils.GenerateJWT(user.ID, os.Getenv("JWT_SECRET"))
}
