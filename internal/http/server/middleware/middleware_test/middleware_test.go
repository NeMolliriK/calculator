package middleware_test

import (
	"bytes"
	"calculator/internal/http/server/middleware"
	"calculator/pkg/loggers"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	loggers.InitLogger("server", os.DevNull)
	code := m.Run()
	loggers.CloseAllLoggers()
	os.Exit(code)
}

func TestResponseWriterWrapper(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &middleware.ResponseWriterWrapper{
		ResponseWriter: rr,
		Body:           &bytes.Buffer{},
		StatusCode:     http.StatusOK,
	}
	rw.WriteHeader(http.StatusTeapot)
	n, err := rw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len("hello") {
		t.Errorf("Write returned %d, want %d", n, len("hello"))
	}
	if rw.StatusCode != http.StatusTeapot {
		t.Errorf("StatusCode = %d, want %d", rw.StatusCode, http.StatusTeapot)
	}
	if got := rw.Body.String(); got != "hello" {
		t.Errorf("Body = %q, want %q", got, "hello")
	}
	if rr.Code != http.StatusTeapot {
		t.Errorf("Recorder.Code = %d, want %d", rr.Code, http.StatusTeapot)
	}
	if rr.Body.String() != "hello" {
		t.Errorf("Recorder.Body = %q, want %q", rr.Body.String(), "hello")
	}
}

func TestLoggingMiddleware_PassThrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
		w.Write(data)
	})
	wrapped := middleware.LoggingMiddleware()(next)
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("data"))
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if body := rr.Body.String(); body != "data" {
		t.Errorf("Body = %q, want %q", body, "data")
	}
}

func TestErrorRecoveryMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	wrapped := middleware.ErrorRecoveryMiddleware()(next)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	var resp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if resp.Error != "internal server error" {
		t.Errorf("Error = %q, want %q", resp.Error, "internal server error")
	}
}

func TestJWTMiddleware_NoHeader(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	wrapper := middleware.JWTMiddleware()(next)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	wrapper.ServeHTTP(rr, req)
	if nextCalled {
		t.Error("next handler should not be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	var resp struct {
		Error string `json:"error"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Error != "unauthorized" {
		t.Errorf("Error = %q, want %q", resp.Error, "unauthorized")
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	wrapper := middleware.JWTMiddleware()(next)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rr := httptest.NewRecorder()
	wrapper.ServeHTTP(rr, req)
	if nextCalled {
		t.Error("next handler should not be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	var resp struct {
		Error string `json:"error"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Error != "invalid token" {
		t.Errorf("Error = %q, want %q", resp.Error, "invalid token")
	}
}

func TestJWTMiddleware_InvalidClaims(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	defer os.Unsetenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"foo": "bar"})
	tokenStr, _ := token.SignedString([]byte("secret"))
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	wrapper := middleware.JWTMiddleware()(next)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	wrapper.ServeHTTP(rr, req)
	if nextCalled {
		t.Error("next handler should not be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	var resp struct {
		Error string `json:"error"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Error != "invalid claims" {
		t.Errorf("Error = %q, want %q", resp.Error, "invalid claims")
	}
}

func TestJWTMiddleware_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	defer os.Unsetenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 42})
	tokenStr, _ := token.SignedString([]byte("secret"))
	var seenID uint
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(middleware.UserIDKey)
		if id, ok := v.(uint); ok {
			seenID = id
		}
		w.WriteHeader(http.StatusOK)
	})
	wrapper := middleware.JWTMiddleware()(next)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	wrapper.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", rr.Code, http.StatusOK)
	}
	if seenID != 42 {
		t.Errorf("UserID = %d, want %d", seenID, 42)
	}
}
