package database

import (
	"calculator/pkg/global"
)

type Expression struct {
	ID     string  `gorm:"primaryKey"`
	Data   string  `gorm:"not null"`
	Status string  `gorm:"not null"`
	Result float64 `gorm:"not null"`
}

func NewExpressionFromDTO(dto global.ExpressionDTO) *Expression {
	return &Expression{dto.ID, dto.Data, dto.Status, dto.Result}
}

func (e *Expression) ToDTO() global.ExpressionDTO {
	return global.ExpressionDTO{ID: e.ID, Data: e.Data, Status: e.Status, Result: e.Result}
}

func CreateExpression(expr *Expression) error {
	return DB.Create(expr).Error
}

func GetExpressionByID(id string) (*Expression, error) {
	var expr Expression
	err := DB.First(&expr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

func UpdateExpressionStatus(id string, status string) error {
	return DB.Model(&Expression{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateExpressionResult(id string, result float64) error {
	return DB.Model(&Expression{}).Where("id = ?", id).Update("result", result).Error
}

func GetAllExpressions() ([]Expression, error) {
	var expressions []Expression
	err := DB.Find(&expressions).Error
	return expressions, err
}
