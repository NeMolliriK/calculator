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
	DB, err = gorm.Open(sqlite.Open("sqlite.db?_pragma=foreign_keys(ON)"), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to the database:", err)
		panic(err)
	}
	if err := DB.AutoMigrate(&Expression{}, &User{}); err != nil {
		logger.Error("migration error:", err)
		panic(err)
	}
	var expressions []Expression
	err = DB.Find(&expressions).Error
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
}
