package models

// Herb 对应 Sequelize Herb 模型

type Herb struct {
	ID     uint `gorm:"primaryKey;column:id"`
	UserID uint `gorm:"column:user_id"`

	HerbID  string `gorm:"column:herb_id"`
	Name    string `gorm:"column:name"`
	Count   int    `gorm:"column:count"`
	Quality string `gorm:"column:quality"` // 品质: common, uncommon, rare, epic, legendary
}

func (Herb) TableName() string {
	return "herbs"
}
