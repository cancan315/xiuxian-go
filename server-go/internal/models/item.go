package models

import "gorm.io/datatypes"

// Item 对应 Sequelize Item 模型

type Item struct {
	ID     string `gorm:"type:uuid;primaryKey;column:id"`
	UserID uint   `gorm:"column:userId"`

	ItemID string `gorm:"column:itemId"`
	Name   string `gorm:"column:name"`
	Type   string `gorm:"column:type"`

	Details  datatypes.JSON `gorm:"column:details"`
	Slot     *string        `gorm:"column:slot"`
	Stats    datatypes.JSON `gorm:"column:stats"`
	Quality  string         `gorm:"column:quality"`
	Equipped bool           `gorm:"column:equipped"`
}

func (Item) TableName() string {
	return "Items"
}
