package models

import "gorm.io/datatypes"

// Pill 对应 Sequelize Pill 模型

type Pill struct {
	ID     uint `gorm:"primaryKey;column:id"`
	UserID uint `gorm:"column:user_id"`

	PillID string `gorm:"column:pill_id"`
	Name   string `gorm:"column:name"`

	Description string         `gorm:"column:description"`
	Effect      datatypes.JSON `gorm:"column:effect"`
}

func (Pill) TableName() string {
	return "pills"
}
