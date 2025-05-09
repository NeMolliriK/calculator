package server

import (
	"calculator/http/server/handler"
	"calculator/http/server/middleware"
	"calculator/pkg/loggers"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func new(ctx context.Context) (http.Handler, error) {
	muxHandler, err := handler.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("handler initialization error: %w", err)
	}
	muxHandler = handler.Decorate(
		muxHandler,
		middleware.ErrorRecoveryMiddleware(),
		middleware.LoggingMiddleware(),
	)
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
	fmt.Printf("The server is running at http://localhost:%s/", port)
	return srv.Shutdown, nil
}
