package agent

import (
	"calculator/http/server/handler"
	"calculator/pkg/global"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCalculate(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var receivedID string
	var receivedResult float64

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var solved handler.SolvedTaskResponse
			err := json.NewDecoder(r.Body).Decode(&solved)
			if err != nil {
				t.Errorf("error decoding JSON: %v", err)
			}
			receivedID = solved.ID
			receivedResult = solved.Result
			wg.Done()
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to listen on port 8080: %v", err)
	}
	testServer := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: mux},
	}
	testServer.Start()
	defer testServer.Close()

	taskID := "test-task-id"
	operationTime := 10 // короткое время выполнения
	calculate(taskID, 3, 4, "+", operationTime)

	wg.Wait()
	if receivedID != taskID {
		t.Errorf("expected ID %s, got %s", taskID, receivedID)
	}
	if receivedResult != 7 {
		t.Errorf("expected result 7, got %f", receivedResult)
	}
}

// TestGetTask проверяет функцию getTask, имитируя GET /internal/task.
func TestGetTask(t *testing.T) {
	task := global.Task{
		ID:            "test-task-2",
		Arg1:          10,
		Arg2:          5,
		Operation:     "-",
		OperationTime: 10,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(task)
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to listen on port 8080: %v", err)
	}
	testServer := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: mux},
	}
	testServer.Start()
	defer testServer.Close()

	gotTask := getTask()
	if gotTask == nil {
		t.Fatal("expected task, got nil")
	}
	if gotTask.ID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, gotTask.ID)
	}
	if gotTask.Arg1 != task.Arg1 || gotTask.Arg2 != task.Arg2 {
		t.Errorf("expected task arguments %f, %f, got %f, %f", task.Arg1, task.Arg2, gotTask.Arg1, gotTask.Arg2)
	}
	if gotTask.Operation != task.Operation {
		t.Errorf("expected operation %s, got %s", task.Operation, gotTask.Operation)
	}
}
