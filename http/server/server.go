package server

import (
	"bytes"
	"calculator/http/server/handler"
	"calculator/pkg/loggers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

func (rw *ResponseWriterWrapper) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func new(ctx context.Context) (http.Handler, error) {
	muxHandler, err := handler.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("handler initialization error: %w", err)
	}
	muxHandler = handler.Decorate(muxHandler, errorRecoveryMiddleware(), loggingMiddleware())
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

func loggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger := loggers.GetLogger("server")
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			cleanBody := strings.ReplaceAll(string(bodyBytes), "\r\n", "")
			logger.Info(
				"HTTP request",
				"method",
				r.Method,
				"path",
				r.URL.Path,
				"body",
				cleanBody,
			)
			rw := &ResponseWriterWrapper{
				ResponseWriter: w,
				Body:           &bytes.Buffer{},
				StatusCode:     http.StatusOK,
			}
			next.ServeHTTP(rw, r)
			duration := time.Since(start)
			cleanBody = strings.ReplaceAll(rw.Body.String(), "\r\n", "")
			logger.Info(
				"HTTP response",
				"status",
				rw.StatusCode,
				"body",
				cleanBody,
				"duration",
				duration,
			)
		})
	}
}

func errorRecoveryMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger := loggers.GetLogger("server")
					logger.Error("panic recovered", "error", rec)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(handler.ErrorData{Error: "internal server error"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
