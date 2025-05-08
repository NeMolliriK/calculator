package database

import (
	"calculator/pkg/global"

	"gorm.io/gorm"
)

type Expression struct {
	ID     string `gorm:"primaryKey"`
	Data   string
	Status string
	Result float64
}

func NewExpressionFromDTO(dto global.ExpressionDTO) *Expression {
	return &Expression{dto.ID, dto.Data, dto.Status, dto.Result}
}

func (e *Expression) ToDTO() global.ExpressionDTO {
	return global.ExpressionDTO{ID: e.ID, Data: e.Data, Status: e.Status, Result: e.Result}
}

func CreateExpression(db *gorm.DB, expr *Expression) error {
	return db.Create(expr).Error
}

func GetExpressionByID(db *gorm.DB, id string) (*Expression, error) {
	var expr Expression
	err := db.First(&expr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

func UpdateExpressionStatus(db *gorm.DB, id string, status string) error {
	return db.Model(&Expression{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateExpressionResult(db *gorm.DB, id string, result float64) error {
	return db.Model(&Expression{}).Where("id = ?", id).Update("result", result).Error
}

func GetAllExpressions(db *gorm.DB) ([]Expression, error) {
	var expressions []Expression
	err := db.Find(&expressions).Error
	return expressions, err
}
