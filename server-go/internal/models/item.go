package models

import "gorm.io/datatypes"

// Item 对应 Sequelize Item 模型

type Item struct {
	ID     string `gorm:"type:uuid;primaryKey;column:id"`
	UserID uint   `gorm:"column:user_id"`

	ItemID string `gorm:"column:item_id"`
	Name   string `gorm:"column:name"`
	Type   string `gorm:"column:type"`

	Details   datatypes.JSON `gorm:"column:details"`
	Slot      *string        `gorm:"column:slot"`    // 添加 Slot 字段，装备槽位
	Stats     datatypes.JSON `gorm:"column:stats"`   // 添加 Stats 字段，装备属性
	Quality   string         `gorm:"column:quality"` // 添加 Quality 字段，装备品质
	Equipped  bool           `gorm:"column:equipped"`
	EquipType *string        `gorm:"column:equip_type"` // 添加 EquipType 字段，装备类型
}

func (Item) TableName() string {
	return "items"
}
