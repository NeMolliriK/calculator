package agent

import (
	"bytes"
	"calculator/internal/http/server/handler"
	"calculator/pkg/global"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, falling back to system environment variables")
	}
	n, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil {
		n = 10
	}
	sem := make(chan struct{}, n)
	for {
		task := getTask()
		if task != nil {
			sem <- struct{}{}
			go func(id string, a, b float64, operator string, operationTime int) {
				defer func() { <-sem }()
				calculate(id, a, b, operator, operationTime)
			}(task.ID, task.Arg1, task.Arg2, task.Operation, task.OperationTime)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
