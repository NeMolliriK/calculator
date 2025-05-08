package calculator_test

import (
	"calculator/pkg/calculator"
	"calculator/pkg/global"
	"calculator/pkg/loggers"
	"os"
	"strings"
	"testing"
	"time"
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

func processTasks() {
	global.TasksMap.Range(func(key, value interface{}) bool {
		task := value.(*global.Task)
		var res float64
		switch task.Operation {
		case "+":
			res = task.Arg1 + task.Arg2
		case "-":
			res = task.Arg1 - task.Arg2
		case "*":
			res = task.Arg1 * task.Arg2
		case "/":
			res = task.Arg1 / task.Arg2
		}
		if futureInterface, ok := global.FuturesMap.Load(task.ID); ok {
			future := futureInterface.(*global.Future)
			future.SetResult(res)
		}
		global.TasksMap.Delete(key)
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

func TestCalcSimple(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "2+2*2",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	go func() {
		for {
			processTasks()
			time.Sleep(10 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
			}
		}
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if expr.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", expr.Status)
	}
	if expr.Result != 6 {
		t.Errorf("expected result 6, got %f", expr.Result)
	}
}

func TestCalcWithParentheses(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "2*(3+4)",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	go func() {
		for {
			processTasks()
			time.Sleep(10 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
			}
		}
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if expr.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", expr.Status)
	}
	if expr.Result != 14 {
		t.Errorf("expected result 14, got %f", expr.Result)
	}
}

func TestCalcDivision(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "8/2",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	go func() {
		for {
			processTasks()
			time.Sleep(10 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
			}
		}
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if expr.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", expr.Status)
	}
	if expr.Result != 4 {
		t.Errorf("expected result 4, got %f", expr.Result)
	}
}

func TestCalcDivisionByZero(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "10/0",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	go func() {
		for {
			processTasks()
			time.Sleep(10 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
			}
		}
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if !strings.Contains(expr.Status, "division by zero") {
		t.Errorf("expected division by zero error, got %s", expr.Status)
	}
}

func TestCalcInvalidExpression(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "2+",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if !strings.Contains(expr.Status, "invalid expression") {
		t.Errorf("expected invalid expression error, got %s", expr.Status)
	}
}

func TestCalcInvalidNumberFormat(t *testing.T) {
	clearGlobalMaps()
	expr := &global.ExpressionDTO{
		Data:   "2..2+3",
		Status: "pending",
	}
	doneCh := make(chan struct{})
	go func() {
		calculator.Calc(expr)
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("Calc timed out")
	}
	if !strings.Contains(expr.Status, "invalid number format") {
		t.Errorf("expected invalid number format error, got %s", expr.Status)
	}
}

func TestFuture(t *testing.T) {
	future := global.NewFuture()
	go func() {
		time.Sleep(10 * time.Millisecond)
		future.SetResult(42)
	}()
	res := future.Get()
	if res != 42 {
		t.Errorf("expected 42, got %f", res)
	}
}
