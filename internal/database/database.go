package database

import (
	"calculator/pkg/loggers"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Users *gorm.DB
	DB    *gorm.DB
)

func Init() {
	var err error
	logger := loggers.GetLogger("general")
	DB, err = gorm.Open(sqlite.Open("expressions.db?_pragma=foreign_keys(ON)"), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to the database:", err)
		panic(err)
	}
	if err := DB.AutoMigrate(&Expression{}); err != nil {
		logger.Error("migration error:", err)
		panic(err)
	}
	expressions, err := GetAllExpressions()
	if err != nil {
		logger.Error("failed to get all expressions", err)
		panic(err)
	}
	for _, expression := range expressions {
		if expression.Status == "processing" {
			err = UpdateExpressionStatus(expression.ID, "failed due to a server error")
			if err != nil {
				logger.Error("failed to change status", err)
				panic(err)
			}
		}
	}
	logger.Info("database initialized")
	Users, err = gorm.Open(sqlite.Open("users.db?_pragma=foreign_keys(ON)"), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to the database:", err)
		panic(err)
	}
	if err := Users.AutoMigrate(&User{}); err != nil {
		logger.Error("migration error:", err)
		panic(err)
	}
	logger.Info("database users initialized")
}
