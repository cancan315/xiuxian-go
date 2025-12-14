package models

// PillFragment 丹方残页模型
type PillFragment struct {
	ID       uint   `gorm:"primaryKey;column:id"`
	UserID   uint   `gorm:"column:user_id"`
	RecipeID string `gorm:"column:recipe_id"`
	Count    int    `gorm:"column:count"`
}

func (PillFragment) TableName() string {
	return "pill_fragments"
}