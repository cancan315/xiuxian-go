package models

import "gorm.io/datatypes"

// Pet 对应 Sequelize Pet 模型

type Pet struct {
	ID     string `gorm:"type:uuid;primaryKey;column:id"`
	UserID uint   `gorm:"column:userId"`

	PetID string `gorm:"column:petId"`
	Name  string `gorm:"column:name"`
	Type  string `gorm:"column:type"`

	Rarity string `gorm:"column:rarity"`
	Level  int    `gorm:"column:level"`
	Star   int    `gorm:"column:star"`

	Experience    int `gorm:"column:experience"`
	MaxExperience int `gorm:"column:maxExperience"`

	Quality          datatypes.JSON `gorm:"column:quality"`
	CombatAttributes datatypes.JSON `gorm:"column:combatAttributes"`

	IsActive bool `gorm:"column:isActive"`

	AttackBonus  float64 `gorm:"column:attackBonus"`
	DefenseBonus float64 `gorm:"column:defenseBonus"`
	HealthBonus  float64 `gorm:"column:healthBonus"`
}

func (Pet) TableName() string {
	return "Pets"
}
