package calculator

import (
	"calculator/internal/global"
	"calculator/pkg/loggers"
	"errors"
	"os"
	"strconv"
	"unicode"

	"github.com/google/uuid"
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

func tokenize(expr string) ([]token, error) {
	var tokens []token
	logger := loggers.GetLogger("orchestrator")
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

func wrapValueAsFuture(val float64) *global.Future {
	future := global.NewFuture()
	future.SetResult(val)
	return future
}

func evalRPN(tokens []token) (float64, error) {
	var stack []*global.Future
	for _, tok := range tokens {
		switch tok.typ {
		case tokenNumber:
			num, err := strconv.ParseFloat(tok.val, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, wrapValueAsFuture(num))
		case tokenOperator:
			if len(stack) < 2 {
				return 0, errors.New("invalid expression")
			}
			b := stack[len(stack)-1].Get()
			a := stack[len(stack)-2].Get()
			stack = stack[:len(stack)-2]
			future := global.NewFuture()
			stack = append(stack, future)
			var err error
			t := 1000
			switch tok.val {
			case "+":
				t, err = strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
				if err != nil {
					t = 1000
				}
			case "-":
				t, err = strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
				if err != nil {
					t = 1000
				}
			case "*":
				t, err = strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
				if err != nil {
					t = 1000
				}
			case "/":
				t, err = strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
				if b == 0 {
					return 0, errors.New("division by zero")
				}
				if err != nil {
					t = 1000
				}
			default:
				return 0, errors.New("unknown operator: " + tok.val)
			}
			task := global.Task{
				ID:            uuid.New().String(),
				Arg1:          a,
				Arg2:          b,
				Operation:     tok.val,
				OperationTime: t,
			}
			global.TasksMap.Store(task.ID, &task)
			global.FuturesMap.Store(task.ID, future)
		}
	}
	if len(stack) != 1 {
		return 0, errors.New("invalid expression")
	}
	result := stack[0].Get()
	return result, nil
}

type db interface {
	UpdateExpressionStatus(id string, status string) error
	GetExpressionByID(id string) (*global.ExpressionDTO, error)
	UpdateExpressionResult(id string, result float64) error
}

func Calc(store db, expressionID string) {
	err := store.UpdateExpressionStatus(expressionID, "processing")
	if err != nil {
		panic(err)
	}
	expression, err := store.GetExpressionByID(expressionID)
	if err != nil {
		panic(err)
	}
	tokens, err := tokenize(expression.Data)
	if err != nil {
		err := store.UpdateExpressionStatus(expressionID, "calculation error: "+err.Error())
		if err != nil {
			panic(err)
		}
		return
	}
	rpn, err := shuntingYard(tokens)
	if err != nil {
		err := store.UpdateExpressionStatus(expressionID, "calculation error: "+err.Error())
		if err != nil {
			panic(err)
		}
		return
	}
	res, err := evalRPN(rpn)
	if err != nil {
		err := store.UpdateExpressionStatus(expressionID, "calculation error: "+err.Error())
		if err != nil {
			panic(err)
		}
	} else {
		err := store.UpdateExpressionStatus(expressionID, "completed")
		if err != nil {
			panic(err)
		}
	}
	err = store.UpdateExpressionResult(expressionID, res)
	if err != nil {
		panic(err)
	}
}
