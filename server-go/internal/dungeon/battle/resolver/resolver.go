package resolver

import (
	"math"
	"math/rand"

	"xiuxian/server-go/internal/dungeon/battle"
	"xiuxian/server-go/internal/dungeon/battle/formula"
)

// TakeDamage 被伤害 - 对应前端 takeDamage 方法
func TakeDamage(defender *battle.CombatStats, currentHealth float64, incomingDamage float64, source *battle.CombatStats) *battle.TakeDamageResult {
	result := &battle.TakeDamageResult{}

	// 闪避判定：actualDodgeRate = MAX(0, MIN(0.8, dodgeRate - (source ? source.dodgeResist : 0)))
	var actualDodgeRate float64
	if source != nil {
		actualDodgeRate = math.Max(0, math.Min(0.8, defender.DodgeRate-source.DodgeResist))
	} else {
		actualDodgeRate = math.Max(0, math.Min(0.8, defender.DodgeRate))
	}

	// 闪避成功
	if rand.Float64() < actualDodgeRate {
		result.Dodged = true
		result.Damage = 0
		result.CurrentHealth = currentHealth
		return result
	}

	// 计算实际伤害
	reducedDamage := formula.CalculateDamageReduction(defender, incomingDamage, source)
	newHealth := math.Max(0, currentHealth-reducedDamage)

	// 反击判定：finalCounterRate = MAX(0, MIN(0.8, counterRate - source.counterResist))
	if source != nil {
		finalCounterRate := math.Max(0, math.Min(0.8, defender.CounterRate-source.CounterResist))
		if rand.Float64() < finalCounterRate {
			result.IsCounter = true
		}
	}

	result.Dodged = false
	result.Damage = reducedDamage
	result.CurrentHealth = newHealth
	result.IsDead = newHealth <= 0

	return result
}

// ApplyBuffEffectsToStats 将已选增益效果应用到战斗属性
func ApplyBuffEffectsToStats(stats *battle.CombatStats, buffEffects map[string]interface{}) *battle.CombatStats {
	for key, value := range buffEffects {
		if val, ok := value.(float64); ok {
			switch key {
			case "health":
				stats.Health *= (1 + val)
				stats.MaxHealth *= (1 + val)
			case "finalDamageBoost", "damage":
				stats.Damage *= (1 + val)
			case "defense":
				stats.Defense *= (1 + val)
			case "speed":
				stats.Speed *= (1 + val)
			case "critRate":
				stats.CritRate += val
			case "dodgeRate":
				stats.DodgeRate += val
			case "comboRate":
				stats.ComboRate += val
			case "vampireRate":
				stats.VampireRate += val
			case "stunRate":
				stats.StunRate += val
			case "critDamageBoost":
				stats.CritDamageBoost += val
			case "finalDamageReduce":
				stats.FinalDamageReduce += val
			case "combatBoost":
				stats.CombatBoost += val
			case "resistanceBoost":
				stats.ResistanceBoost += val
			}
		}
	}
	return stats
}
