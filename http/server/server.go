package server

import (
	"calculator/http/server/handler"
	"calculator/pkg/loggers"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type ErrorData struct {
	Error string `json:"error"`
}

func new(ctx context.Context) (http.Handler, error) {
	muxHandler, err := handler.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("handler initialization error: %w", err)
	}
	muxHandler = handler.Decorate(muxHandler, loggingMiddleware())
	return muxHandler, nil
}
func Run(ctx context.Context) (func(context.Context) error, error) {
	muxHandler, err := new(ctx)
	if err != nil {
		return nil, err
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: muxHandler}
	logger := loggers.GetLogger("server")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("ListenAndServe", slog.String("err", err.Error()))
		}
	}()
	fmt.Printf("The only endpoint is available at http://localhost:%s/api/v1/calculate", port)
	return srv.Shutdown, nil
}
func loggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger := loggers.GetLogger("server")
			logger.Info("HTTP request", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Duration("duration", duration))
		})
	}
}
