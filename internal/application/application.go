package application

import (
	"calculator/internal/http/server"
	rpcserver "calculator/internal/rpc"
	"calculator/pkg/loggers"
	"context"
	"os"
	"os/signal"
)

type Application struct{}

func New() *Application {
	return &Application{}
}
func (a *Application) Run(ctx context.Context) int {
	logger := loggers.GetLogger("general")
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	httpShutdown, err := server.Run(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	grpcShutdown, err := rpcserver.Run(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	<-ctx.Done()
	if err := grpcShutdown(context.Background()); err != nil {
		logger.Error("gRPC shutdown: " + err.Error())
	}
	if err := httpShutdown(context.Background()); err != nil {
		logger.Error("HTTP shutdown: " + err.Error())
	}
	return 0
}
