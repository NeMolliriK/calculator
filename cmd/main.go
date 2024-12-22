package main

import (
	"calculator/internal/application"
	"log/slog"
	"os"
)

func main() {
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	app := application.New(logger)
	if err := app.Run(); err != nil {
		logger.Error("Application failed to run", "error", err)
	}
}
