package models

import (
	"time"

	"gorm.io/datatypes"
)

// User 对应原 Node.js Sequelize User 模型的主要字段
// 目前主要用于认证和玩家基础属性、灵力等

type User struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Username string `gorm:"size:255;uniqueIndex;not null;column:username" json:"username"`
	Password string `gorm:"size:255;not null;column:password" json:"password"`

	// Player basic info
	PlayerName       string  `gorm:"column:player_name" json:"playerName"`
	Level            int     `gorm:"column:level" json:"level"`
	Realm            string  `gorm:"column:realm" json:"realm"`
	Cultivation      float64 `gorm:"column:cultivation" json:"cultivation"`
	MaxCultivation   float64 `gorm:"column:max_cultivation" json:"maxCultivation"`
	Spirit           float64 `gorm:"column:spirit" json:"spirit"`
	SpiritStones     int     `gorm:"column:spirit_stones" json:"spiritStones"`
	ReinforceStones  int     `gorm:"column:reinforce_stones" json:"reinforceStones"`
	RefinementStones int     `gorm:"column:refinement_stones" json:"refinementStones"`
	PetEssence       int     `gorm:"column:pet_essence" json:"petEssence"`

	// 战斗相关属性（JSON 存储，与 Node User 模型字段对应）
	BaseAttributes    datatypes.JSON `gorm:"column:base_attributes" json:"baseAttributes"`
	CombatAttributes  datatypes.JSON `gorm:"column:combat_attributes" json:"combatAttributes"`
	CombatResistance  datatypes.JSON `gorm:"column:combat_resistance" json:"combatResistance"`
	SpecialAttributes datatypes.JSON `gorm:"column:special_attributes" json:"specialAttributes"`

	// 灵力自动增长相关字段
	LastSpiritGainTime time.Time `gorm:"column:last_spirit_gain_time" json:"lastSpiritGainTime"`

	// 其他字段先不全部展开，后续需要再补充

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName 显式指定与 Sequelize 使用的表名一致
func (User) TableName() string {
	return "users"
}
