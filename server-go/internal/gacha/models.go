package gacha

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// ---------- 随机生成逻辑 ----------

// GachaEquipment 抽卡获得的装备数据结构
// 用于在抽卡过程中临时存储生成的装备信息
type GachaEquipment struct {
	ID            string                 `json:"id"`               // 装备唯一标识符
	Name          string                 `json:"name"`             // 装备名称
	Type          string                 `json:"type"`             // 装备类型
	Quality       string                 `json:"quality"`          // 装备品质
	EquipType     string                 `json:"equip_type"`       // 装备类型
	Slot          string                 `json:"slot"`             // 装备槽位
	Level         int                    `json:"level"`            // 装备等级
	RequiredRealm int                    `json:"required_realm"`   // 需要境界等级
	EnhanceLevel  int                    `json:"enhance_level"`    // 强化等级
	Stats         map[string]float64     `json:"stats"`            // 装备基础属性
	ExtraAttrs    map[string]interface{} `json:"extra_attributes"` // 装备额外属性
}

// GachaPet 抽卡获得的灵宠数据结构
// 用于在抽卡过程中临时存储生成的灵宠信息
type GachaPet struct {
	ID              string             `json:"id"`                // 灵宠唯一标识符
	Name            string             `json:"name"`              // 灵宠名称
	Type            string             `json:"type"`              // 类型（固定为"pet"）
	Rarity          string             `json:"rarity"`            // 稀有度
	Level           int                `json:"level"`             // 等级
	Star            int                `json:"star"`              // 星级
	Exp             int                `json:"exp"`               // 经验值
	Description     string             `json:"description"`       // 描述信息
	CombatAttrs     map[string]float64 `json:"combat_attributes"` // 战斗属性
	AttackBonus     float64            `json:"attack_bonus"`      // 攻击加成
	DefenseBonus    float64            `json:"defense_bonus"`     // 防御加成
	HealthBonus     float64            `json:"health_bonus"`      // 生命加成
	CreatedAtMillis int64              `json:"created_at"`        // 创建时间戳
}

// GenerateRandomEquipment 生成随机装备（使用固定概率）
// 参数: level - 玩家当前等级（此参数现在仅用于装备等级，不影响品质概率）
// 返回: 生成的装备对象
func GenerateRandomEquipment(level int) GachaEquipment {
	// i) 选择品质（使用固定概率）
	qualities := []string{"mythic", "legendary", "epic", "rare", "uncommon", "common"}
	weights := make([]float64, len(qualities))
	for i, q := range qualities {
		weights[i] = EquipmentQualityProbabilities[q] * 100
	}
	qIdx := WeightedRandom(weights)
	quality := qualities[qIdx]

	// ii) 装备类型
	keys := make([]string, 0, len(EquipmentTypes))
	for k := range EquipmentTypes {
		keys = append(keys, k)
	}
	etypeKey := keys[rand.Intn(len(keys))]
	et := EquipmentTypes[etypeKey]

	prefixes := et.Prefixes[quality]
	prefix := prefixes[rand.Intn(len(prefixes))]
	name := prefix + et.Name

	stats := GenerateQualityBasedAttributes("equipment", quality, etypeKey)

	numericLevel := level
	if numericLevel <= 0 {
		numericLevel = 1
	}

	// 境界计算逻辑按 Node 注释映射
	requiredRealm := 1
	switch {
	case numericLevel >= 1 && numericLevel <= 9:
		requiredRealm = 1
	case numericLevel >= 10 && numericLevel <= 18:
		requiredRealm = 2
	case numericLevel >= 19 && numericLevel <= 27:
		requiredRealm = 3
	case numericLevel >= 28 && numericLevel <= 36:
		requiredRealm = 4
	case numericLevel >= 37 && numericLevel <= 45:
		requiredRealm = 5
	case numericLevel >= 46 && numericLevel <= 54:
		requiredRealm = 6
	case numericLevel >= 55 && numericLevel <= 63:
		requiredRealm = 7
	case numericLevel >= 64 && numericLevel <= 72:
		requiredRealm = 8
	case numericLevel >= 73 && numericLevel <= 81:
		requiredRealm = 9
	case numericLevel >= 82 && numericLevel <= 90:
		requiredRealm = 10
	case numericLevel >= 91 && numericLevel <= 99:
		requiredRealm = 11
	case numericLevel >= 100 && numericLevel <= 108:
		requiredRealm = 12
	case numericLevel >= 109 && numericLevel <= 117:
		requiredRealm = 13
	case numericLevel >= 118 && numericLevel <= 126:
		requiredRealm = 14
	case numericLevel >= 127:
		requiredRealm = 15
	}

	return GachaEquipment{
		ID:            uuid.NewString(),
		Name:          name,
		Type:          "equipment",
		Quality:       quality,
		EquipType:     etypeKey,
		Level:         numericLevel,
		RequiredRealm: requiredRealm,
		EnhanceLevel:  0,
		Stats:         stats,
		ExtraAttrs:    map[string]interface{}{},
	}
}

// GenerateRandomPet 生成随机灵宠（使用固定概率）
// 参数: level - 玩家当前等级（此参数现在仅用于灵宠等级，不影响稀有度概率）
// 返回: 生成的灵宠对象
func GenerateRandomPet(level int) GachaPet {
	rarities := []string{"mythic", "legendary", "epic", "rare", "uncommon", "common"}
	weights := make([]float64, len(rarities))
	for i, r := range rarities {
		weights[i] = PetRarityProbabilities[r] * 1000
	}
	rIdx := WeightedRandom(weights)
	rarity := rarities[rIdx]

	names := PetNamesByRarity[rarity]
	name := names[rand.Intn(len(names))]
	descs := PetDescriptionsByRarity[rarity]
	desc := descs[rand.Intn(len(descs))]

	combatAttrs := GenerateQualityBasedAttributes("pet", rarity, "")

	qualityBonusMap := map[string]float64{
		"mythic":    0.15,
		"legendary": 0.12,
		"epic":      0.09,
		"rare":      0.06,
		"uncommon":  0.03,
		"common":    0.03,
	}
	starBonusPerQuality := map[string]float64{
		"mythic":    0.02,
		"legendary": 0.01,
		"epic":      0.01,
		"rare":      0.01,
		"uncommon":  0.01,
		"common":    0.01,
	}

	baseBonus := qualityBonusMap[rarity]
	starBonus := 0 * starBonusPerQuality[rarity]
	levelBonus := 0 * (baseBonus * 0.1)
	phase := 0
	phaseBonus := float64(phase) * (baseBonus * 0.5)
	finalBonus := baseBonus + starBonus + levelBonus + phaseBonus

	return GachaPet{
		ID:              uuid.NewString(),
		Name:            name,
		Type:            "pet",
		Rarity:          rarity,
		Level:           1,
		Star:            0,
		Exp:             0,
		Description:     desc,
		CombatAttrs:     combatAttrs,
		AttackBonus:     finalBonus,
		DefenseBonus:    finalBonus,
		HealthBonus:     finalBonus,
		CreatedAtMillis: time.Now().UnixMilli(),
	}
}
