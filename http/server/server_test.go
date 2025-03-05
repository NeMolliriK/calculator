package server_test

import (
	"calculator/http/server"
	"calculator/pkg/loggers"
	"context"
	"os"
	"testing"
	"time"
)

func TestServerRunAndShutdown(t *testing.T) {
	os.Setenv("PORT", "9090")
	loggers.InitLogger("server", os.DevNull)
	loggers.InitLogger("orchestrator", os.DevNull)
	loggers.InitLogger("general", os.DevNull)
	ctx := context.Background()
	shutdown, err := server.Run(ctx)
	if err != nil {
		t.Fatalf("server.Run returned error: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	err = shutdown(ctx)
	if err != nil {
		t.Errorf("shutdown returned error: %v", err)
	}
	loggers.CloseAllLoggers()
}
