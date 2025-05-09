package handler_test

import (
	"calculator/http/server/handler"
	"calculator/pkg/global"
	"calculator/pkg/loggers"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func clearGlobalMaps() {
	global.TasksMap.Range(func(key, value interface{}) bool {
		global.TasksMap.Delete(key)
		return true
	})
	global.FuturesMap.Range(func(key, value interface{}) bool {
		global.FuturesMap.Delete(key)
		return true
	})
}

func TestMain(m *testing.M) {
	loggers.InitLogger("orchestrator", os.DevNull)
	loggers.InitLogger("server", os.DevNull)
	loggers.InitLogger("general", os.DevNull)
	code := m.Run()
	loggers.CloseAllLoggers()
	os.Exit(code)
}

func TestCalculateEndpoint(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{"expression": "2+2*2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
	var idResp struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &idResp)
	if err != nil || idResp.ID == "" {
		t.Errorf("expected valid ID in response, got error: %v", err)
	}
}

func TestCalculateEndpoint_InvalidJSON(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestCalculateEndpoint_NoExpression(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{"expression": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", rr.Code)
	}
}

func TestExpressionsEndpoint(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{"expression": "2+2*2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rrGet.Code)
	}
	var exprsResp struct {
		Expressions []struct {
			ID     string  `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"Result"`
		} `json:"expressions"`
	}
	err = json.Unmarshal(rrGet.Body.Bytes(), &exprsResp)
	if err != nil {
		t.Errorf("error unmarshalling response: %v", err)
	}
	if len(exprsResp.Expressions) == 0 {
		t.Error("expected at least one expression in response")
	}
}

func TestExpressionByIDEndpoint(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{"expression": "2+2*2"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
	var idResp struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &idResp)
	if err != nil || idResp.ID == "" {
		t.Errorf("expected valid ID, got error: %v", err)
	}
	url := "/api/v1/expressions/" + idResp.ID
	reqGet := httptest.NewRequest(http.MethodGet, url, nil)
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rrGet.Code)
	}
}

func TestExpressionByIDEndpoint_InvalidID(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestInternalTaskEndpoint_NoTask(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/internal/task", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 when no task, got %d", rr.Code)
	}
}

func TestInternalTaskEndpoint_InvalidJSON(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	reqBody := strings.NewReader(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/task", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestInternalTaskEndpoint_InvalidMethod(t *testing.T) {
	clearGlobalMaps()
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPut, "/internal/task", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestIndexHandler(t *testing.T) {
	h, err := handler.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", ct)
	}
}
