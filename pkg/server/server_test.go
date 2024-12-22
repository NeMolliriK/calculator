package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculatorHandler_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	handler := CalculatorHandler(logger)

	tests := []struct {
		name       string
		input      string
		expected   string
		expectCode int
	}{
		{"Simple addition", `{"expression": "2+2"}`, "4", http.StatusOK},
		{"Complex expression", `{"expression": "(2+3)*4"}`, "20", http.StatusOK},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/calculate",
			bytes.NewBuffer([]byte(test.input)),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != test.expectCode {
			t.Errorf("%s: expected status %d, got %d", test.name, test.expectCode, resp.StatusCode)
		}

		var responseData ResponseData
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			t.Fatalf("%s: failed to decode response: %v", test.name, err)
		}

		if responseData.Result != test.expected {
			t.Errorf(
				"%s: expected result %s, got %s",
				test.name,
				test.expected,
				responseData.Result,
			)
		}
	}
}

func TestCalculatorHandler_ValidationErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	handler := CalculatorHandler(logger)

	tests := []struct {
		name       string
		input      string
		expectCode int
	}{
		{"Invalid character", `{"expression": "2+2a"}`, http.StatusUnprocessableEntity},
		{"Unmatched parentheses", `{"expression": "(2+2"}`, http.StatusUnprocessableEntity},
		{"Empty expression", `{"expression": ""}`, http.StatusUnprocessableEntity},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/calculate",
			bytes.NewBuffer([]byte(test.input)),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != test.expectCode {
			t.Errorf("%s: expected status %d, got %d", test.name, test.expectCode, resp.StatusCode)
		}
	}
}

func TestCalculatorHandler_MethodNotAllowed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	handler := CalculatorHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/calculate", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
func TestCalculatorHandler_DivideByZero(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	handler := CalculatorHandler(logger)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/calculate",
		bytes.NewBuffer([]byte(`{"expression": "2/0"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}
func TestCalculatorHandler_Errors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	handler := CalculatorHandler(logger)

	tests := []struct {
		name       string
		method     string
		body       string
		expectCode int
		expectErr  string
	}{
		{"Invalid JSON", http.MethodPost, `{"expression":`, http.StatusBadRequest, "invalid JSON"},
		{
			"Missing expression",
			http.MethodPost,
			`{}`,
			http.StatusUnprocessableEntity,
			"no expression provided",
		},
		{
			"Method not allowed",
			http.MethodGet,
			"",
			http.StatusMethodNotAllowed,
			"only POST method is allowed",
		},
		{
			"Invalid character",
			http.MethodPost,
			`{"expression": "2+2a"}`,
			http.StatusUnprocessableEntity,
			"an invalid character is present in the expression: a",
		},
	}

	for _, test := range tests {
		req := httptest.NewRequest(
			test.method,
			"/api/v1/calculate",
			bytes.NewBuffer([]byte(test.body)),
		)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != test.expectCode {
			t.Errorf("%s: expected status %d, got %d", test.name, test.expectCode, resp.StatusCode)
		}

		var errData ErrorData
		if resp.StatusCode >= 400 {
			if err := json.NewDecoder(resp.Body).Decode(&errData); err != nil {
				t.Fatalf("%s: failed to decode error response: %v", test.name, err)
			}
			if errData.Error != test.expectErr {
				t.Errorf("%s: expected error %q, got %q", test.name, test.expectErr, errData.Error)
			}
		}
	}
}
func TestLoggingMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{}))
	handler := LoggingMiddleware(CalculatorHandler(logger), logger)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/calculate",
		bytes.NewBuffer([]byte(`{"expression": "2+2"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestErrorRecoveryMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{}))
	handler := ErrorRecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
		panic("Test panic")
	}, logger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}
func TestLoggingMiddlewareIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{}))
	handler := LoggingMiddleware(CalculatorHandler(logger), logger)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/calculate",
		bytes.NewBuffer([]byte(`{"expression": "2+2"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
}
