package exploration

import "math"

// HerbQualities 灵草品质配置
var HerbQualities = map[string]map[string]interface{}{
	"common":    {"name": "普通", "value": 1},
	"uncommon":  {"name": "优质", "value": 1.5},
	"rare":      {"name": "稀有", "value": 2},
	"epic":      {"name": "极品", "value": 3},
	"legendary": {"name": "仙品", "value": 5},
}

// HerbQualitiesSlice 品质数组（用于按概率选择）
var HerbQualitiesSlice = []string{"common", "uncommon", "rare", "epic", "legendary"}

// HerbConfigs 灵草配置
var HerbConfigs = []HerbConfig{
	{
		ID:          "spirit_grass",
		Name:        "灵精草",
		Description: "最常见的灵草，蕴含少量灵气",
		BaseValue:   10,
		Category:    "spirit",
		Chance:      0.25,
	},
	{
		ID:          "cloud_flower",
		Name:        "云雾花",
		Description: "生长在云雾缭绕处的灵花，有助于修炼",
		BaseValue:   15,
		Category:    "cultivation",
		Chance:      0.25,
	},
	{
		ID:          "thunder_root",
		Name:        "雷击根",
		Description: "经过雷霆淬炼的灵根，蕴含强大能量",
		BaseValue:   25,
		Category:    "attribute",
		Chance:      0.15,
	},
	{
		ID:          "dragon_breath_herb",
		Name:        "龙息草",
		Description: "吸收龙气孕育的灵草，极为珍贵",
		BaseValue:   40,
		Category:    "special",
		Chance:      0.05,
	},
	{
		ID:          "immortal_jade_grass",
		Name:        "仙玉草",
		Description: "传说中生长在仙境的灵草，可遇不可求",
		BaseValue:   60,
		Category:    "special",
		Chance:      0.00,
	},
	{
		ID:          "dark_yin_grass",
		Name:        "玄阴草",
		Description: "生长在阴暗处的奇特灵草，具有独特的灵气属性",
		BaseValue:   30,
		Category:    "spirit",
		Chance:      0.1,
	},
	{
		ID:          "nine_leaf_lingzhi",
		Name:        "九叶灵芝",
		Description: "传说中的灵芝，拥有九片叶子，蕴含强大的生命力",
		BaseValue:   45,
		Category:    "cultivation",
		Chance:      0.05,
	},
	{
		ID:          "purple_ginseng",
		Name:        "紫金参",
		Description: "千年紫参，散发着淡淡的黄金，大补元气",
		BaseValue:   50,
		Category:    "attribute",
		Chance:      0.05,
	},
	{
		ID:          "frost_lotus",
		Name:        "寒霜莲",
		Description: "生长在极寒之地的莲花，可以提升修炼者的灵力纯度",
		BaseValue:   55,
		Category:    "spirit",
		Chance:      0.1,
	},
	{
		ID:          "fire_heart_flower",
		Name:        "火心花",
		Description: "生长在火山口的奇花，花心似火焰跳动",
		BaseValue:   35,
		Category:    "attribute",
		Chance:      0.00,
	},
	{
		ID:          "moonlight_orchid",
		Name:        "月华兰",
		Description: "只在月圆之夜绽放的神秘兰花，能吸收月华精华",
		BaseValue:   70,
		Category:    "spirit",
		Chance:      0.00,
	},
	{
		ID:          "sun_essence_flower",
		Name:        "日精花",
		Description: "吸收太阳精华的奇花，蕴含纯阳之力",
		BaseValue:   75,
		Category:    "cultivation",
		Chance:      0.00,
	},
	{
		ID:          "five_elements_grass",
		Name:        "五行草",
		Description: "一株草同时具备金木水火土五种属性的奇珍",
		BaseValue:   80,
		Category:    "attribute",
		Chance:      0.00,
	},
	{
		ID:          "phoenix_feather_herb",
		Name:        "凤羽草",
		Description: "传说生长在不死火凤栖息地的神草，具有涅槃之力",
		BaseValue:   85,
		Category:    "special",
		Chance:      0.00,
	},
	{
		ID:          "celestial_dew_grass",
		Name:        "天露草",
		Description: "凝聚天地精华的仙草，千年一遇",
		BaseValue:   90,
		Category:    "special",
		Chance:      0.00,
	},
}

