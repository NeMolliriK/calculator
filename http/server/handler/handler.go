package handler

import (
	"bytes"
	"calculator/pkg/calculator"
	"calculator/pkg/structures"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync"
)

type Decorator func(http.Handler) http.Handler

type RequestData struct {
	Expression string `json:"expression"`
}

type ResponseData struct {
	Result float64 `json:"result"`
}

type ErrorData struct {
	Error string `json:"error"`
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

type IDResponse struct {
	ID string `json:"id"`
}

type ExpressionResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ExpressionsResponse struct {
	Expressions []ExpressionResponse `json:"expressions"`
}

var (
	expressionsMap sync.Map
)

func New(ctx context.Context) (http.Handler, error) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/api/v1/calculate", calculatorAPIHandler)
	//serveMux.HandleFunc("/calculate", calculatorHandler)
	serveMux.HandleFunc("/api/v1/expressions", expressionsHandler)
	serveMux.HandleFunc("/api/v1/expressions/", expressionHandler)
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
	expressionID := uuid.New().String()
	expression := structures.Expression{ID: expressionID, Data: data.Expression, Status: "pending", Result: 0}
	expressionsMap.Store(expressionID, &expression)
	w.WriteHeader(http.StatusCreated)
	go calculator.Calc(&expression)
	json.NewEncoder(w).Encode(IDResponse{expressionID})
}

func expressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorData{Error: "only GET method is allowed"})
		return
	}
	expressions := ExpressionsResponse{}
	expressionsMap.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(*structures.Expression)
		expressions.Expressions = append(expressions.Expressions, ExpressionResponse{k, v.Status, v.Result})
		return true
	})
	json.NewEncoder(w).Encode(expressions)
}

func expressionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorData{Error: "only GET method is allowed"})
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorData{Error: "ID not provided"})
		return
	}
	value, ok := expressionsMap.Load(parts[4])
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorData{Error: "there is no such expression"})
		return
	}
	expression := value.(*structures.Expression)
	json.NewEncoder(w).Encode(ExpressionResponse{parts[4], expression.Status, expression.Result})
}
