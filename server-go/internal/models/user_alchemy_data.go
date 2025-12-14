package models

// UserAlchemyData 用户炼丹统计数据
type UserAlchemyData struct {
	ID              uint    `gorm:"primaryKey;column:id"`
	UserID          uint    `gorm:"column:user_id;uniqueIndex"`
	RecipesUnlocked string  `gorm:"column:recipes_unlocked"` // JSON格式存储已解锁的丹方ID列表
	PillsCrafted    int     `gorm:"column:pills_crafted"`    // 总炼制次数
	PillsConsumed   int     `gorm:"column:pills_consumed"`   // 总服用次数
	AlchemyLevel    int     `gorm:"column:alchemy_level"`    // 炼丹等级
	AlchemyRate     float64 `gorm:"column:alchemy_rate"`     // 炼丹加成率
}

func (UserAlchemyData) TableName() string {
	return "user_alchemy_data"
}
