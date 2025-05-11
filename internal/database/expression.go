package database

import (
	"calculator/internal/global"
)

type Expression struct {
	ID     string `gorm:"primaryKey"`
	UserID uint
	User   User    `gorm:"constraint:OnDelete:CASCADE"`
	Data   string  `gorm:"not null"`
	Status string  `gorm:"not null"`
	Result float64 `gorm:"not null"`
}

func (e *Expression) ToDTO() global.ExpressionDTO {
	return global.ExpressionDTO{ID: e.ID, UserID: e.UserID, Data: e.Data, Status: e.Status, Result: e.Result}
}

func CreateExpression(expr *Expression) error {
	return DB.Create(expr).Error
}

func GetExpressionByID(id string) (*global.ExpressionDTO, error) {
	var expr Expression
	err := DB.First(&expr, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	dto := expr.ToDTO()
	return &dto, nil
}

func UpdateExpressionStatus(id string, status string) error {
	return DB.Model(&Expression{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateExpressionResult(id string, result float64) error {
	return DB.Model(&Expression{}).Where("id = ?", id).Update("result", result).Error
}

func GetAllExpressions(userID uint) ([]Expression, error) {
	var expressions []Expression
	err := DB.Find(&expressions, "user_id = ?", userID).Error
	return expressions, err
}
