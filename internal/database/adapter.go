package database

import "calculator/internal/global"

type DBStore struct{}

func (s DBStore) UpdateExpressionStatus(id string, status string) error {
	return UpdateExpressionStatus(id, status)
}

func (s DBStore) GetExpressionByID(id string) (*global.ExpressionDTO, error) {
	return GetExpressionByID(id)
}

func (s DBStore) UpdateExpressionResult(id string, res float64) error {
	return UpdateExpressionResult(id, res)
}
