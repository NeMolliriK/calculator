package server

import (
	"bytes"
	"calculator/http/server/handler"
	"calculator/pkg/loggers"
	"context"
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
	muxHandler = handler.Decorate(muxHandler, loggingMiddleware())
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
	fmt.Printf("The only endpoint is available at http://localhost:%s/api/v1/calculate", port)
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
