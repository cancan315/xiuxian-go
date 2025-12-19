package player

import (
	"xiuxian/server-go/internal/models"
)

// AttributeManager 统一管理玩家属性变更
type AttributeManager struct {
	BaseAttrs    map[string]float64
	CombatAttrs  map[string]float64
	CombatRes    map[string]float64
	SpecialAttrs map[string]float64
}

// NewAttributeManager 创建新的属性管理器
func NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs map[string]float64) *AttributeManager {
	return &AttributeManager{
		BaseAttrs:    baseAttrs,
		CombatAttrs:  combatAttrs,
		CombatRes:    combatRes,
		SpecialAttrs: specialAttrs,
	}
}

// ApplyEquipmentStats 应用装备属性加成
// 顺序：基础属性 → 战斗属性 → 战斗抗性 → 特殊属性
func (am *AttributeManager) ApplyEquipmentStats(equipStats map[string]float64) {
	for stat, v := range equipStats {
		if v == 0 {
			continue
		}
		switch stat {
		case "attack", "health", "defense", "speed":
			am.BaseAttrs[stat] = am.BaseAttrs[stat] + v
		case "critRate", "comboRate", "counterRate", "stunRate", "dodgeRate", "vampireRate":
			am.CombatAttrs[stat] = am.CombatAttrs[stat] + v
		case "critResist", "comboResist", "counterResist", "stunResist", "dodgeResist", "vampireResist":
			am.CombatRes[stat] = am.CombatRes[stat] + v
		default:
			am.SpecialAttrs[stat] = am.SpecialAttrs[stat] + v
		}
	}
}

// RemoveEquipmentStats 移除装备属性加成（与应用逆向）
func (am *AttributeManager) RemoveEquipmentStats(equipStats map[string]float64) {
	for stat, v := range equipStats {
		if v == 0 {
			continue
		}
		var m map[string]float64
		switch stat {
		case "attack", "health", "defense", "speed":
			m = am.BaseAttrs
		case "critRate", "comboRate", "counterRate", "stunRate", "dodgeRate", "vampireRate":
			m = am.CombatAttrs
		case "critResist", "comboResist", "counterResist", "stunResist", "dodgeResist", "vampireResist":
			m = am.CombatRes
		default:
			m = am.SpecialAttrs
		}
		m[stat] = m[stat] - v
		if m[stat] < 0 {
			m[stat] = 0
		}
	}
}

// ApplyPetBonuses 应用灵宠加成
// 顺序：
// 1. 基础属性数值加成（combat attributes 中的 attack/defense/health/speed）
// 2. 战斗属性直接累加（有 0-1 上限）
// 3. 战斗抗性直接累加（有 0-1 上限）
// 4. 特殊属性累加
// 5. 基础属性百分比加成（必须最后，影响最终的基础属性值）
func (am *AttributeManager) ApplyPetBonuses(pet *models.Pet, petCombat map[string]float64) {
	// 第一步：基础属性数值加成（从 combatAttributes 中提取）
	if v, ok := petCombat["attack"]; ok {
		am.BaseAttrs["attack"] = am.BaseAttrs["attack"] + v
	}
	if v, ok := petCombat["defense"]; ok {
		am.BaseAttrs["defense"] = am.BaseAttrs["defense"] + v
	}
	if v, ok := petCombat["health"]; ok {
		am.BaseAttrs["health"] = am.BaseAttrs["health"] + v
	}
	if v, ok := petCombat["speed"]; ok {
		am.BaseAttrs["speed"] = am.BaseAttrs["speed"] + v
	}

	// 第二步：战斗属性（有 0-1 上限）
	am.applyRateAttribute(am.CombatAttrs, "critRate", petCombat)
	am.applyRateAttribute(am.CombatAttrs, "comboRate", petCombat)
	am.applyRateAttribute(am.CombatAttrs, "counterRate", petCombat)
	am.applyRateAttribute(am.CombatAttrs, "stunRate", petCombat)
	am.applyRateAttribute(am.CombatAttrs, "dodgeRate", petCombat)
	am.applyRateAttribute(am.CombatAttrs, "vampireRate", petCombat)

	// 第三步：战斗抗性（有 0-1 上限）
	am.applyRateAttribute(am.CombatRes, "critResist", petCombat)
	am.applyRateAttribute(am.CombatRes, "comboResist", petCombat)
	am.applyRateAttribute(am.CombatRes, "counterResist", petCombat)
	am.applyRateAttribute(am.CombatRes, "stunResist", petCombat)
	am.applyRateAttribute(am.CombatRes, "dodgeResist", petCombat)
	am.applyRateAttribute(am.CombatRes, "vampireResist", petCombat)

	// 第四步：特殊属性（无上限）
	for _, key := range []string{"healBoost", "critDamageBoost", "critDamageReduce", "finalDamageBoost", "finalDamageReduce", "combatBoost", "resistanceBoost"} {
		if v, ok := petCombat[key]; ok {
			am.SpecialAttrs[key] = am.SpecialAttrs[key] + v
		}
	}

	// 第五步：基础属性百分比加成（必须最后，作用于最终的基础属性值）
	if pet.AttackBonus != 0 {
		am.BaseAttrs["attack"] = am.BaseAttrs["attack"] * (1 + pet.AttackBonus)
	}
	if pet.DefenseBonus != 0 {
		am.BaseAttrs["defense"] = am.BaseAttrs["defense"] * (1 + pet.DefenseBonus)
	}
	if pet.HealthBonus != 0 {
		am.BaseAttrs["health"] = am.BaseAttrs["health"] * (1 + pet.HealthBonus)
	}
}

