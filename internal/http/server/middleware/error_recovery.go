package middleware

import (
	"calculator/pkg/loggers"
	"encoding/json"
	"net/http"
)

type errorData struct {
	Error string `json:"error"`
}

func ErrorRecoveryMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger := loggers.GetLogger("server")
					logger.Error("panic recovered", "error", rec)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(errorData{Error: "internal server error"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
