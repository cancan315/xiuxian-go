package gacha

// ---------- 概率与配置，与 Node gachaController.js 保持一致 ----------

// EquipmentQualityProbabilities 装备品质固定概率配置
// 所有玩家无论等级如何，都使用相同的概率抽取各品质装备
var EquipmentQualityProbabilities = map[string]float64{
	"mythic":    0.0001, // 仙器 - 0.01%
	"legendary": 0.0003, // 伪仙器 - 0.03%
	"epic":      0.0026, // 灵器 - 0.26%
	"rare":      0.032,  // 魔器 - 3.2%
	"uncommon":  0.165,  // 法器 - 16%
	"common":    0.80,   // 凡器 - 80%
}

// PetRarityProbabilities 灵宠稀有度固定概率配置
// 所有玩家无论等级如何，都使用相同的概率抽取各稀有度灵宠
var PetRarityProbabilities = map[string]float64{
	"mythic":    0.0001, // 仙兽 - 0.01%
	"legendary": 0.0003, // 瑞兽 - 0.03%
	"epic":      0.0026, // 上古异兽 - 0.26%
	"rare":      0.032,  // 灵兽 - 3.2%
	"uncommon":  0.165,  // 妖兽 - 16%
	"common":    0.80,   // 凡兽 - 80%
}

// AttributePool 属性池结构
type AttributePool struct {
	Base       []string
	Combat     []string
	Resistance []string
	Special    []string
}

// AttributesByType 按属性类型分组的属性池
// 攻击相关、生命相关、防御相关、速度相关、通用属性
var AttributesByType = map[string]AttributePool{
	"attack": {
		Base:    []string{"attack"},                              // 攻击力
		Combat:  []string{"critRate", "stunRate", "vampireRate"}, // 暴击率、眩晕率、吸血率
		Special: []string{"critDamageBoost", "finalDamageBoost"}, // 强化爆伤、最终增伤
	},
	"health": {
		Base:       []string{"health"},        // 生命值
		Resistance: []string{"vampireResist"}, // 抗吸血
		Special:    []string{"healBoost"},     // 强化治疗
	},
	"defense": {
		Base:       []string{"defense"},                                   // 防御力
		Combat:     []string{"counterRate"},                               // 反击率
		Resistance: []string{"critResist", "counterResist", "stunResist"}, // 抗暴击、抗反击、抗眩晕
		Special:    []string{"critDamageReduce", "finalDamageReduce"},     // 弱化爆伤、最终减伤
	},
	"speed": {
		Base:       []string{"speed"},                      // 速度
		Combat:     []string{"comboRate", "dodgeRate"},     // 连击率、闪避率
		Resistance: []string{"comboResist", "dodgeResist"}, // 抗连击、抗闪避
	},
	"common": {
		Special: []string{"combatBoost", "resistanceBoost"}, // 战斗强化、抗性强化
	},
}

// EquipTypeAttributeMapping 装备类型对应的属性类型映射
// faqi生成攻击相关属性
// guanjin生成生命和防御类属性
// daopao生成生命和防御相关属性
// yunlv生成速度相关属性
// fabao生成攻击、生命、防御、速度相关属性及通用属性
var EquipTypeAttributeMapping = map[string][]string{
	"faqi":    {"attack"},
	"guanjin": {"health", "defense"},
	"daopao":  {"health", "defense"},
	"yunlv":   {"speed"},
	"fabao":   {"attack", "health", "defense", "speed", "common"},
}

// QualityAttributeCount 品质对应的属性条数结构
type QualityAttributeCount struct {
	Base       int
	Combat     int
	Resistance int
	Special    int
}

// QualityAttributeRules 装备品质对应的属性条数规则
// common:    1条基础属性
// uncommon:  1条基础属性，1条战斗属性
// rare:      1条基础属性，1条战斗属性，1条战斗抗性
// epic:      1条基础属性，2条战斗属性，2条战斗抗性
// legendary: 1条基础属性，2条战斗属性，2条战斗抗性，1条特殊属性
// mythic:    1条基础属性，2条战斗属性，2条战斗抗性，2条特殊属性
var QualityAttributeRules = map[string]map[string]QualityAttributeCount{
	"equipment": {
		"common":    {Base: 1},
		"uncommon":  {Base: 1, Combat: 1},
		"rare":      {Base: 1, Combat: 1, Resistance: 1},
		"epic":      {Base: 1, Combat: 2, Resistance: 2},
		"legendary": {Base: 1, Combat: 2, Resistance: 2, Special: 1},
		"mythic":    {Base: 1, Combat: 2, Resistance: 2, Special: 2},
	},
	"pet": {
		"common":    {Base: 1},
		"uncommon":  {Base: 2},
		"rare":      {Base: 3},
		"epic":      {Base: 4, Combat: 1},
		"legendary": {Base: 4, Combat: 6, Resistance: 6},
		"mythic":    {Base: 4, Combat: 6, Resistance: 6, Special: 7},
	},
}

