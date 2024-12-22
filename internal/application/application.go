package application

import (
	"calculator/pkg/server"
	"fmt"
	"log/slog"
	"net/http"
)

type Application struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Application {
	return &Application{
		logger: logger,
	}
}
func (a *Application) Run() error {
	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/calculate",
		server.LoggingMiddleware(server.CalculatorHandler(a.logger), a.logger),
	)
	fmt.Println("The only endpoint is available at http://localhost:8080/api/v1/calculate")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		a.logger.Error("Failed to start the server", "error", err)
	}
	return err
}
