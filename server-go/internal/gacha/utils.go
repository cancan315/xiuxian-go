package gacha

import (
	"encoding/json"
	"math/rand"
	"time"

	"gorm.io/datatypes"
)

// ---------- 工具函数 ----------

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

func ShuffleStrings(a []string) {
	rand.Seed(time.Now().UnixNano())
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

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

func GenerateAttributeValue(attrType, quality string, isPet bool) float64 {
	baseRanges := map[string]struct{ Min, Max float64 }{
		"attack":  {30, 100},
		"health":  {60, 200},
		"defense": {15, 50},
		"speed":   {5, 20},
	}
	combatMin, combatMax := 0.01, 0.10
	resistMin, resistMax := 0.01, 0.08
	specialMin, specialMax := 0.005, 0.05

	if Contains(AttributeTypes.BaseAttributes, attrType) {
		rangeVal, ok := baseRanges[attrType]
		if !ok {
			rangeVal = struct{ Min, Max float64 }{10, 50}
		}
		qualityMods := map[string]float64{
			"common":    1.0,
			"uncommon":  1.2,
			"rare":      1.5,
			"epic":      2.0,
			"legendary": 2.5,
			"mythic":    3.0,
		}
		mod := qualityMods[quality]
		if mod == 0 {
			mod = 1.0
		}
		return float64(int((rand.Float64()*(rangeVal.Max-rangeVal.Min)+rangeVal.Min)*mod + 0.5))
	}
	if Contains(AttributeTypes.CombatAttributes, attrType) {
		v := rand.Float64()*(combatMax-combatMin) + combatMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if Contains(AttributeTypes.CombatResistance, attrType) {
		v := rand.Float64()*(resistMax-resistMin) + resistMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if Contains(AttributeTypes.SpecialAttributes, attrType) {
		v := rand.Float64()*(specialMax-specialMin) + specialMin
		return float64(int(v*1000+0.5)) / 1000
	}
	return float64(int((rand.Float64()*40 + 10) + 0.5))
}

func Contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func GenerateQualityBasedAttributes(kind, quality string) map[string]float64 {
	rulesPerKind, ok := QualityAttributeRules[kind]
	if !ok {
		return map[string]float64{}
	}
	rules, ok := rulesPerKind[quality]
	if !ok {
		return map[string]float64{}
	}
	attrs := make(map[string]float64)

	if rules.Base > 0 {
		for _, attr := range SelectRandomAttributes(AttributeTypes.BaseAttributes, rules.Base) {
			attrs[attr] = GenerateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Combat > 0 {
		for _, attr := range SelectRandomAttributes(AttributeTypes.CombatAttributes, rules.Combat) {
			attrs[attr] = GenerateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Resistance > 0 {
		for _, attr := range SelectRandomAttributes(AttributeTypes.CombatResistance, rules.Resistance) {
			attrs[attr] = GenerateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Special > 0 {
		for _, attr := range SelectRandomAttributes(AttributeTypes.SpecialAttributes, rules.Special) {
			attrs[attr] = GenerateAttributeValue(attr, quality, kind == "pet")
		}
	}
	return attrs
}