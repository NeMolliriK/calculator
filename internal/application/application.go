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
func (a *Application) Run(port string) error {
	mux := http.NewServeMux()
	mux.Handle(
		"/api/v1/calculate",
		server.LoggingMiddleware(
			server.ErrorRecoveryMiddleware(server.CalculatorHandler(a.logger), a.logger),
			a.logger,
		),
	)
	fmt.Println("The only endpoint is available at http://localhost:8080/api/v1/calculate")
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil {
		a.logger.Error("Failed to start the server", "error", err)
	}
	return err
}
