package agent

import (
	"os"
	"testing"
)

func TestCalcOperations(t *testing.T) {
	tests := []struct {
		a, b float64
		op   string
		want float64
	}{
		{2, 3, "+", 5},
		{5, 2, "-", 3},
		{4, 3, "*", 12},
		{10, 2, "/", 5},
	}

	for _, tt := range tests {
		got := calc(tt.a, tt.b, tt.op)
		if got != tt.want {
			t.Errorf("calc(%v, %v, %q) = %v, want %v", tt.a, tt.b, tt.op, got, tt.want)
		}
	}
}

func TestCalcUnknown(t *testing.T) {
	got := calc(1, 1, "%")
	if got != 0 {
		t.Errorf("calc(1, 1, \"%%\") = %v, want 0", got)
	}
}

func TestGetenvDefault(t *testing.T) {
	os.Unsetenv("TEST_AGENT_POWER")
	got := getenv("TEST_AGENT_POWER", "42")
	if got != "42" {
		t.Errorf("getenv without set = %q, want %q", got, "42")
	}
}

func TestGetenvOverride(t *testing.T) {
	// Устанавливаем переменную окружения
	os.Setenv("TEST_AGENT_POWER", "17")
	defer os.Unsetenv("TEST_AGENT_POWER")
	got := getenv("TEST_AGENT_POWER", "99")
	if got != "17" {
		t.Errorf("getenv with set = %q, want %q", got, "17")
	}
}
