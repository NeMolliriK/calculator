package calculator

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"unicode"
)

type SubExpression struct {
	Start      int
	End        int
	Expression []string
}

func isArithmeticSign(r rune) bool {
	for _, item := range [6]rune{'+', '-', '/', '*', '(', ')'} {
		if item == r {
			return true
		}
	}
	return false
}
func isExactlyArithmeticSign(r rune) bool {
	for _, item := range [4]rune{'+', '-', '/', '*'} {
		if item == r {
			return true
		}
	}
	return false
}
func multiplyOrDivide(singleton []string) float64 {
	num, _ := strconv.ParseFloat(singleton[0], 64)
	num_2, _ := strconv.ParseFloat(singleton[2], 64)
	if singleton[1] == "*" {
		return num * num_2
	}
	return num / num_2
}
func strToSlice(expression string) []string {
	var elements []string
	var previous string
	for _, s := range expression {
		if previous == "" || isArithmeticSign([]rune(previous)[0]) != isArithmeticSign(s) ||
			isArithmeticSign(s) {
			elements = append(elements, string(s))
			previous = string(s)
		} else {
			elements[len(elements)-1] += string(s)
			previous = elements[len(elements)-1]
		}
	}
	return elements
}
func calculate(elements []string, logger *slog.Logger) float64 {
	var subExpressions []SubExpression
	var isStarted bool
	var otherParentheses int
	for n, s := range elements {
		if isStarted {
			if s == ")" {
				if otherParentheses > 0 {
					otherParentheses--
					subExpressions[len(subExpressions)-1].Expression = append(
						subExpressions[len(subExpressions)-1].Expression,
						s,
					)
					continue
				}
				isStarted = false
				subExpressions[len(subExpressions)-1].End = n + 1
			} else if s == "(" {
				otherParentheses++
				subExpressions[len(subExpressions)-1].Expression = append(subExpressions[len(
					subExpressions)-1].Expression, s)
			} else {
				subExpressions[len(subExpressions)-1].Expression = append(subExpressions[len(
					subExpressions)-1].Expression, s)
			}
		} else if s == "(" {
			isStarted = true
			subExpressions = append(subExpressions, SubExpression{Start: n, End: n + 1, Expression: []string{}})
		}
	}
	if len(subExpressions) > 0 {
		logger.Debug(fmt.Sprintf("%+v", subExpressions))
		subExpression := subExpressions[0]
		var offset int
		newElements := append([]string{}, elements[:subExpression.Start]...)
		newElements = append(
			newElements,
			fmt.Sprintf("%f", calculate(subExpression.Expression, logger)),
		)
		newElements = append(newElements, elements[subExpression.End:]...)
		offset = len(subExpressions[0].Expression) + 1
		for _, subExpression := range subExpressions[1:] {
			newElements = append(newElements[:subExpression.Start-offset],
				fmt.Sprintf("%f", calculate(subExpression.Expression, logger)))
			offset += len(subExpression.Expression) + 1
			newElements = append(newElements, elements[subExpression.End:]...)
		}
		elements = newElements
	}
	for n, element := range elements {
		if element == "-" {
			elements[n] = "+"
			elements[n+1] = "-" + elements[n+1]
		}
	}
	var singletons [][]string
	var lastSingleton []string
	for _, element := range elements {
		if element == "+" {
			singletons = append(singletons, lastSingleton)
			lastSingleton = []string{}
		} else {
			lastSingleton = append(lastSingleton, element)
		}
	}
	singletons = append(singletons, lastSingleton)
	var answer float64
	for _, singleton := range singletons {
		if len(singleton) == 1 {
			num, _ := strconv.ParseFloat(singleton[0], 64)
			answer += num
		} else {
			subAnswer := multiplyOrDivide(singleton[:3])
			if len(singleton) > 3 {
				for n, element := range singleton[3 : len(singleton)-1] {
					if n%2 == 0 {
						subAnswer = multiplyOrDivide([]string{fmt.Sprintf("%f", subAnswer),
							element, singleton[n+4]})
					}
				}
			}
			answer += subAnswer
		}
	}
	return answer
}
func ValidateExpression(expression string) error {
	if len(expression) == 0 {
		return errors.New("expression is empty")
	}
	if isExactlyArithmeticSign(rune(expression[0])) {
		return &InvalidCharacterAtBeginningOrEndError{rune(expression[0])}
	}
	if isExactlyArithmeticSign(rune(expression[len(expression)-1])) {
		return &InvalidCharacterAtBeginningOrEndError{rune(expression[len(expression)-
			1])}
	}
	if len(expression) == 1 && unicode.IsDigit(rune(expression[0])) {
		return nil
	}
	for n, char := range expression[:len(expression)-2] {
		if isExactlyArithmeticSign(char) && isExactlyArithmeticSign(rune(expression[n+1])) ||
			char == '(' && isExactlyArithmeticSign(rune(expression[n+1])) ||
			expression[n+1] == ')' && isExactlyArithmeticSign(char) {
			return &InvalidCharacterCombinationError{string(char) + string(expression[n+1])}
		}
	}
	var fCounter, sCounter int
	for _, char := range expression {
		if !(isArithmeticSign(char) || unicode.IsDigit(char)) {
			return &InvalidCharacter{char}
		}
		if char == '(' {
			fCounter++
		} else if char == ')' {
			if fCounter == 0 {
				return &InvalidUseOfParentheses{"the expression contains a closing " +
					"parenthesis without a preceding opening parenthesis"}
			} else {
				sCounter++
			}
		}
	}
	if fCounter != sCounter {
		return &InvalidUseOfParentheses{
			"the number of opening parentheses does not equal the number of closing " +
				"parentheses",
		}
	}
	return nil
}
func Calc(expression string, logger *slog.Logger) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = fmt.Errorf("calculation error: %v", r)
		}
	}()
	err = ValidateExpression(expression)
	if err != nil {
		return 0, err
	}
	if len(expression) == 1 {
		return strconv.ParseFloat(expression, 64)
	}
	result = calculate(strToSlice(expression), logger)
	return result, nil
}
