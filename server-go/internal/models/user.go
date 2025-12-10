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
	PlayerName       string  `gorm:"column:playerName"`
	Level            int     `gorm:"column:level"`
	Realm            string  `gorm:"column:realm"`
	Cultivation      float64 `gorm:"column:cultivation"`
	MaxCultivation   float64 `gorm:"column:maxCultivation"`
	Spirit           float64 `gorm:"column:spirit"`
	SpiritStones     int     `gorm:"column:spiritStones"`
	ReinforceStones  int     `gorm:"column:reinforceStones"`
	RefinementStones int     `gorm:"column:refinementStones"`
	PetEssence       int     `gorm:"column:petEssence"`

	// 战斗相关属性（JSON 存储，与 Node User 模型字段对应）
	BaseAttributes    datatypes.JSON `gorm:"column:baseAttributes"`
	CombatAttributes  datatypes.JSON `gorm:"column:combatAttributes"`
	CombatResistance  datatypes.JSON `gorm:"column:combatResistance"`
	SpecialAttributes datatypes.JSON `gorm:"column:specialAttributes"`

	// 其他字段先不全部展开，后续需要再补充

	CreatedAt time.Time `gorm:"column:createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt"`
}

// TableName 显式指定与 Sequelize 使用的表名一致
func (User) TableName() string {
	return "Users"
}
