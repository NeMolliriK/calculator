package loggers

import (
	"log/slog"
	"os"
	"sync"
)

var (
	loggers  = make(map[string]*slog.Logger)
	logFiles = make(map[string]*os.File)
	mu       sync.Mutex
)

func InitLogger(name, file string) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := loggers[name]; exists {
		return
	}
	logLevel := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logFile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: level}))
	loggers[name] = logger
	logFiles[name] = logFile
}
func GetLogger(name string) *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	logger, exists := loggers[name]
	if !exists {
		panic("Logger not initialized: " + name)
	}
	return logger
}
func CloseAllLoggers() {
	mu.Lock()
	defer mu.Unlock()
	for name, file := range logFiles {
		file.Close()
		delete(logFiles, name)
		delete(loggers, name)
	}
}
