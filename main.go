package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type SubExpression struct {
	Start      int
	End        int
	Expression []string
}
type RequestData struct {
	Expression string `json:"expression"`
}
type ResponseData struct {
	Result string `json:"result"`
}
type ErrorData struct {
	Error string `json:"error"`
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
		if previous == "" || isArithmeticSign([]rune(previous)[0]) !=
			isArithmeticSign(s) || isArithmeticSign(s) {
			elements = append(elements, string(s))
			previous = string(s)
		} else {
			elements[len(elements)-1] += string(s)
			previous = elements[len(elements)-1]
		}
	}
	for n, element := range elements[:len(elements)-1] {
		q := []rune(element)[0]
		w := []rune(elements[n+1])[0]
		if len(element) == len(elements[n+1]) && len(element) == 1 &&
			isArithmeticSign(q) == isArithmeticSign(w) &&
			(isExactlyArithmeticSign(q) == isExactlyArithmeticSign(w) ||
				q == '(' && isExactlyArithmeticSign(w) ||
				isExactlyArithmeticSign(q) && w == ')') {
			panic("Panic!")
		}
	}
	return elements
}
func calculate(elements []string) float64 {
	var subExpressions []SubExpression
	var isStarted bool
	var otherParentheses int
	for n, s := range elements {
		if isStarted {
			if s == ")" {
				if otherParentheses > 0 {
					otherParentheses--
					subExpressions[len(subExpressions)-1].Expression = append(
						subExpressions[len(subExpressions)-1].Expression, s)
					continue
				}
				isStarted = false
				subExpressions[len(subExpressions)-1].End = n + 1
			} else if s == "(" {
				otherParentheses++
				subExpressions[len(subExpressions)-1].Expression = append(
					subExpressions[len(subExpressions)-1].Expression, s)
			} else {
				subExpressions[len(subExpressions)-1].Expression = append(
					subExpressions[len(subExpressions)-1].Expression, s)
			}
		} else if s == "(" {
			isStarted = true
			subExpressions = append(subExpressions, SubExpression{Start: n,
				End: n + 1, Expression: []string{}})
		}
	}
	if len(subExpressions) > 0 {
		subExpression := subExpressions[0]
		var offset int
		newElements := append([]string{}, elements[:subExpression.Start]...)
		newElements = append(newElements, fmt.Sprintf("%f", calculate(
			subExpression.Expression)))
		newElements = append(newElements, elements[subExpression.End:]...)
		offset = len(subExpressions[0].Expression) + 1
		for _, subExpression := range subExpressions[1:] {
			newElements = append(newElements[:subExpression.Start-offset],
				fmt.Sprintf("%f", calculate(subExpression.Expression)))
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
						subAnswer = multiplyOrDivide([]string{fmt.Sprintf(
							"%f", subAnswer), element, singleton[n+4]})
					}
				}
			}
			answer += subAnswer
		}
	}
	return answer
}
func ValidateExpression(expression string) error {
	if isExactlyArithmeticSign(rune(expression[0])) {
		return &InvalidCharacterAtBeginningOrEndError{rune(
			expression[0])}
	}
	if isExactlyArithmeticSign(rune(expression[len(expression)-1])) {
		return &InvalidCharacterAtBeginningOrEndError{rune(
			expression[len(expression)-1])}
	}
	for n, char := range expression[:len(expression)-2] {
		if isExactlyArithmeticSign(char) && isExactlyArithmeticSign(rune(
			expression[n+1])) || char == '(' && isExactlyArithmeticSign(rune(
			expression[n+1])) || expression[n+1] == ')' &&
			isExactlyArithmeticSign(char) {
			return &InvalidCharacterCombinationError{string(char) +
				string(expression[n+1])}
		}
	}
	var fCounter, sCounter int
	for _, char := range expression {
		if char == '(' {
			fCounter++
		} else if char == ')' {
			if fCounter == 0 {
				return &InvalidUseOfParentheses{"the expression " +
					"contains a closing parenthesis without a preceding " +
					"opening parenthesis"}
			} else {
				sCounter++
			}
		}
	}
	if fCounter != sCounter {
		return &InvalidUseOfParentheses{"the number of opening " +
			"parentheses does not equal the number of closing parentheses"}
	}
	return nil
}
func Calc(expression string) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = 0
			err = fmt.Errorf("calculation error: %v", r)
		}
	}()
	result = calculate(strToSlice(expression))
	return result, nil
}
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorData{Error: "Only POST method is allowed"})
		return
	}
	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorData{Error: "Invalid JSON"})
		return
	}
	err = ValidateExpression(data.Expression)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorData{Error: err.Error()})
		return
	}
	result, err := Calc(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: "There's an error in the expression"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseData{Result: strconv.FormatFloat(result,
		'f', -1, 64)})

}
func main() {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/calculate", http.HandlerFunc(HelloHandler))
	fmt.Println("The only endpoint is available at " +
		"http://localhost:8080/api/v1/calculate")
	http.ListenAndServe(":8080", mux)
}
