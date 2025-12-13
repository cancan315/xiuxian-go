package gacha

import (
	"encoding/json"
	"math/rand"
	"time"

	"gorm.io/datatypes"
)

// ---------- 工具函数 ----------

// ToJSON 将任意接口类型数据转换为数据库可用的JSON格式
// 参数: v - 需要转换的任意类型数据
// 返回: 转换后的datatypes.JSON类型数据
func ToJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(b)
}

// WeightedRandom 根据权重随机选择一个索引
// 参数: weights - 权重数组，每个元素代表对应索引的权重
// 返回: 根据权重随机选择的索引值
func WeightedRandom(weights []float64) int {
	var total float64
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return 0
	}
	r := rand.Float64() * total
	for i, w := range weights {
		if r < w {
			return i
		}
		r -= w
	}
	return len(weights) - 1
}

// ShuffleStrings 随机打乱字符串切片顺序
// 参数: a - 需要打乱顺序的字符串切片
func ShuffleStrings(a []string) {
	rand.Seed(time.Now().UnixNano())
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

// SelectRandomAttributes 从属性列表中随机选择指定数量的属性
// 参数: list - 可选属性列表
//
//	count - 需要选择的属性数量
//
// 返回: 随机选择的属性列表
func SelectRandomAttributes(list []string, count int) []string {
	if count <= 0 || len(list) == 0 {
		return nil
	}
	if count >= len(list) {
		return append([]string{}, list...)
	}
	copyList := append([]string{}, list...)
	ShuffleStrings(copyList)
	return copyList[:count]
}

// GetAttributesByEquipType 根据装备类型获取相关属性列表
// 参数: equipType - 装备类型(faqi/guanjin/daopao/yunlv/fabao)
//
//	attrCategory - 属性类别(Base/Combat/Resistance/Special)
//
// 返回: 对应的属性列表
func GetAttributesByEquipType(equipType, attrCategory string) []string {
	// 获取装备类型关联的属性类型
	attrTypes, exists := EquipTypeAttributeMapping[equipType]
	if !exists {
		// 默认返回空切片
		return []string{}
	}

	// 收集所有相关属性
	var result []string
	for _, attrType := range attrTypes {
		pool, exists := AttributesByType[attrType]
		if !exists {
			continue
		}

		switch attrCategory {
		case "Base":
			result = append(result, pool.Base...)
		case "Combat":
			result = append(result, pool.Combat...)
		case "Resistance":
			result = append(result, pool.Resistance...)
		case "Special":
			result = append(result, pool.Special...)
		}
	}

	return result
}

// GenerateAttributeValue 根据属性类型和品质生成属性值
// 参数: attrType - 属性类型（如attack、health等）
//
//	quality - 物品品质（common、uncommon、rare等）
//	isPet - 是否为灵宠属性
//
// 返回: 生成的属性值
//
// 各品质属性范围:
// 基础属性(attack,health,defense,speed)固定范围:
//
//	common:    attack(10-50) health(100-300) defense(5-30) speed(5-20)
//	uncommon:  attack(50-100) health(300-600) defense(30-60) speed(30-60)
//	rare:      attack(100-200) health(600-1200) defense(60-120) speed(60-120)
//	epic:      attack(200-500) health(1200-2500) defense(120-250) speed(120-250)
//	legendary: attack(500-1000) health(2500-5000) defense(250-500) speed(250-500)
//	mythic:    attack(1000-2000) health(5000-10000) defense(500-1000) speed(500-1000)
//
// 战斗属性(critRate等)范围:
//
//	common:    1%-5%
//	uncommon:  5%-10%
//	rare:      10%-15%
//	epic:      15%-20%
//	legendary: 20%-30%
//	mythic:    30%-50%
//
// 抗性属性(critResist等)范围:
//
//	common:    1%-5%
//	uncommon:  5%-10%
//	rare:      10%-15%
//	epic:      15%-20%
//	legendary: 20%-30%
//	mythic:    30%-50%
//
// 特殊属性(healBoost等)范围:
//
//	common:    1%-5%
//	uncommon:  5%-10%
//	rare:      10%-15%
//	epic:      15%-20%
//	legendary: 20%-30%
//	mythic:    30%-50%
func GenerateAttributeValue(attrType, quality string, isPet bool) float64 {
	// 根据品质确定基础属性范围
	qualityBaseRanges := map[string]map[string]struct{ Min, Max float64 }{
		"common": {
			"attack":  {10, 50},
			"health":  {100, 300},
			"defense": {5, 30},
			"speed":   {5, 20},
		},
		"uncommon": {
			"attack":  {50, 100},
			"health":  {300, 600},
			"defense": {30, 60},
			"speed":   {30, 60},
		},
		"rare": {
			"attack":  {100, 200},
			"health":  {600, 1200},
			"defense": {60, 120},
			"speed":   {60, 120},
		},
		"epic": {
			"attack":  {200, 500},
			"health":  {1200, 2500},
			"defense": {120, 250},
			"speed":   {120, 250},
		},
		"legendary": {
			"attack":  {500, 1000},
			"health":  {2500, 5000},
			"defense": {250, 500},
			"speed":   {250, 500},
		},
		"mythic": {
			"attack":  {1000, 2000},
			"health":  {5000, 10000},
			"defense": {500, 1000},
			"speed":   {500, 1000},
		},
	}

	// 根据品质确定战斗属性、抗性属性和特殊属性的范围
	qualityRanges := map[string]struct {
		CombatMin, CombatMax   float64
		ResistMin, ResistMax   float64
		SpecialMin, SpecialMax float64
	}{
		"common":    {0.01, 0.05, 0.01, 0.05, 0.01, 0.05},
		"uncommon":  {0.05, 0.10, 0.05, 0.10, 0.05, 0.10},
		"rare":      {0.10, 0.15, 0.10, 0.15, 0.10, 0.15},
		"epic":      {0.15, 0.20, 0.15, 0.20, 0.15, 0.20},
		"legendary": {0.20, 0.30, 0.20, 0.30, 0.20, 0.30},
		"mythic":    {0.30, 0.50, 0.30, 0.50, 0.30, 0.50},
	}

	// 创建一个包含所有基础属性的列表用于判断
	var allBaseAttributes []string
	for _, pool := range AttributesByType {
		allBaseAttributes = append(allBaseAttributes, pool.Base...)
	}

	// 创建一个包含所有战斗属性的列表用于判断
	var allCombatAttributes []string
	for _, pool := range AttributesByType {
		allCombatAttributes = append(allCombatAttributes, pool.Combat...)
	}

	// 创建一个包含所有抗性属性的列表用于判断
	var allResistanceAttributes []string
	for _, pool := range AttributesByType {
		allResistanceAttributes = append(allResistanceAttributes, pool.Resistance...)
	}

	// 创建一个包含所有特殊属性的列表用于判断
	var allSpecialAttributes []string
	for _, pool := range AttributesByType {
		allSpecialAttributes = append(allSpecialAttributes, pool.Special...)
	}

	if Contains(allBaseAttributes, attrType) {
		ranges, ok := qualityBaseRanges[quality]
		if !ok {
			// 默认使用common品质范围
			ranges = qualityBaseRanges["common"]
		}
		rangeVal, ok := ranges[attrType]
		if !ok {
			rangeVal = struct{ Min, Max float64 }{10, 50}
		}
		value := rand.Float64()*(rangeVal.Max-rangeVal.Min) + rangeVal.Min
		return float64(int(value + 0.5))
	}
	if Contains(allCombatAttributes, attrType) {
		ranges := qualityRanges[quality]
		v := rand.Float64()*(ranges.CombatMax-ranges.CombatMin) + ranges.CombatMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if Contains(allResistanceAttributes, attrType) {
		ranges := qualityRanges[quality]
		v := rand.Float64()*(ranges.ResistMax-ranges.ResistMin) + ranges.ResistMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if Contains(allSpecialAttributes, attrType) {
		ranges := qualityRanges[quality]
		v := rand.Float64()*(ranges.SpecialMax-ranges.SpecialMin) + ranges.SpecialMin
		return float64(int(v*1000+0.5)) / 1000
	}
	return float64(int((rand.Float64()*40 + 10) + 0.5))
}

// Contains 检查字符串是否存在于字符串切片中
// 参数: list - 字符串切片
//
//	v - 需要检查的字符串
//
// 返回: 如果存在返回true，否则返回false
func Contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// GenerateQualityBasedAttributes 根据物品类型、品质和装备类型生成属性集合
// 参数: kind - 物品类型（equipment装备或pet灵宠）
//
//	quality - 物品品质
//	equipType - 装备类型（仅对装备有效）
//
// 返回: 生成的属性映射表
func GenerateQualityBasedAttributes(kind, quality, equipType string) map[string]float64 {
	rulesPerKind, ok := QualityAttributeRules[kind]
	if !ok {
		return map[string]float64{}
	}
	rules, ok := rulesPerKind[quality]
	if !ok {
		return map[string]float64{}
	}
	attrs := make(map[string]float64)

	// 对于宠物，使用所有属性类型
	if kind == "pet" {
		var allBaseAttributes []string
		var allCombatAttributes []string
		var allResistanceAttributes []string
		var allSpecialAttributes []string

		for _, pool := range AttributesByType {
			allBaseAttributes = append(allBaseAttributes, pool.Base...)
			allCombatAttributes = append(allCombatAttributes, pool.Combat...)
			allResistanceAttributes = append(allResistanceAttributes, pool.Resistance...)
			allSpecialAttributes = append(allSpecialAttributes, pool.Special...)
		}

		if rules.Base > 0 {
			for _, attr := range SelectRandomAttributes(allBaseAttributes, rules.Base) {
				attrs[attr] = GenerateAttributeValue(attr, quality, true)
			}
		}
		if rules.Combat > 0 {
			for _, attr := range SelectRandomAttributes(allCombatAttributes, rules.Combat) {
				attrs[attr] = GenerateAttributeValue(attr, quality, true)
			}
		}
		if rules.Resistance > 0 {
			for _, attr := range SelectRandomAttributes(allResistanceAttributes, rules.Resistance) {
				attrs[attr] = GenerateAttributeValue(attr, quality, true)
			}
		}
		if rules.Special > 0 {
			for _, attr := range SelectRandomAttributes(allSpecialAttributes, rules.Special) {
				attrs[attr] = GenerateAttributeValue(attr, quality, true)
			}
		}
	} else {
		// 对于装备，根据装备类型选择属性
		if rules.Base > 0 {
			baseAttrs := GetAttributesByEquipType(equipType, "Base")
			for _, attr := range SelectRandomAttributes(baseAttrs, rules.Base) {
				attrs[attr] = GenerateAttributeValue(attr, quality, false)
			}
		}
		if rules.Combat > 0 {
			combatAttrs := GetAttributesByEquipType(equipType, "Combat")
			for _, attr := range SelectRandomAttributes(combatAttrs, rules.Combat) {
				attrs[attr] = GenerateAttributeValue(attr, quality, false)
			}
		}
		if rules.Resistance > 0 {
			resistanceAttrs := GetAttributesByEquipType(equipType, "Resistance")
			for _, attr := range SelectRandomAttributes(resistanceAttrs, rules.Resistance) {
				attrs[attr] = GenerateAttributeValue(attr, quality, false)
			}
		}
		if rules.Special > 0 {
			specialAttrs := GetAttributesByEquipType(equipType, "Special")
			for _, attr := range SelectRandomAttributes(specialAttrs, rules.Special) {
				attrs[attr] = GenerateAttributeValue(attr, quality, false)
			}
		}
	}
	return attrs
}