// 简化：只实现与随机生成直接相关的名称/描述表，细节与 Node 相同意义即可
var PetNamesByRarity = map[string][]string{
	"common":    {"白狐", "灰狼", "黄牛", "黑虎", "赤马", "棕熊", "青蛇", "紫貂", "银鼠", "金蝉", "彩雀", "田园犬"},
	"uncommon":  {"妖猴", "妖驴", "妖豹", "妖蛇", "妖虎", "妖熊", "妖鹰", "妖蝎", "妖蛛", "妖蝠", "妖蟾", "妖蜈"},
	"rare":      {"灵狐", "灵鹿", "灵龟", "灵蛇", "灵猿", "灵鹤", "灵鲤", "灵雀", "灵虎", "灵豹", "灵猫", "灵犬"},
	"epic":      {"饕餮", "穷奇", "梼杌", "混沌", "九婴", "相柳", "凿齿", "修蛇", "封豨", "大风", "巴蛇", "朱厌"},
	"legendary": {"麒麟", "凤凰", "龙龟", "白泽", "重明鸟", "当康", "乘黄", "英招", "夫诸", "天马", "青牛", "玄龟"},
	"mythic":    {"应龙", "夔牛", "毕方", "饕餮", "九尾狐", "玉兔", "金蟾", "青鸾", "火凤", "水麒麟", "土蝼", "陆吾"},
}

var PetDescriptionsByRarity = map[string][]string{
	"common":    {"一只普通的小动物，刚刚开启灵智"},
	"uncommon":  {"已具初步妖力的妖兽，有一定培养价值"},
	"rare":      {"天生蕴含灵气的珍稀兽类，颇具潜力"},
	"epic":      {"来自上古时代的神秘异兽，力量强大"},
	"legendary": {"祥瑞降临，世间少有的瑞兽"},
	"mythic":    {"超越凡俗的仙界仙兽，几近传说"},
}

// 装备类型配置 EquipType字段是装备类型
var EquipmentTypes = map[string]struct {
	Name      string
	EquipType string
	Prefixes  map[string][]string
}{
	"faqi": {
		Name:      "法宝",
		EquipType: "faqi",
		Prefixes: map[string][]string{
			"common":    {"青竹剑", "柳叶刀", "精铁剑", "黑木杖"},
			"uncommon":  {"法剑", "法刀", "法杖", "法珠"},
			"rare":      {"上品", "精品", "珍品", "灵韵"},
			"epic":      {"异宝", "奇珍", "灵光", "宝华"},
			"legendary": {"法宝", "神兵", "灵宝", "仙锋"},
			"mythic":    {"仙器", "神器", "天器", "圣器"},
		},
	},
	"guanjin": {
		Name:      "冠巾",
		EquipType: "guanjin",
		Prefixes: map[string][]string{
			"common":    {"布制", "麻织", "粗布", "素巾"},
			"uncommon":  {"丝织", "锦制", "绣冠", "轻纱"},
			"rare":      {"灵丝", "云锦", "霓裳", "羽冠"},
			"epic":      {"宝冠", "灵冠", "星辰", "月华"},
			"legendary": {"仙冠", "神冠", "紫金", "龙纹"},
			"mythic":    {"天冠", "圣冠", "混沌", "太虚"},
		},
	},
	"daopao": {
		Name:      "道袍",
		EquipType: "daopao",
		Prefixes: map[string][]string{
			"common":    {"粗布", "麻衣", "布衣", "素袍"},
			"uncommon":  {"丝绸", "锦袍", "绣衣", "轻衫"},
			"rare":      {"灵绸", "云裳", "霞衣", "霓裳"},
			"epic":      {"宝衣", "灵衣", "星辰", "月华"},
			"legendary": {"仙衣", "神袍", "紫绶", "龙袍"},
			"mythic":    {"天衣", "圣袍", "混沌", "太虚"},
		},
	},
	"yunlv": {
		Name:      "云履",
		EquipType: "yunlv",
		Prefixes: map[string][]string{
			"common":    {"布鞋", "草鞋", "麻鞋", "木屐"},
			"uncommon":  {"皮靴", "丝履", "锦履", "绣鞋"},
			"rare":      {"灵靴", "云履", "霞履", "霓履"},
			"epic":      {"宝履", "灵履", "星辰", "月华"},
			"legendary": {"仙履", "神履", "紫霞", "龙履"},
			"mythic":    {"天履", "圣履", "混沌", "太虚"},
		},
	},
	"fabao": {
		Name:      "本命法宝",
		EquipType: "fabao",
		Prefixes: map[string][]string{
			"common":    {"粗制", "劣质", "仿制", "赝品"},
			"uncommon":  {"精制", "良品", "上品", "优质"},
			"rare":      {"灵宝", "珍宝", "异宝", "奇宝"},
			"epic":      {"重宝", "至宝", "灵光", "宝华"},
			"legendary": {"仙宝", "神宝", "天宝", "圣宝"},
			"mythic":    {"至尊", "无上", "混沌", "太虚"},
		},
	},
}
