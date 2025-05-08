package main

import (
	"calculator/internal/application"
	"calculator/internal/database"
	"calculator/pkg/loggers"
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	os.Exit(mainWithExitCode(ctx))
}
func mainWithExitCode(ctx context.Context) int {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, falling back to system environment variables")
	}
	loggers.InitLogger("server", "server_logs.txt")
	loggers.InitLogger("orchestrator", "calculations_logs.txt")
	loggers.InitLogger("general", "general_logs.txt")
	defer loggers.CloseAllLoggers()
	database.Init()
	app := application.New()
	return app.Run(ctx)
}
