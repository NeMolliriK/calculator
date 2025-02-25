package handler

import (
	"bytes"
	"calculator/pkg/calculator"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

type Decorator func(http.Handler) http.Handler
type RequestData struct {
	Expression string `json:"expression"`
}
type ResponseData struct {
	Result string `json:"result"`
}
type ErrorData struct {
	Error string `json:"error"`
}
type ResponseWriterWrapper struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

func New(ctx context.Context) (http.Handler, error) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/api/v1/calculate", calculatorHandler)
	return serveMux, nil
}
func Decorate(next http.Handler, ds ...Decorator) http.Handler {
	decorated := next
	for d := len(ds) - 1; d >= 0; d-- {
		decorated = ds[d](decorated)
	}
	return decorated
}
func calculatorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorData{Error: "only POST method is allowed"})
		return
	}
	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorData{Error: "invalid JSON"})
		return
	}
	if data.Expression == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: "no expression provided"})
		return
	}
	err = calculator.ValidateExpression(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: err.Error()})
		return
	}
	result, err := calculator.Calc(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).
			Encode(ErrorData{Error: "there's an unknown error in the expression"})
		return
	}
	if math.IsInf(result, 0) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: "You can't divide by zero!"})
		return
	}
	json.NewEncoder(w).Encode(ResponseData{Result: strconv.FormatFloat(result, 'f', -1, 64)})
}

//func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
//	rw.Body.Write(b)
//	return rw.ResponseWriter.Write(b)
//}
//func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
//	rw.StatusCode = statusCode
//	rw.ResponseWriter.WriteHeader(statusCode)
//}

//func (a *Application) Run(port string) error {
//	err := godotenv.Load()
//	if err != nil {
//		fmt.Println("Warning: .env file not found, falling back to system environment variables")
//	}
//	loggers := setupLogger()
//	mux := http.NewServeMux()
//	mux.Handle(
//		"/api/v1/calculate",
//		server.LoggingMiddleware(
//			server.ErrorRecoveryMiddleware(server.CalculatorHandler(a.loggers), a.loggers),
//			a.loggers,
//		),
//	)
//	fmt.Println("The only endpoint is available at http://localhost:8080/api/v1/calculate")
//	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
//	if err != nil {
//		a.loggers.Error("Failed to start the server", "error", err)
//	}
//	return err
//}
