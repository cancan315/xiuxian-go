package formula

import (
	"math"
	"math/rand"

	"xiuxian/server-go/internal/dungeon/battle"
)

// CalculateDamage 计算伤害 - 按战斗逻辑设定文档实现
// 基础伤害 = A.Damage - B.Defense
// 暴击触发概率 = critRate - critResist
// 连击触发概率 = comboRate - comboResist
func CalculateDamage(attacker *battle.CombatStats, defender *battle.CombatStats) *battle.DamageResult {
	result := &battle.DamageResult{}

	// 基础伤害 = A.Damage - B.Defense，最小为1
	baseDamage := math.Max(1, attacker.Damage-defender.Defense)
	result.BaseDamage = baseDamage
	result.TotalDamage = baseDamage

	// 暴击判定：critRate - critResist = 触发概率
	critChance := math.Max(0, math.Min(1, attacker.CritRate-defender.CritResist))
	if rand.Float64() < critChance {
		result.IsCrit = true
		result.CritDamage = baseDamage // 暴击伤害 = 基础伤害
		result.TotalDamage += result.CritDamage
	}

	// 连击判定：ComboRate - ComboResist = 触发概率
	comboChance := math.Max(0, math.Min(1, attacker.ComboRate-defender.ComboResist))
	if rand.Float64() < comboChance {
		result.IsCombo = true
		result.ComboDamage = baseDamage // 连击伤害 = 基础伤害
		result.TotalDamage += result.ComboDamage
	}

	// 吸血判定：vampireRate - vampireResist = 触发概率
	vampireChance := math.Max(0, math.Min(1, attacker.VampireRate-defender.VampireResist))
	if rand.Float64() < vampireChance {
		result.IsVampire = true
		result.VampireHeal = baseDamage * 0.2 // 吸血回复 = 基础伤害 * 20%
	}

	// 眩晕判定：stunRate - stunResist = 触发概率
	stunChance := math.Max(0, math.Min(1, attacker.StunRate-defender.StunResist))
	if rand.Float64() < stunChance {
		result.IsStun = true
	}

	return result
}

// CalculateDamageReduction 计算伤害减免
func CalculateDamageReduction(defender *battle.CombatStats, incomingDamage float64, attacker *battle.CombatStats) float64 {
	// 简化版：直接返回传入伤害（因为已经在 CalculateDamage 中减去了防御）
	return math.Max(0, incomingDamage)
}
