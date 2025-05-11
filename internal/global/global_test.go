package global

import (
	"sync"
	"testing"
	"time"
)

// TestNewFutureSetGet проверяет, что Future возвращает установленное значение.
func TestNewFutureSetGet(t *testing.T) {
	f := NewFuture() // buffer size 1 citeturn20file1
	f.SetResult(3.14)
	got := f.Get()
	if got != 3.14 {
		t.Errorf("Future.Get() = %v, want %v", got, 3.14)
	}
}

// TestFutureGetBlocksUntilSet проверяет, что Get() блокируется до вызова SetResult.
func TestFutureGetBlocksUntilSet(t *testing.T) {
	f := NewFuture()
	// Запустим установку результата через 20ms
	d := 20 * time.Millisecond
	done := make(chan struct{})
	go func() {
		time.Sleep(d)
		f.SetResult(2.71)
		close(done)
	}()
	start := time.Now()
	got := f.Get()
	delta := time.Since(start)
	if got != 2.71 {
		t.Errorf("Future.Get() = %v, want %v", got, 2.71)
	}
	if delta < d {
		t.Errorf("Future.Get() returned too early: %v, want at least %v", delta, d)
	}
	<-done
}

// TestFuturesMapStoreLoad проверяет работу глобальной карты FuturesMap.
func TestFuturesMapStoreLoad(t *testing.T) {
	// Очистим карту перед тестом
	FuturesMap.Range(func(key, _ any) bool { FuturesMap.Delete(key); return true })
	f := NewFuture()
	FuturesMap.Store("k1", f) // sync.Map citeturn20file0
	v, ok := FuturesMap.Load("k1")
	if !ok {
		t.Fatal("expected key k1 in FuturesMap")
	}
	f2, ok := v.(*Future)
	if !ok {
		t.Fatalf("value type = %T, want *Future", v)
	}
	f2.SetResult(1.23)
	if got := f2.Get(); got != 1.23 {
		t.Errorf("Get after load = %v, want %v", got, 1.23)
	}
}

// TestTasksMapStoreLoad проверяет работу глобальной карты TasksMap.
func TestTasksMapStoreLoad(t *testing.T) {
	// Очистим карту перед тестом
	TasksMap.Range(func(key, _ any) bool { TasksMap.Delete(key); return true })
	task := &Task{ID: "t1", Arg1: 10, Arg2: 5, Operation: "*", OperationTime: 100}
	TasksMap.Store("t1", task) // sync.Map citeturn20file0
	v, ok := TasksMap.Load("t1")
	if !ok {
		t.Fatal("expected key t1 in TasksMap")
	}
	t2, ok := v.(*Task)
	if !ok {
		t.Fatalf("value type = %T, want *Task", v)
	}
	if t2.ID != task.ID || t2.Arg1 != task.Arg1 || t2.Arg2 != task.Arg2 || t2.Operation != task.Operation {
		t.Errorf("loaded task = %+v, want %+v", t2, task)
	}
}

// TestConcurrentAccess проверяет конкурентный доступ к картам.
func TestConcurrentAccess(t *testing.T) {
	// Очистим обе карты
	FuturesMap.Range(func(k, _ any) bool { FuturesMap.Delete(k); return true })
	TasksMap.Range(func(k, _ any) bool { TasksMap.Delete(k); return true })

	var wg sync.WaitGroup
	n := 100
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			f := NewFuture()
			FuturesMap.Store(i, f)
			f.SetResult(float64(i))
			if got := f.Get(); got != float64(i) {
				t.Errorf("concurrent future %d = %v", i, got)
			}
		}(i)
		go func(i int) {
			defer wg.Done()
			tk := &Task{ID: string(rune(i)), Arg1: 0, Arg2: 0, Operation: "+", OperationTime: 0}
			TasksMap.Store(i, tk)
			v, ok := TasksMap.Load(i)
			if !ok {
				t.Errorf("missing task %d", i)
				return
			}
			// тоже просто проверяем тип
			if _, ok := v.(*Task); !ok {
				t.Errorf("task %d type = %T", i, v)
			}
		}(i)
	}
	wg.Wait()
}
