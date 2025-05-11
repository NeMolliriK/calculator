//go:build integration
// +build integration

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"calculator/internal/agent"
	"calculator/internal/application"
	"calculator/internal/database"
	"calculator/pkg/loggers"
)

type tokenResponse struct {
	Info  string `json:"info"`
	Token string `json:"token"`
}

type idResponse struct {
	ID string `json:"id"`
}

type expressionResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

func waitFor(timeout time.Duration, tick time.Duration, fn func() (bool, error)) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := fn()
		if ok {
			return nil
		}
		if time.Now().After(deadline) {
			if err == nil {
				err = fmt.Errorf("timeout")
			}
			return err
		}
		time.Sleep(tick)
	}
}

func doJSON(method, url, token string, body any, v any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if v != nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(v); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func registerAndLogin(t *testing.T, baseURL string, suffix string) string {
	creds := map[string]string{
		"login":    "user" + suffix,
		"password": "pass" + suffix,
	}
	resp, err := doJSON("POST", baseURL+"/api/v1/register", "", creds, nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("register failed: %v, status=%d", err, resp.StatusCode)
	}
	var tok tokenResponse
	resp, err = doJSON("POST", baseURL+"/api/v1/login", "", creds, &tok)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %v, status=%d", err, resp.StatusCode)
	}
	return tok.Token
}

func startSystem(t *testing.T, port int, withAgent bool) context.CancelFunc {
	t.Helper()

	os.Setenv("PORT", strconv.Itoa(port))
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("TIME_ADDITION_MS", "5")
	os.Setenv("TIME_SUBTRACTION_MS", "5")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "5")
	os.Setenv("TIME_DIVISIONS_MS", "5")

	_ = os.Remove("sqlite.db")
	t.Cleanup(func() {
		if database.DB != nil {
			if sqlDB, err := database.DB.DB(); err == nil {
				sqlDB.Close()
			}
		}
		_ = os.Remove("sqlite.db")
	})

	loggers.InitLogger("server", os.DevNull)
	loggers.InitLogger("orchestrator", os.DevNull)
	loggers.InitLogger("general", os.DevNull)
	loggers.InitLogger("agent", os.DevNull)
	t.Cleanup(func() { loggers.CloseAllLoggers() })

	database.Init()

	ctx, cancel := context.WithCancel(context.Background())

	app := application.New()
	go func() { _ = app.Run(ctx) }()

	if withAgent {
		go agent.Run()
	}

	baseURL := fmt.Sprintf("http://localhost:%d/api/v1/register", port)
	err := waitFor(5*time.Second, 100*time.Millisecond, func() (bool, error) {
		resp, err := http.Post(baseURL, "application/json", strings.NewReader("{}"))
		if err != nil {
			return false, nil
		}
		_ = resp.Body.Close()
		return true, nil
	})
	if err != nil {
		cancel()
		t.Fatalf("server did not start: %v", err)
	}

	return cancel
}

func TestHappyPathEndToEnd(t *testing.T) {
	const port = 18080
	cancel := startSystem(t, port, true)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	token := registerAndLogin(t, baseURL, "1")

	var idResp idResponse
	_, err := doJSON("POST", baseURL+"/api/v1/calculate", token, map[string]string{"expression": "2+2*2"}, &idResp)
	if err != nil || idResp.ID == "" {
		t.Fatalf("calculate failed: %v", err)
	}

	var res expressionResponse
	pollURL := fmt.Sprintf("%s/api/v1/expressions/%s", baseURL, idResp.ID)
	err = waitFor(10*time.Second, 100*time.Millisecond, func() (bool, error) {
		_, err := doJSON("GET", pollURL, token, nil, &res)
		if err != nil {
			return false, err
		}
		return res.Status == "completed", nil
	})
	if err != nil || res.Result != 6 {
		t.Fatalf("unexpected calculation result: %+v, err=%v", res, err)
	}
}

