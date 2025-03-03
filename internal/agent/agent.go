package agent

import (
	"bytes"
	"calculator/http/server/handler"
	"calculator/pkg/global"
	"encoding/json"
	"net/http"
	"time"
)

func getTask() *global.Task {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()
	var task global.Task
	json.NewDecoder(resp.Body).Decode(&task)
	return &task
}

func sendResult(id string, result float64) {
	data := handler.SolvedTaskResponse{ID: id, Result: result}
	body, _ := json.Marshal(data)
	http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(body))
}

func calculate(id string, a, b float64, operator string, operationTime int) {
	time.Sleep(time.Duration(operationTime) * time.Millisecond)
	var result float64
	switch operator {
	case "+":
		result = a + b
	case "-":
		result = a - b
	case "*":
		result = a * b
	case "/":
		result = a / b
	}
	sendResult(id, result)
}

func Run() {
	for {
		task := getTask()
		if task != nil {
			go calculate(task.ID, task.Arg1, task.Arg2, task.Operation, task.OperationTime)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
