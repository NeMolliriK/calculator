package loggers_test

import (
	"calculator/pkg/loggers"
	"os"
	"testing"
)

func TestLoggerLifecycle(t *testing.T) {
	loggers.InitLogger("test", os.DevNull)
	_ = loggers.GetLogger("test")
	loggers.CloseAllLoggers()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic after logger closed, but did not panic")
		}
	}()
	_ = loggers.GetLogger("test")
}