// RemovePetBonuses 移除灵宠加成（与应用逆向）
// 顺序逆向：先移除百分比，再移除其他属性
func (am *AttributeManager) RemovePetBonuses(pet *models.Pet, petCombat map[string]float64) {
	// 第一步：先移除基础属性百分比加成（顺序与应用完全相反）
	if pet.AttackBonus != 0 {
		if cur := am.BaseAttrs["attack"]; cur != 0 {
			am.BaseAttrs["attack"] = cur / (1 + pet.AttackBonus)
		}
	}
	if pet.DefenseBonus != 0 {
		if cur := am.BaseAttrs["defense"]; cur != 0 {
			am.BaseAttrs["defense"] = cur / (1 + pet.DefenseBonus)
		}
	}
	if pet.HealthBonus != 0 {
		if cur := am.BaseAttrs["health"]; cur != 0 {
			am.BaseAttrs["health"] = cur / (1 + pet.HealthBonus)
		}
	}

	// 第二步：移除基础属性数值加成
	if v, ok := petCombat["attack"]; ok && v != 0 {
		am.BaseAttrs["attack"] = am.BaseAttrs["attack"] - v
		if am.BaseAttrs["attack"] < 0 {
			am.BaseAttrs["attack"] = 0
		}
	}
	if v, ok := petCombat["defense"]; ok && v != 0 {
		am.BaseAttrs["defense"] = am.BaseAttrs["defense"] - v
		if am.BaseAttrs["defense"] < 0 {
			am.BaseAttrs["defense"] = 0
		}
	}
	if v, ok := petCombat["health"]; ok && v != 0 {
		am.BaseAttrs["health"] = am.BaseAttrs["health"] - v
		if am.BaseAttrs["health"] < 0 {
			am.BaseAttrs["health"] = 0
		}
	}
	if v, ok := petCombat["speed"]; ok && v != 0 {
		am.BaseAttrs["speed"] = am.BaseAttrs["speed"] - v
		if am.BaseAttrs["speed"] < 0 {
			am.BaseAttrs["speed"] = 0
		}
	}

	// 第三步：移除战斗属性（有 0-1 下限）
	am.removeRateAttribute(am.CombatAttrs, "critRate", petCombat)
	am.removeRateAttribute(am.CombatAttrs, "comboRate", petCombat)
	am.removeRateAttribute(am.CombatAttrs, "counterRate", petCombat)
	am.removeRateAttribute(am.CombatAttrs, "stunRate", petCombat)
	am.removeRateAttribute(am.CombatAttrs, "dodgeRate", petCombat)
	am.removeRateAttribute(am.CombatAttrs, "vampireRate", petCombat)

	// 第四步：移除战斗抗性（有 0-1 下限）
	am.removeRateAttribute(am.CombatRes, "critResist", petCombat)
	am.removeRateAttribute(am.CombatRes, "comboResist", petCombat)
	am.removeRateAttribute(am.CombatRes, "counterResist", petCombat)
	am.removeRateAttribute(am.CombatRes, "stunResist", petCombat)
	am.removeRateAttribute(am.CombatRes, "dodgeResist", petCombat)
	am.removeRateAttribute(am.CombatRes, "vampireResist", petCombat)

	// 第五步：移除特殊属性
	for _, key := range []string{"healBoost", "critDamageBoost", "critDamageReduce", "finalDamageBoost", "finalDamageReduce", "combatBoost", "resistanceBoost"} {
		if v, ok := petCombat[key]; ok && v != 0 {
			am.SpecialAttrs[key] = am.SpecialAttrs[key] - v
			if am.SpecialAttrs[key] < 0 {
				am.SpecialAttrs[key] = 0
			}
		}
	}
}

// applyRateAttribute 应用速率属性（战斗属性、战斗抗性）
// 这些属性有 0-1 的上限
func (am *AttributeManager) applyRateAttribute(attrs map[string]float64, key string, petCombat map[string]float64) {
	if v, ok := petCombat[key]; ok {
		attrs[key] = attrs[key] + v
		if attrs[key] > 1 {
			attrs[key] = 1
		}
	}
}

// removeRateAttribute 移除速率属性
func (am *AttributeManager) removeRateAttribute(attrs map[string]float64, key string, petCombat map[string]float64) {
	if v, ok := petCombat[key]; ok && v != 0 {
		attrs[key] = attrs[key] - v
		if attrs[key] < 0 {
			attrs[key] = 0
		}
	}
}

// ApplyEquipmentAndPet 同时应用装备和灵宠加成
// 用于穿戴装备时的完整处理
func (am *AttributeManager) ApplyEquipmentAndPet(equipStats map[string]float64, pet *models.Pet, petCombat map[string]float64) {
	am.ApplyEquipmentStats(equipStats)
	am.ApplyPetBonuses(pet, petCombat)
}

// RemoveEquipmentAndApplyPet 移除装备，重新应用灵宠加成
// 用于卸下装备时的完整处理
func (am *AttributeManager) RemoveEquipmentAndApplyPet(equipStats map[string]float64, pet *models.Pet, petCombat map[string]float64) {
	am.RemoveEquipmentStats(equipStats)
	am.ApplyPetBonuses(pet, petCombat)
}
