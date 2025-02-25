package application

import (
	"calculator/http/server"
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
	shutDownFunc, err := server.Run(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-c
	cancel()
	err = shutDownFunc(ctx)
	if err != nil {
		logger.Error(err.Error())
		return 1
	}
	return 0
}
