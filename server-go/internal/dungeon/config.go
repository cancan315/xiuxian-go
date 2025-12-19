package dungeon

// BuffConfigPool 增益配置池
var BuffConfigPool = map[string][]BuffConfig{
	"common": {
		{ID: "heal", Name: "气血增加", Description: "增加10%血量", Type: "common", Effects: map[string]interface{}{"health": 0.1}},
		{ID: "small_buff", Name: "小幅强化", Description: "增加10%伤害", Type: "common", Effects: map[string]interface{}{"finalDamageBoost": 0.1}},
		{ID: "defense_boost", Name: "铁壁", Description: "提升20%防御力", Type: "common", Effects: map[string]interface{}{"defense": 0.2}},
		{ID: "speed_boost", Name: "疾风", Description: "提升15%速度", Type: "common", Effects: map[string]interface{}{"speed": 0.15}},
		{ID: "crit_boost", Name: "会心", Description: "提升15%暴击率", Type: "common", Effects: map[string]interface{}{"critRate": 0.15}},
		{ID: "dodge_boost", Name: "轻身", Description: "提升15%闪避率", Type: "common", Effects: map[string]interface{}{"dodgeRate": 0.15}},
		{ID: "vampire_boost", Name: "吸血", Description: "提升10%吸血率", Type: "common", Effects: map[string]interface{}{"vampireRate": 0.1}},
		{ID: "combat_boost", Name: "战意", Description: "提升10%战斗属性", Type: "common", Effects: map[string]interface{}{"combatBoost": 0.1}},
	},
	"rare": {
		{ID: "defense_master", Name: "防御大师", Description: "防御力提升10%", Type: "rare", Effects: map[string]interface{}{"defense": 0.1}},
		{ID: "crit_mastery", Name: "会心精通", Description: "暴击率+10%, 爆伤+20%", Type: "rare", Effects: map[string]interface{}{"critRate": 0.1, "critDamageBoost": 0.2}},
		{ID: "dodge_master", Name: "无影", Description: "闪避率提升10%", Type: "rare", Effects: map[string]interface{}{"dodgeRate": 0.1}},
		{ID: "combo_master", Name: "连击精通", Description: "连击率提升10%", Type: "rare", Effects: map[string]interface{}{"comboRate": 0.1}},
		{ID: "vampire_master", Name: "血魔", Description: "吸血率提升5%", Type: "rare", Effects: map[string]interface{}{"vampireRate": 0.05}},
		{ID: "stun_master", Name: "震慑", Description: "眩晕率提升5%", Type: "rare", Effects: map[string]interface{}{"stunRate": 0.05}},
	},
	"epic": {
		{ID: "ultimate_power", Name: "极限突破", Description: "所有战斗属性提升50%", Type: "epic", Effects: map[string]interface{}{"combatBoost": 0.5, "finalDamageBoost": 0.5}},
		{ID: "divine_protection", Name: "天道庇护", Description: "最终减伤提升30%", Type: "epic", Effects: map[string]interface{}{"finalDamageReduce": 0.3}},
		{ID: "combat_master", Name: "战斗大师", Description: "战斗属性+25%, 抗性+25%", Type: "epic", Effects: map[string]interface{}{"combatBoost": 0.25, "resistanceBoost": 0.25}},
		{ID: "immortal_body", Name: "不朽之躯", Description: "生命上限+100%, 减伤+20%", Type: "epic", Effects: map[string]interface{}{"health": 1.0, "finalDamageReduce": 0.2}},
		{ID: "celestial_might", Name: "天人合一", Description: "战斗属性+40%, 血量+50%", Type: "epic", Effects: map[string]interface{}{"combatBoost": 0.4, "health": 0.5}},
		{ID: "battle_sage_supreme", Name: "战圣至尊", Description: "暴击率+40%, 爆伤+80%, 伤害+20%", Type: "epic", Effects: map[string]interface{}{"critRate": 0.4, "critDamageBoost": 0.8, "finalDamageBoost": 0.2}},
	},
}

// DifficultyConfig 难度配置
var DifficultyConfig = map[string]DifficultyModifier{
	"easy":   {HealthMod: 0.8, DamageMod: 0.8, RewardMod: 0.8},
	"normal": {HealthMod: 1.0, DamageMod: 1.0, RewardMod: 1.0},
	"hard":   {HealthMod: 1.2, DamageMod: 1.2, RewardMod: 1.5},
	"expert": {HealthMod: 1.5, DamageMod: 1.5, RewardMod: 2.0},
}

// GetBuffConfig 根据ID获取增益配置
func GetBuffConfig(buffID string) *BuffConfig {
	for _, poolConfigs := range BuffConfigPool {
		for i, cfg := range poolConfigs {
			if cfg.ID == buffID {
				return &poolConfigs[i]
			}
		}
	}
	return nil
}

// GetDifficultyModifier 获取难度修饰符
func GetDifficultyModifier(difficulty string) DifficultyModifier {
	if mod, ok := DifficultyConfig[difficulty]; ok {
		return mod
	}
	return DifficultyConfig["normal"]
}
