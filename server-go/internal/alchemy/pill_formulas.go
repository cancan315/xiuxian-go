package alchemy

import "math"

// PillFormulaConfig 存储每种丹药的效果值计算公式
// 每种丹药可以有不同的计算方式，灵活调整
type PillFormulaConfig struct {
	PillID    string `json:"pillId"`    // 丹药ID
	Name      string `json:"name"`      // 丹药名称
	EffectKey string `json:"effectKey"` // 效果类型
	Calculate func(baseValue float64, level int) float64
}

// pillFormulas 定义所有丹药的计算公式
// 后期可根据需要灵活调整每种丹药的效果计算方式
var pillFormulas = map[string]PillFormulaConfig{
	// 聚灵丹：880 × 1.2^(Level-1)
	"spirit_gathering": {
		PillID:    "spirit_gathering",
		Name:      "聚灵丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 聚气丹：20 × 1.2^(Level-1)
	"cultivation_boost": {
		PillID:    "cultivation_boost",
		Name:      "聚气丹",
		EffectKey: "cultivation",
		Calculate: func(baseValue float64, level int) float64 {
			return 20 * math.Pow(1.2, float64(level-1))
		},
	},
	// 回灵丹：1880 × 1.2^(Level-1)
	"spirit_recovery": {
		PillID:    "spirit_recovery",
		Name:      "回灵丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 1880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 雷灵丹：10 × 1.2^(Level-1)
	"thunder_power": {
		PillID:    "thunder_power",
		Name:      "雷灵丹",
		EffectKey: "attributeAttack",
		Calculate: func(baseValue float64, level int) float64 {
			return 10 * math.Pow(1.2, float64(level-1))
		},
	},
	// 凝元丹：50 × 1.2^(Level-1)
	"essence_condensation": {
		PillID:    "essence_condensation",
		Name:      "凝元丹",
		EffectKey: "cultivation",
		Calculate: func(baseValue float64, level int) float64 {
			return 50 * math.Pow(1.2, float64(level-1))
		},
	},
	// 清心丹：880 × 1.2^(Level-1)
	"mind_clarity": {
		PillID:    "mind_clarity",
		Name:      "清心丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 仙灵丹：880 × 1.2^(Level-1)
	"immortal_essence": {
		PillID:    "immortal_essence",
		Name:      "仙灵丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 火元丹：880 × 1.2^(Level-1)
	"fire_essence": {
		PillID:    "fire_essence",
		Name:      "火元丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 五行丹：880 × 1.2^(Level-1)
	"five_elements_pill": {
		PillID:    "five_elements_pill",
		Name:      "五行丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 天元丹：880 × 1.2^(Level-1)
	"celestial_essence_pill": {
		PillID:    "celestial_essence_pill",
		Name:      "天元丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 日月丹：880 × 1.2^(Level-1)
	"sun_moon_pill": {
		PillID:    "sun_moon_pill",
		Name:      "日月丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 涅槃丹：880 × 1.2^(Level-1)
	"phoenix_rebirth_pill": {
		PillID:    "phoenix_rebirth_pill",
		Name:      "涅槃丹",
		EffectKey: "spirit",
		Calculate: func(baseValue float64, level int) float64 {
			return 880 * math.Pow(1.2, float64(level-1))
		},
	},
	// 渡劫丹：效果类型jieYingRate，服用后增加渡劫成功率
	"jie_ying_pill": {
		PillID:    "jie_ying_pill",
		Name:      "渡劫丹",
		EffectKey: "jieYingRate",
		Calculate: func(baseValue float64, level int) float64 {
			// 渡劫丹效果：提高渡劫成功率 0.05 (5%)
			return 0.05
		},
	},
}

// GetPillFormula 根据丹药ID获取对应的计算公式
// 如果丹药不在配置中，返回默认的统一公式
func GetPillFormula(pillID string) *PillFormulaConfig {
	if formula, ok := pillFormulas[pillID]; ok {
		return &formula
	}
	// 默认公式：baseValue × 1.2^(Level-1)
	return &PillFormulaConfig{
		PillID:    pillID,
		Name:      "未知丹药",
		EffectKey: "unknown",
		Calculate: func(baseValue float64, level int) float64 {
			return baseValue * math.Pow(1.2, float64(level-1))
		},
	}
}

// CalculatePillEffect 计算丹药效果值
// 根据丹药ID查找对应的公式，然后计算效果值
func CalculatePillEffect(pillID string, baseValue float64, playerLevel int) float64 {
	formula := GetPillFormula(pillID)
	return formula.Calculate(baseValue, playerLevel)
}