// PillRecipes 丹药配方配置（Weight 为掉率权重，数字越大越容易掉）
var PillRecipes = []PillRecipe{
	{
		ID:              "spirit_gathering",
		Name:            "聚灵丹",
		Description:     "十年灵草炼制，服用后恢复少量灵力",
		Grade:           "grade1",
		Type:            "spirit",
		FragmentsNeeded: 10,
		Weight:          25, // 最常见
	},
	{
		ID:              "cultivation_boost",
		Name:            "聚气丹",
		Description:     "十年灵草炼制，服用后增加少量修为",
		Grade:           "grade2",
		Type:            "cultivation",
		FragmentsNeeded: 15,
		Weight:          30, // 常见
	},
	{
		ID:              "spirit_recovery",
		Name:            "回灵丹",
		Description:     "百年灵草炼制，服用后恢复大量灵力",
		Grade:           "grade2",
		Type:            "spirit",
		FragmentsNeeded: 15,
		Weight:          20, // 常见
	},
	{
		ID:              "thunder_power",
		Name:            "雷灵丹",
		Description:     "千年灵草炼制，蕴含狂暴的天雷能量，服用后增加攻击",
		Grade:           "grade3",
		Type:            "attribute",
		FragmentsNeeded: 20,
		Weight:          10, // 较少
	},
	{
		ID:              "essence_condensation",
		Name:            "凝元丹",
		Description:     "千年灵草炼制，服用后增加大量修为",
		Grade:           "grade3",
		Type:            "cultivation",
		FragmentsNeeded: 20,
		Weight:          5, // 较少
	},
	{
		ID:              "jie_ying_pill",
		Name:            "渡劫丹",
		Description:     "万年灵草炼制，服用后增加少许渡劫成功率",
		Grade:           "grade3",
		Type:            "cultivation",
		FragmentsNeeded: 20,
		Weight:          10, // 较少
	},
	{
		ID:              "mind_clarity",
		Name:            "清心丹",
		Description:     "暂未开放",
		Grade:           "grade3",
		Type:            "spirit",
		FragmentsNeeded: 20,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "immortal_essence",
		Name:            "仙灵丹",
		Description:     "暂未开放",
		Grade:           "grade4",
		Type:            "spirit",
		FragmentsNeeded: 25,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "fire_essence",
		Name:            "火元丹",
		Description:     "暂未开放",
		Grade:           "grade4",
		Type:            "spirit",
		FragmentsNeeded: 25,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "five_elements_pill",
		Name:            "五行丹",
		Description:     "暂未开放",
		Grade:           "grade5",
		Type:            "spirit",
		FragmentsNeeded: 30,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "celestial_essence_pill",
		Name:            "天元丹",
		Description:     "暂未开放",
		Grade:           "grade6",
		Type:            "spirit",
		FragmentsNeeded: 35,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "sun_moon_pill",
		Name:            "日月丹",
		Description:     "暂未开放",
		Grade:           "grade7",
		Type:            "spirit",
		FragmentsNeeded: 40,
		Weight:          0, // 暂未开放
	},
	{
		ID:              "phoenix_rebirth_pill",
		Name:            "涅槃丹",
		Description:     "暂未开放",
		Grade:           "grade8",
		Type:            "spirit",
		FragmentsNeeded: 45,
		Weight:          0, // 暂未开放
	},
}

// GetHerbValue 计算灵草价值
func GetHerbValue(baseValue int, quality string) int {
	qualityInfo, ok := HerbQualities[quality]
	if !ok {
		return baseValue
	}
	multiplier, _ := qualityInfo["value"].(float64)
	return int(math.Floor(float64(baseValue) * multiplier))
}

// GetRandomQuality 获取随机品质
func GetRandomQuality(rand float64) string {
	// 50% 普通, 30% 优质, 15% 稀有, 4% 极品, 1% 仙品
	if rand < 0.5 {
		return "common"
	} else if rand < 0.8 {
		return "uncommon"
	} else if rand < 0.95 {
		return "rare"
	} else if rand < 0.99 {
		return "epic"
	}
	return "legendary"
}
