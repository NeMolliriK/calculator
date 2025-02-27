package calculator

import (
	"calculator/pkg/loggers"
	"calculator/pkg/structures"
	"errors"
	"os"
	"strconv"
	"time"
	"unicode"
)

type tokenType int

const (
	tokenNumber tokenType = iota
	tokenOperator
	tokenLParen
	tokenRParen
)

type token struct {
	typ tokenType
	val string
}

type Future struct {
	result chan structures.ResultWithError
}

func tokenize(expr string) ([]token, error) {
	var tokens []token
	logger := loggers.GetLogger("calculator")
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == ' ' {
			i++
			continue
		}
		// Число может состоять из цифр и, возможно, одной десятичной точки.
		if unicode.IsDigit(rune(ch)) || ch == '.' {
			start := i
			dotCount := 0
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.') {
				if expr[i] == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, errors.New("invalid number format")
					}
				}
				i++
			}
			tokens = append(tokens, token{typ: tokenNumber, val: expr[start:i]})
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			tokens = append(tokens, token{typ: tokenOperator, val: string(ch)})
			i++
		} else if ch == '(' {
			tokens = append(tokens, token{typ: tokenLParen, val: string(ch)})
			i++
		} else if ch == ')' {
			tokens = append(tokens, token{typ: tokenRParen, val: string(ch)})
			i++
		} else {
			return nil, errors.New("invalid character: " + string(ch))
		}
	}
	logger.Debug("tokenize", "tokens", tokens)
	return tokens, nil
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func shuntingYard(tokens []token) ([]token, error) {
	var output []token
	var stack []token
	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			output = append(output, tok)
		case tokenOperator:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.typ == tokenOperator && precedence(top.val) >= precedence(tok.val) {
					output = append(output, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, tok)
		case tokenLParen:
			stack = append(stack, tok)
		case tokenRParen:
			foundLParen := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.typ == tokenLParen {
					foundLParen = true
					break
				} else {
					output = append(output, top)
				}
			}
			if !foundLParen {
				return nil, errors.New("bracket mismatch")
			}
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.typ == tokenLParen || top.typ == tokenRParen {
			return nil, errors.New("bracket mismatch")
		}
		output = append(output, top)
	}
	return output, nil
}

func NewFuture() *Future {
	return &Future{result: make(chan structures.ResultWithError, 1)}
}

func (f *Future) SetResult(val float64, err error) {
	f.result <- structures.ResultWithError{Value: val, Err: err}
}

func (f *Future) Get() (float64, error) {
	res := <-f.result
	return res.Value, res.Err
}

func calculate(a, b float64, operator string, f *Future, sem chan struct{}) {
	defer func() { <-sem }()
	switch operator {
	case "+":
		t, err := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
		if err != nil {
			t = 1000
		}
		time.Sleep(time.Duration(t) * time.Millisecond)
		f.SetResult(a+b, nil)
	case "-":
		t, err := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
		if err != nil {
			t = 1000
		}
		time.Sleep(time.Duration(t) * time.Millisecond)
		f.SetResult(a-b, nil)
	case "*":
		t, err := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
		if err != nil {
			t = 1000
		}
		time.Sleep(time.Duration(t) * time.Millisecond)
		f.SetResult(a*b, nil)
	case "/":
		if b == 0 {
			f.SetResult(0, errors.New("division by zero"))
		}
		t, err := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
		if err != nil {
			t = 1000
		}
		time.Sleep(time.Duration(t) * time.Millisecond)
		f.SetResult(a/b, nil)
	default:
		f.SetResult(0, errors.New("unknown operator: "+operator))
	}
}

func WrapValueAsFuture(val float64) *Future {
	future := NewFuture()
	future.SetResult(val, nil)
	return future
}

func evalRPN(tokens []token) (float64, error) {
	var stack []*Future
	n, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil {
		n = 10
	}
	sem := make(chan struct{}, n)
	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			num, err := strconv.ParseFloat(tok.val, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, WrapValueAsFuture(num))
		case tokenOperator:
			if len(stack) < 2 {
				return 0, errors.New("invalid expression")
			}
			b, err := stack[len(stack)-1].Get()
			if err != nil {
				return 0, err
			}
			a, err := stack[len(stack)-2].Get()
			if err != nil {
				return 0, err
			}
			stack = stack[:len(stack)-2]
			sem <- struct{}{}
			future := NewFuture()
			stack = append(stack, future)
			go calculate(a, b, tok.val, future, sem)
		}
	}
	if len(stack) != 1 {
		return 0, errors.New("invalid expression")
	}
	result, err := stack[0].Get()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func Calc(expression *structures.Expression) {
	expression.Status = "processing"
	tokens, err := tokenize(expression.Data)
	if err != nil {
		expression.Status = "calculation error: " + err.Error()
		return
	}
	rpn, err := shuntingYard(tokens)
	if err != nil {
		expression.Status = "calculation error: " + err.Error()
		return
	}
	res, err := evalRPN(rpn)
	if err != nil {
		expression.Status = "calculation error: " + err.Error()
	} else {
		expression.Status = "completed"
	}
	expression.Result = res
}
