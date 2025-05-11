package middleware

import (
	"bytes"
	"calculator/pkg/loggers"
	"io"
	"net/http"
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

func LoggingMiddleware() func(next http.Handler) http.Handler {
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
