package models

import "gorm.io/datatypes"

// Equipment 对应 Sequelize Equipment 模型

type Equipment struct {
	ID     string `gorm:"type:uuid;primaryKey;column:id"`
	UserID uint   `gorm:"column:userId"`

	EquipmentID string `gorm:"column:equipmentId"`
	Name        string `gorm:"column:name"`
	Type        string `gorm:"column:type"`

	Slot      *string        `gorm:"column:slot"`
	EquipType *string        `gorm:"column:equipType"`
	Details   datatypes.JSON `gorm:"column:details"`
	Stats     datatypes.JSON `gorm:"column:stats"`
	// 额外属性，与 Node Equipment.extraAttributes 对应
	ExtraAttributes datatypes.JSON `gorm:"column:extraAttributes"`

	Quality      string `gorm:"column:quality"`
	EnhanceLevel int    `gorm:"column:enhanceLevel"`
	Equipped     bool   `gorm:"column:equipped"`

	Description   *string `gorm:"column:description"`
	RequiredRealm int     `gorm:"column:requiredRealm"`
	Level         int     `gorm:"column:level"`
}

func (Equipment) TableName() string {
	return "Equipment"
}
