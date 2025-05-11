package database_test

import (
	"calculator/internal/database"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := database.DB.AutoMigrate(&database.User{}, &database.Expression{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}
}

func TestCreateAndGetUser(t *testing.T) {
	setupTestDB(t)
	login := "alice"
	user := &database.User{Login: login}
	user.SetPassword("hashpwd")
	if err := database.CreateUser(user); err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	u, err := database.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("GetUserByLogin error: %v", err)
	}
	if u.Login != login {
		t.Errorf("Login = %q, want %q", u.Login, login)
	}
	if u.GetPassword() != "hashpwd" {
		t.Errorf("Password = %q, want %q", u.GetPassword(), "hashpwd")
	}
	u2, err := database.GetUserByID(u.ID)
	if err != nil {
		t.Fatalf("GetUserByID error: %v", err)
	}
	if u2.ID != u.ID {
		t.Errorf("GetUserByID returned ID %d, want %d", u2.ID, u.ID)
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	setupTestDB(t)
	user := &database.User{Login: "bob"}
	user.SetPassword("h1")
	if err := database.CreateUser(user); err != nil {
		t.Fatalf("first CreateUser error: %v", err)
	}
	err := database.CreateUser(&database.User{Login: "bob", Password: "h2"})
	if err == nil {
		t.Fatal("expected duplicate key error, got nil")
	}
}

func TestCreateAndGetExpression(t *testing.T) {
	setupTestDB(t)
	expr := &database.Expression{
		ID:     "expr1",
		UserID: 1,
		Data:   "2+2",
		Status: "pending",
		Result: 0,
	}
	if err := database.CreateExpression(expr); err != nil {
		t.Fatalf("CreateExpression error: %v", err)
	}
	dto, err := database.GetExpressionByID("expr1")
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if dto.ID != expr.ID || dto.UserID != expr.UserID || dto.Data != expr.Data || dto.Status != expr.Status {
		t.Errorf("Got DTO %+v, want matching fields from %+v", dto, expr)
	}
	if dto.Result != 0 {
		t.Errorf("initial Result = %v, want 0", dto.Result)
	}
}

func TestUpdateExpressionStatusAndResult(t *testing.T) {
	setupTestDB(t)
	expr := &database.Expression{ID: "e2", UserID: 2, Data: "3*3", Status: "pending", Result: 0}
	database.CreateExpression(expr)
	if err := database.UpdateExpressionStatus("e2", "done"); err != nil {
		t.Fatalf("UpdateExpressionStatus error: %v", err)
	}
	if err := database.UpdateExpressionResult("e2", 9); err != nil {
		t.Fatalf("UpdateExpressionResult error: %v", err)
	}
	dto, err := database.GetExpressionByID("e2")
	if err != nil {
		t.Fatalf("GetExpressionByID error: %v", err)
	}
	if dto.Status != "done" {
		t.Errorf("Status = %q, want %q", dto.Status, "done")
	}
	if dto.Result != 9 {
		t.Errorf("Result = %v, want 9", dto.Result)
	}
}

func TestGetAllExpressions(t *testing.T) {
	setupTestDB(t)
	database.CreateExpression(&database.Expression{ID: "a1", UserID: 1, Data: "x", Status: "s", Result: 0})
	database.CreateExpression(&database.Expression{ID: "a2", UserID: 1, Data: "y", Status: "s", Result: 0})
	database.CreateExpression(&database.Expression{ID: "b1", UserID: 2, Data: "z", Status: "s", Result: 0})

	exprs, err := database.GetAllExpressions(1)
	if err != nil {
		t.Fatalf("GetAllExpressions error: %v", err)
	}
	if len(exprs) != 2 {
		t.Errorf("len = %d, want 2", len(exprs))
	}
	ids := map[string]bool{exprs[0].ID: true, exprs[1].ID: true}
	if !ids["a1"] || !ids["a2"] {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestDBStoreAdapter(t *testing.T) {
	setupTestDB(t)
	expr := &database.Expression{ID: "d1", UserID: 3, Data: "1+1", Status: "p", Result: 0}
	database.CreateExpression(expr)
	store := database.DBStore{}
	if err := store.UpdateExpressionStatus("d1", "ok"); err != nil {
		t.Fatalf("DBStore.UpdateExpressionStatus error: %v", err)
	}
	if err := store.UpdateExpressionResult("d1", 2); err != nil {
		t.Fatalf("DBStore.UpdateExpressionResult error: %v", err)
	}
	dto, err := store.GetExpressionByID("d1")
	if err != nil {
		t.Fatalf("DBStore.GetExpressionByID error: %v", err)
	}
	if dto.Status != "ok" || dto.Result != 2 {
		t.Errorf("Adapter DTO = %+v, want status 'ok' and result 2", dto)
	}
}
