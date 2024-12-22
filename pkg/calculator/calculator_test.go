package calculator

import (
	"log/slog"
	"testing"
)

func TestValidateExpression(t *testing.T) {
	tests := []struct {
		expression string
		shouldFail bool
		name       string
	}{
		{"2+2", false, "Valid simple expression"},
		{"(2+2)*2", false, "Valid expression with parentheses"},
		{"(2+2*3", true, "Unmatched opening parenthesis"},
		{"2+2)*3", true, "Unmatched closing parenthesis"},
		{"2++3", true, "Invalid operator sequence"},
		{"+2+3", true, "Invalid operator at beginning"},
		{"2+3-", true, "Invalid operator at end"},
	}

	for _, test := range tests {
		err := ValidateExpression(test.expression)
		if (err != nil) != test.shouldFail {
			t.Errorf("%s: expected failure: %v, got: %v", test.name, test.shouldFail, err)
		}
	}
}

func TestCalc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))

	tests := []struct {
		expression string
		expected   float64
		name       string
	}{
		{"2+2", 4, "Addition"},
		{"2-2", 0, "Subtraction"},
		{"2*3", 6, "Multiplication"},
		{"6/2", 3, "Division"},
		{"2+2*2", 6, "Operator precedence"},
		{"(2+2)*2", 8, "Parentheses"},
	}

	for _, test := range tests {
		result, err := Calc(test.expression, logger)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}
		if result != test.expected {
			t.Errorf("%s: expected %f, got %f", test.name, test.expected, result)
		}
	}
}

func TestInvalidCalc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))

	tests := []struct {
		expression string
		name       string
	}{
		{"2//2", "Double operator"},
		{"2+2a", "Invalid character"},
		{"(2+2", "Unmatched opening parenthesis"},
		{"2+2)", "Unmatched closing parenthesis"},
	}

	for _, test := range tests {
		_, err := Calc(test.expression, logger)
		if err == nil {
			t.Errorf("%s: expected error, got nil", test.name)
		}
	}
}
func TestValidateExpressionAdditional(t *testing.T) {
	tests := []struct {
		expression string
		shouldFail bool
		name       string
	}{
		{"", true, "Empty expression"},
		{"2+3*(4-2)/", true, "Trailing operator"},
		{"((((2+2))))", false, "Nested parentheses"},
		{"2/0", false, "Division by zero (should not fail validation)"},
	}

	for _, test := range tests {
		err := ValidateExpression(test.expression)
		if (err != nil) != test.shouldFail {
			t.Errorf("%s: expected failure: %v, got: %v", test.name, test.shouldFail, err)
		}
	}
}
func TestCalcPanicRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))
	_, err := Calc("((2+3)", logger)
	if err == nil {
		t.Errorf("Expected error due to panic, got nil")
	}
}
func TestCalculateComplexExpressions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))

	tests := []struct {
		expression string
		expected   float64
		name       string
	}{
		{"(2+3)*(4-2)", 10, "Nested parentheses"},
		{"2+3*4/2", 8, "Mixed operations"},
		{"((2+3)*4)/2", 10, "Double nested parentheses"},
	}

	for _, test := range tests {
		result, err := Calc(test.expression, logger)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
		}
		if result != test.expected {
			t.Errorf("%s: expected %f, got %f", test.name, test.expected, result)
		}
	}
}
func TestCalcAdditionalCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))

	tests := []struct {
		expression string
		expectErr  bool
		name       string
	}{
		{"2+++2", true, "Invalid multiple operators"},
		{"((2+3))", false, "Valid nested parentheses"},
		{"2**3", true, "Invalid double operator"},
	}

	for _, test := range tests {
		_, err := Calc(test.expression, logger)
		if (err != nil) != test.expectErr {
			t.Errorf("%s: expected error: %v, got: %v", test.name, test.expectErr, err)
		}
	}
}
func TestCalcEdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{}))

	tests := []struct {
		expression string
		expectErr  bool
		name       string
	}{
		{"2///2", true, "Triple operators"},
		{"((2+3)", true, "Unmatched parentheses"},
		{"2+((3-1)", true, "Nested unmatched parentheses"},
		{"2+(3-(1+2))", false, "Valid nested parentheses"},
		{"1000000000*1000000000", false, "Large multiplication"},
	}

	for _, test := range tests {
		_, err := Calc(test.expression, logger)
		if (err != nil) != test.expectErr {
			t.Errorf("%s: expected error: %v, got: %v", test.name, test.expectErr, err)
		}
	}
}
