package exploration

import "math"

// HerbQualities 灵草品质配置
var HerbQualities = map[string]map[string]interface{}{
	"common": {"name": "普通", "value": 1},
	"uncommon": {"name": "优质", "value": 1.5},
	"rare": {"name": "稀有", "value": 2},
	"epic": {"name": "极品", "value": 3},
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
		Chance:      0.4,
	},
	{
		ID:          "cloud_flower",
		Name:        "云雾花",
		Description: "生长在云雾缭绕处的灵花，有助于修炼",
		BaseValue:   15,
		Category:    "cultivation",
		Chance:      0.3,
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
		Chance:      0.1,
	},
	{
		ID:          "immortal_jade_grass",
		Name:        "仙玉草",
		Description: "传说中生长在仙境的灵草，可遇不可求",
		BaseValue:   60,
		Category:    "special",
		Chance:      0.05,
	},
	{
		ID:          "dark_yin_grass",
		Name:        "玄阴草",
		Description: "生长在阴暗处的奇特灵草，具有独特的灵气属性",
		BaseValue:   30,
		Category:    "spirit",
		Chance:      0.2,
	},
	{
		ID:          "nine_leaf_lingzhi",
		Name:        "九叶灵芝",
		Description: "传说中的灵芝，拥有九片叶子，蕴含强大的生命力",
		BaseValue:   45,
		Category:    "cultivation",
		Chance:      0.12,
	},
	{
		ID:          "purple_ginseng",
		Name:        "紫金参",
		Description: "千年紫参，散发着淡淡的黄金，大补元气",
		BaseValue:   50,
		Category:    "attribute",
		Chance:      0.08,
	},
	{
		ID:          "frost_lotus",
		Name:        "寒霜莲",
		Description: "生长在极寒之地的莲花，可以提升修炼者的灵力纯度",
		BaseValue:   55,
		Category:    "spirit",
		Chance:      0.07,
	},
	{
		ID:          "fire_heart_flower",
		Name:        "火心花",
		Description: "生长在火山口的奇花，花心似火焰跳动",
		BaseValue:   35,
		Category:    "attribute",
		Chance:      0.15,
	},
	{
		ID:          "moonlight_orchid",
		Name:        "月华兰",
		Description: "只在月圆之夜绽放的神秘兰花，能吸收月华精华",
		BaseValue:   70,
		Category:    "spirit",
		Chance:      0.04,
	},
	{
		ID:          "sun_essence_flower",
		Name:        "日精花",
		Description: "吸收太阳精华的奇花，蕴含纯阳之力",
		BaseValue:   75,
		Category:    "cultivation",
		Chance:      0.03,
	},
	{
		ID:          "five_elements_grass",
		Name:        "五行草",
		Description: "一株草同时具备金木水火土五种属性的奇珍",
		BaseValue:   80,
		Category:    "attribute",
		Chance:      0.02,
	},
	{
		ID:          "phoenix_feather_herb",
		Name:        "凤羽草",
		Description: "传说生长在不死火凤栖息地的神草，具有涅槃之力",
		BaseValue:   85,
		Category:    "special",
		Chance:      0.015,
	},
	{
		ID:          "celestial_dew_grass",
		Name:        "天露草",
		Description: "凝聚天地精华的仙草，千年一遇",
		BaseValue:   90,
		Category:    "special",
		Chance:      0.01,
	},
}

// PillRecipes 丹药配方配置
var PillRecipes = []PillRecipe{
	{
		ID:              "spirit_gathering",
		Name:            "聚灵丹",
		Description:    "提升灵力恢复速度的丹药",
		Grade:           "grade1",
		Type:            "spirit",
		FragmentsNeeded: 10,
	},
	{
		ID:              "cultivation_boost",
		Name:            "聚气丹",
		Description:    "提升修炼速度的丹药",
		Grade:           "grade2",
		Type:            "cultivation",
		FragmentsNeeded: 15,
	},
	{
		ID:              "thunder_power",
		Name:            "雷灵丹",
		Description:    "提升战斗属性的丹药",
		Grade:           "grade3",
		Type:            "attribute",
		FragmentsNeeded: 20,
	},
	{
		ID:              "immortal_essence",
		Name:            "仙灵丹",
		Description:    "全属性提升的神奇丹药",
		Grade:           "grade4",
		Type:            "special",
		FragmentsNeeded: 25,
	},
	{
		ID:              "five_elements_pill",
		Name:            "五行丹",
		Description:    "融合五行之力的神奇丹药，全面提升修炼者素质",
		Grade:           "grade5",
		Type:            "attribute",
		FragmentsNeeded: 30,
	},
	{
		ID:              "celestial_essence_pill",
		Name:            "天元丹",
		Description:    "凝聚天地精华的极品丹药，大幅提升修炼速度",
		Grade:           "grade6",
		Type:            "cultivation",
		FragmentsNeeded: 35,
	},
	{
		ID:              "sun_moon_pill",
		Name:            "日月丹",
		Description:    "融合日月精华的丹药，能大幅提升灵力上限",
		Grade:           "grade7",
		Type:            "spirit",
		FragmentsNeeded: 40,
	},
	{
		ID:              "phoenix_rebirth_pill",
		Name:            "涅槃丹",
		Description:    "蕴含不死凤凰之力的神丹，能在战斗中自动恢复生命",
		Grade:           "grade8",
		Type:            "special",
		FragmentsNeeded: 45,
	},
	{
		ID:              "spirit_recovery",
		Name:            "回灵丹",
		Description:    "快速恢复灵力的丹药",
		Grade:           "grade2",
		Type:            "spirit",
		FragmentsNeeded: 15,
	},
	{
		ID:              "essence_condensation",
		Name:            "凝元丹",
		Description:    "提升修炼效率的高级丹药",
		Grade:           "grade3",
		Type:            "cultivation",
		FragmentsNeeded: 20,
	},
	{
		ID:              "mind_clarity",
		Name:            "清心丹",
		Description:    "提升心境和悟性的丹药",
		Grade:           "grade3",
		Type:            "special",
		FragmentsNeeded: 20,
	},
	{
		ID:              "fire_essence",
		Name:            "火元丹",
		Description:    "提升火属性修炼速度的丹药",
		Grade:           "grade4",
		Type:            "attribute",
		FragmentsNeeded: 25,
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
