package handler

import (
	"bytes"
	"calculator/http/server/middleware"
	"calculator/internal/application/auth"
	"calculator/internal/database"
	"calculator/pkg/calculator"
	"calculator/pkg/global"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type decorator func(http.Handler) http.Handler

type requestData struct {
	Expression string `json:"expression"`
}

type responseData struct {
	Result float64 `json:"result"`
}

type errorData struct {
	Error string `json:"error"`
}

type infoData struct {
	Info string `json:"info"`
}

type tokenData struct {
	Info  string `json:"info"`
	Token string `json:"token"`
}

type responseWriterWrapper struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

type idResponse struct {
	ID string `json:"id"`
}

type expressionResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type expressionsResponse struct {
	Expressions []expressionResponse `json:"expressions"`
}

type SolvedTaskResponse struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func New(ctx context.Context) (http.Handler, error) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/api/v1/calculate", middleware.JWTMiddleware()(calculatorAPIHandler))
	serveMux.HandleFunc("/api/v1/expressions", middleware.JWTMiddleware()(expressionsHandler))
	serveMux.HandleFunc("/api/v1/expressions/", middleware.JWTMiddleware()(expressionHandler))
	serveMux.HandleFunc("/api/v1/register", registerHandler)
	serveMux.HandleFunc("/api/v1/login", loginHandler)
	serveMux.HandleFunc("/internal/task", taskHandler)
	return serveMux, nil
}

func Decorate(next http.Handler, ds ...decorator) http.Handler {
	decorated := next
	for d := len(ds) - 1; d >= 0; d-- {
		decorated = ds[d](decorated)
	}
	return decorated
}

func calculatorAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only POST method is allowed"})
		return
	}
	var data requestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "invalid JSON"})
		return
	}
	if data.Expression == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(errorData{Error: "no expression provided"})
		return
	}
	expressionID := uuid.New().String()
	userIDRaw := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorData{Error: "user ID not found"})
		return
	}
	err = database.CreateExpression(
		&database.Expression{
			ID:     expressionID,
			UserID: userID,
			Data:   data.Expression,
			Status: "pending",
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorData{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	go calculator.Calc(expressionID)
	json.NewEncoder(w).Encode(idResponse{expressionID})
}

func expressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only GET method is allowed"})
		return
	}
	userIDRaw := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorData{Error: "user ID not found"})
		return
	}
	expressions, err := database.GetAllExpressions(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorData{Error: err.Error()})
		return
	}
	expressionsResponse := expressionsResponse{}
	for _, expression := range expressions {
		expressionsResponse.Expressions = append(
			expressionsResponse.Expressions,
			expressionResponse{expression.ID, expression.Status, expression.Result},
		)
	}
	json.NewEncoder(w).Encode(expressionsResponse)
}

func expressionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only GET method is allowed"})
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "ID not provided"})
		return
	}
	expression, err := database.GetExpressionByID(parts[4])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorData{Error: "there is no such expression"})
		return
	}
	userIDRaw := r.Context().Value(middleware.UserIDKey)
	userID, ok := userIDRaw.(uint)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorData{Error: "user ID not found"})
		return
	}
	if expression.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(errorData{Error: "you do not have access to this information"})
		return
	}
	json.NewEncoder(w).Encode(expressionResponse{parts[4], expression.Status, expression.Result})
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		var count int
		global.TasksMap.Range(func(key, value interface{}) bool {
			count++
			k := key.(string)
			v := value.(*global.Task)
			json.NewEncoder(w).Encode(v)
			global.TasksMap.Delete(k)
			return false
		})
		if count == 0 {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if r.Method == http.MethodPost {
		var data SolvedTaskResponse
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorData{Error: "invalid JSON"})
			return
		}
		futureInterface, ok := global.FuturesMap.Load(data.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(errorData{Error: "there is no such task"})
			return
		}
		future, ok := futureInterface.(*global.Future)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errorData{Error: "server error"})
			return
		}
		future.SetResult(data.Result)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only GET and POST methods are allowed"})
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only POST method is allowed"})
		return
	}
	var creds credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "invalid or unknown fields"})
		return
	}
	if creds.Login == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "login and password are required"})
		return
	}
	if err := auth.Register(creds.Login, creds.Password); err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorData{Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(infoData{Info: "OK"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorData{Error: "only POST method is allowed"})
		return
	}
	var creds credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "invalid or unknown fields"})
		return
	}
	if creds.Login == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorData{Error: "login and password are required"})
		return
	}
	token, err := auth.Login(creds.Login, creds.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorData{Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(tokenData{Info: "OK", Token: token})
}
