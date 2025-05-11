package calculator

import (
	"calculator/pkg/loggers"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	loggers.InitLogger("orchestrator", os.DevNull)
	code := m.Run()
	loggers.CloseAllLoggers()
	os.Exit(code)
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		expr    string
		want    []token
		wantErr bool
	}{
		{
			"3 + 4.5 * (2 - 1)",
			[]token{
				{typ: tokenNumber, val: "3"},
				{typ: tokenOperator, val: "+"},
				{typ: tokenNumber, val: "4.5"},
				{typ: tokenOperator, val: "*"},
				{typ: tokenLParen, val: "("},
				{typ: tokenNumber, val: "2"},
				{typ: tokenOperator, val: "-"},
				{typ: tokenNumber, val: "1"},
				{typ: tokenRParen, val: ")"},
			},
			false,
		},
		{"1.2.3 + 4", nil, true},
		{"3 & 4", nil, true},
	}

	for _, tt := range tests {
		got, err := tokenize(tt.expr)
		if (err != nil) != tt.wantErr {
			t.Errorf("tokenize(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("tokenize(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestPrecedence(t *testing.T) {
	tests := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2, "^": 0}
	for op, want := range tests {
		if got := precedence(op); got != want {
			t.Errorf("precedence(%q) = %d, want %d", op, got, want)
		}
	}
}

func TestShuntingYard(t *testing.T) {
	tokens := []token{
		{tokenNumber, "2"},
		{tokenOperator, "+"},
		{tokenNumber, "3"},
		{tokenOperator, "*"},
		{tokenNumber, "4"},
	}
	rpn, err := shuntingYard(tokens)
	if err != nil {
		t.Fatalf("shuntingYard error: %v", err)
	}
	want := []string{"2", "3", "4", "*", "+"}
	if len(rpn) != len(want) {
		t.Fatalf("rpn len = %d, want %d", len(rpn), len(want))
	}
	for i, tok := range rpn {
		if tok.val != want[i] {
			t.Errorf("rpn[%d] = %q, want %q", i, tok.val, want[i])
		}
	}
}

func TestEvalRPN_Errors(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token
		errSub string
	}{
		{"empty", []token{}, "invalid expression"},
		{"parse error", []token{{typ: tokenNumber, val: "x"}}, "invalid syntax"},
		{
			"div zero",
			[]token{{typ: tokenNumber, val: "1"}, {typ: tokenNumber, val: "0"}, {typ: tokenOperator, val: "/"}},
			"division by zero",
		},
	}
	for _, tt := range tests {
		_, err := evalRPN(tt.tokens)
		if err == nil || !strings.Contains(err.Error(), tt.errSub) {
			t.Errorf("%s: error = %v, want contain %q", tt.name, err, tt.errSub)
		}
	}
}

func evalRPNSync(tokens []token) (float64, error) {
	var stack []float64
	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			num, err := strconv.ParseFloat(tok.val, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, num)
		case tokenOperator:
			if len(stack) < 2 {
				return 0, errors.New("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var res float64
			switch tok.val {
			case "+":
				res = a + b
			case "-":
				res = a - b
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return 0, errors.New("division by zero")
				}
				res = a / b
			default:
				return 0, errors.New("unknown operator: " + tok.val)
			}
			stack = append(stack, res)
		}
	}
	if len(stack) != 1 {
		return 0, errors.New("invalid expression")
	}
	return stack[0], nil
}

func TestEvalRPNSync(t *testing.T) {
	tests := []struct {
		tokens []token
		want   float64
	}{
		{[]token{{tokenNumber, "2"}, {tokenNumber, "3"}, {tokenOperator, "+"}}, 5},
		{[]token{{tokenNumber, "10"}, {tokenNumber, "2"}, {tokenOperator, "-"}}, 8},
		{[]token{{tokenNumber, "6"}, {tokenNumber, "7"}, {tokenOperator, "*"}}, 42},
		{[]token{{tokenNumber, "8"}, {tokenNumber, "4"}, {tokenOperator, "/"}}, 2},
	}
	for _, tt := range tests {
		got, err := evalRPNSync(tt.tokens)
		if err != nil {
			t.Errorf("evalRPNSync(%v) error: %v", tt.tokens, err)
			continue
		}
		if got != tt.want {
			t.Errorf("evalRPNSync(%v) = %v, want %v", tt.tokens, got, tt.want)
		}
	}
}
