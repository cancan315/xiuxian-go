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
	PlayerName       string  `gorm:"column:player_name" json:"playerName"`                      // 玩家名称
	NameChangeCount  int     `gorm:"column:name_change_count;default:0" json:"nameChangeCount"` // 道号修改次数
	Level            int     `gorm:"column:level" json:"level"`                                 // 等级
	Realm            string  `gorm:"column:realm" json:"realm"`                                 // 境界
	Cultivation      float64 `gorm:"column:cultivation" json:"cultivation"`                     // 修为
	MaxCultivation   float64 `gorm:"column:max_cultivation" json:"maxCultivation"`              // 最大修为
	Spirit           float64 `gorm:"column:spirit" json:"spirit"`                               // 灵力
	SpiritStones     int     `gorm:"column:spirit_stones" json:"spiritStones"`                  // 灵石
	ReinforceStones  int     `gorm:"column:reinforce_stones" json:"reinforceStones"`            // 强化石
	RefinementStones int     `gorm:"column:refinement_stones" json:"refinementStones"`          // 洗练石
	PetEssence       int     `gorm:"column:pet_essence" json:"petEssence"`                      // 宠物精华

	// 战斗相关属性（JSON 存储，与 Node User 模型字段对应）
	BaseAttributes    datatypes.JSON `gorm:"column:base_attributes" json:"baseAttributes"`       // 基础属性
	CombatAttributes  datatypes.JSON `gorm:"column:combat_attributes" json:"combatAttributes"`   // 战斗属性
	CombatResistance  datatypes.JSON `gorm:"column:combat_resistance" json:"combatResistance"`   // 战斗抗性
	SpecialAttributes datatypes.JSON `gorm:"column:special_attributes" json:"specialAttributes"` // 特殊属性

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
