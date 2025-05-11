package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calculator/internal/database"
	"calculator/internal/http/server/middleware"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}

	database.DB = db
	if err := database.DB.AutoMigrate(&database.User{}, &database.Expression{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func TestCalculatorAPIHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/calculate", nil)
	rr := httptest.NewRecorder()

	calculatorAPIHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET -> %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if !strings.Contains(resp["error"], "only POST") {
		t.Errorf("body = %v", resp)
	}
}

func TestCalculatorAPIHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString("{bad json}"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uint(1)))
	rr := httptest.NewRecorder()

	calculatorAPIHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Bad JSON -> %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCalculatorAPIHandler_NoExpression(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"expression": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uint(1)))
	rr := httptest.NewRecorder()

	calculatorAPIHandler(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Empty expr -> %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestCalculatorAPIHandler_Unauthorized(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"expression": "1+1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	calculatorAPIHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("No userID -> %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestExpressionsHandler_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rr := httptest.NewRecorder()

	expressionsHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("No auth -> %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestExpressionHandler_NotFoundAndForbidden(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uint(1)))
	rr := httptest.NewRecorder()

	expressionHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Missing ID -> %d, want %d", rr.Code, http.StatusBadRequest)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/expressions/nonexist", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uint(1)))
	rr = httptest.NewRecorder()

	expressionHandler(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Not found -> %d, want %d", rr.Code, http.StatusNotFound)
	}

	expr := &database.Expression{ID: "expr1", UserID: 2, Data: "1+1", Status: "pending"}
	database.DB.Create(expr)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/expressions/expr1", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uint(1)))
	rr = httptest.NewRecorder()

	expressionHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("Forbidden -> %d, want %d", rr.Code, http.StatusForbidden)
	}
}