func TestCalculateUnauthorized(t *testing.T) {
	const port = 18081
	cancel := startSystem(t, port, true)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	resp, err := doJSON("POST", baseURL+"/api/v1/calculate", "", map[string]string{"expression": "1+1"}, nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCalculateInvalidExpression(t *testing.T) {
	const port = 18082
	cancel := startSystem(t, port, true)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	token := registerAndLogin(t, baseURL, "2")

	var idResp idResponse
	resp, err := doJSON("POST", baseURL+"/api/v1/calculate", token, map[string]string{"expression": "2+*2"}, &idResp)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created for accepted task, got %d", resp.StatusCode)
	}
	if idResp.ID == "" {
		t.Fatalf("no ID returned for invalid expression")
	}

	var res expressionResponse
	pollURL := fmt.Sprintf("%s/api/v1/expressions/%s", baseURL, idResp.ID)
	err = waitFor(5*time.Second, 100*time.Millisecond, func() (bool, error) {
		_, err := doJSON("GET", pollURL, token, nil, &res)
		if err != nil {
			return false, err
		}
		return strings.HasPrefix(res.Status, "calculation error"), nil
	})
	if err != nil {
		t.Fatalf("expression did not transition to calculation error: %v (status=%s)", err, res.Status)
	}
}

func TestUserIsolation(t *testing.T) {
	const port = 18083
	cancel := startSystem(t, port, true)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	tokenA := registerAndLogin(t, baseURL, "A")
	tokenB := registerAndLogin(t, baseURL, "B")

	var idResp idResponse
	_, err := doJSON("POST", baseURL+"/api/v1/calculate", tokenA, map[string]string{"expression": "3*3"}, &idResp)
	if err != nil || idResp.ID == "" {
		t.Fatalf("calc by user A failed: %v", err)
	}

	var res expressionResponse
	pollURL := fmt.Sprintf("%s/api/v1/expressions/%s", baseURL, idResp.ID)
	_ = waitFor(5*time.Second, 100*time.Millisecond, func() (bool, error) {
		_, _ = doJSON("GET", pollURL, tokenA, nil, &res)
		return res.Status == "completed", nil
	})

	var listB struct {
		Expressions []expressionResponse `json:"expressions"`
	}
	_, err = doJSON("GET", baseURL+"/api/v1/expressions", tokenB, nil, &listB)
	if err != nil {
		t.Fatalf("list for B failed: %v", err)
	}
	for _, e := range listB.Expressions {
		if e.ID == idResp.ID {
			t.Fatalf("user B sees expression of user A: %+v", e)
		}
	}

	var listA struct {
		Expressions []expressionResponse `json:"expressions"`
	}
	_, err = doJSON("GET", baseURL+"/api/v1/expressions", tokenA, nil, &listA)
	if err != nil {
		t.Fatalf("list for A failed: %v", err)
	}
	found := false
	for _, e := range listA.Expressions {
		if e.ID == idResp.ID {
			found = true
			if e.Result != 9 {
				t.Fatalf("wrong result for user A: %+v", e)
			}
		}
	}
	if !found {
		t.Fatalf("user A does not see own expression")
	}
}

func TestCompletedEvenWithoutAgent(t *testing.T) {
	const port = 18084
	cancel := startSystem(t, port, false)
	defer cancel()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	token := registerAndLogin(t, baseURL, "3")

	var idResp idResponse
	_, err := doJSON("POST", baseURL+"/api/v1/calculate", token, map[string]string{"expression": "5+5"}, &idResp)
	if err != nil {
		t.Fatalf("calculate failed: %v", err)
	}

	var res expressionResponse
	pollURL := fmt.Sprintf("%s/api/v1/expressions/%s", baseURL, idResp.ID)
	err = waitFor(5*time.Second, 100*time.Millisecond, func() (bool, error) {
		_, err := doJSON("GET", pollURL, token, nil, &res)
		if err != nil {
			return false, err
		}
		return res.Status == "completed", nil
	})
	if err != nil {
		t.Fatalf("expression did not complete without agent: %v", err)
	}
	if res.Result != 10 {
		t.Fatalf("unexpected result without agent: %+v", res)
	}
}
