package server

import (
	"calculator/pkg/calculator"
	"encoding/json"
	"net/http"
	"strconv"
)

type RequestData struct {
	Expression string `json:"expression"`
}
type ResponseData struct {
	Result string `json:"result"`
}
type ErrorData struct {
	Error string `json:"error"`
}

func CalculatorHandler(w http.ResponseWriter, r *http.Request) {
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
	err = calculator.ValidateExpression(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: err.Error()})
		return
	}
	result, err := calculator.Calc(data.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorData{Error: "There's an error in the expression"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseData{Result: strconv.FormatFloat(result,
		'f', -1, 64)})
}
