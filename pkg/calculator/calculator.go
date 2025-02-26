package calculator

import (
	"calculator/pkg/loggers"
	"errors"
	"strconv"
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

type resultWithError struct {
	value float64
	err   error
}

type Future struct {
	result chan resultWithError
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
	return &Future{result: make(chan resultWithError, 1)}
}

func (f *Future) SetResult(val float64, err error) {
	f.result <- resultWithError{value: val, err: err}
}

func (f *Future) Get() (float64, error) {
	res := <-f.result
	return res.value, res.err
}

func calculate(a, b float64, operator string, f *Future, sem chan struct{}) {
	defer func() { <-sem }()
	switch operator {
	case "+":
		f.SetResult(a+b, nil)
	case "-":
		f.SetResult(a-b, nil)
	case "*":
		f.SetResult(a*b, nil)
	case "/":
		if b == 0 {
			f.SetResult(0, errors.New("division by zero"))
		}
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
	sem := make(chan struct{}, 10)
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

func Calc(expression string) (float64, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return 0, err
	}
	rpn, err := shuntingYard(tokens)
	if err != nil {
		return 0, err
	}
	return evalRPN(rpn)
}
