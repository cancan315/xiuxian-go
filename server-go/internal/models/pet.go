package models

import "gorm.io/datatypes"

// Pet 对应 Sequelize Pet 模型

type Pet struct {
	ID     string `gorm:"type:uuid;primaryKey;column:id"`
	UserID uint   `gorm:"column:user_id"`

	PetID string `gorm:"column:pet_id"`
	Name  string `gorm:"column:name"`
	Type  string `gorm:"column:type"`

	Rarity string `gorm:"column:rarity"`
	Level  int    `gorm:"column:level"`
	Star   int    `gorm:"column:star"`

	Experience    int `gorm:"column:experience"`
	MaxExperience int `gorm:"column:max_experience"`

	Quality          datatypes.JSON `gorm:"column:quality"`
	CombatAttributes datatypes.JSON `gorm:"column:combat_attributes"`

	IsActive bool `gorm:"column:is_active"`

	AttackBonus  float64 `gorm:"column:attack_bonus"`
	DefenseBonus float64 `gorm:"column:defense_bonus"`
	HealthBonus  float64 `gorm:"column:health_bonus"`
}

func (Pet) TableName() string {
	return "pets"
}
