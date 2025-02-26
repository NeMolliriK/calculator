package handler

import (
	"bytes"
	"calculator/pkg/calculator"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"os"
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
	serveMux.HandleFunc("/api/v1/calculate", calculatorAPIHandler)
	serveMux.HandleFunc("/calculate", calculatorHandler)
	return serveMux, nil
}
func Decorate(next http.Handler, ds ...Decorator) http.Handler {
	decorated := next
	for d := len(ds) - 1; d >= 0; d-- {
		decorated = ds[d](decorated)
	}
	return decorated
}
func calculatorAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	result, err := calculator.Calc(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorData{Error: err.Error()})
		return
	}
	if math.IsInf(result, 0) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: "You can't divide by zero!"})
		return
	}
	json.NewEncoder(w).Encode(ResponseData{Result: strconv.FormatFloat(result, 'f', -1, 64)})
}
func calculatorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent, err := os.ReadFile("templates/calculator.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Ошибка: не удалось загрузить шаблон"))
			return
		}
		w.Write(htmlContent)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Method not allowed"))
}
