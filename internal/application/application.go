package application

import (
	"calculator/pkg/server"
	"fmt"
	"net/http"
)

type Application struct {
}

func New() *Application {
	return &Application{}
}
func (a *Application) Run() error {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/calculate", http.HandlerFunc(
		server.CalculatorHandler))
	fmt.Println("The only endpoint is available at " +
		"http://localhost:8080/api/v1/calculate")
	return http.ListenAndServe(":8080", mux)
}
