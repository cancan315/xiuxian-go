package models

import (
	"time"

	"gorm.io/datatypes"
)

// User 对应原 Node.js Sequelize User 模型的主要字段
// 目前主要用于认证和玩家基础属性、灵力等

type User struct {
	ID       uint   `gorm:"primaryKey;column:id"`
	Username string `gorm:"size:255;uniqueIndex;not null;column:username"`
	Password string `gorm:"size:255;not null;column:password"`

	// Player basic info
	PlayerName       string  `gorm:"column:player_name"`
	Level            int     `gorm:"column:level"`
	Realm            string  `gorm:"column:realm"`
	Cultivation      float64 `gorm:"column:cultivation"`
	MaxCultivation   float64 `gorm:"column:max_cultivation"`
	Spirit           float64 `gorm:"column:spirit"`
	SpiritStones     int     `gorm:"column:spirit_stones"`
	ReinforceStones  int     `gorm:"column:reinforce_stones"`
	RefinementStones int     `gorm:"column:refinement_stones"`
	PetEssence       int     `gorm:"column:pet_essence"`

	// 战斗相关属性（JSON 存储，与 Node User 模型字段对应）
	BaseAttributes    datatypes.JSON `gorm:"column:base_attributes"`
	CombatAttributes  datatypes.JSON `gorm:"column:combat_attributes"`
	CombatResistance  datatypes.JSON `gorm:"column:combat_resistance"`
	SpecialAttributes datatypes.JSON `gorm:"column:special_attributes"`

	// 其他字段先不全部展开，后续需要再补充

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 显式指定与 Sequelize 使用的表名一致
func (User) TableName() string {
	return "users"
}
