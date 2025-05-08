package database

import (
	"calculator/pkg/loggers"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

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
	logger.Info("database initialized")
}
