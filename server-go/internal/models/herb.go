package models

// Herb 对应 Sequelize Herb 模型

type Herb struct {
	ID     uint `gorm:"primaryKey;column:id"`
	UserID uint `gorm:"column:userId"`

	HerbID string `gorm:"column:herbId"`
	Name   string `gorm:"column:name"`
	Count  int    `gorm:"column:count"`
}

func (Herb) TableName() string {
	return "Herbs"
}
