package server

import (
	"bytes"
	"calculator/pkg/calculator"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
)

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

func CalculatorHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		result, err := calculator.Calc(data.Expression, logger)
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
}
func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
}
func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
func LoggingMiddleware(next http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqBody bytes.Buffer
		_, err := io.Copy(&reqBody, r.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to read request body: %e", err))
			http.Error(w, "failed to process request", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(&reqBody)
		logger.Info(
			fmt.Sprintf(
				"HTTP Method: %s, Request Body: %s",
				r.Method,
				strings.ReplaceAll(
					strings.ReplaceAll(strings.ReplaceAll(reqBody.String(), "\r", ""), "\n", ""),
					"\"",
					"",
				),
			),
		)
		rw := &ResponseWriterWrapper{
			ResponseWriter: w,
			Body:           &bytes.Buffer{},
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(rw, r)
		logger.Info(
			fmt.Sprintf(
				"HTTP Response Code: %d, Response Body: %s",
				rw.StatusCode,
				strings.ReplaceAll(
					strings.ReplaceAll(strings.ReplaceAll(rw.Body.String(), "\r", ""), "\n", ""),
					"\"",
					"",
				),
			),
		)
	}
}
func ErrorRecoveryMiddleware(next http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(fmt.Sprintf("panic recovered: %v", rec))
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorData{Error: "internal server error"})
			}
		}()
		next(w, r)
	}
}
