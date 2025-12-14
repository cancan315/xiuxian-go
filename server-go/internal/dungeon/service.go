package dungeon

import (
	"fmt"
	"math"
	"math/rand"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// DungeonService 秘境服务
type DungeonService struct {
	userID uint
}

// NewDungeonService 创建秘境服务实例
func NewDungeonService(userID uint) *DungeonService {
	return &DungeonService{
		userID: userID,
	}
}

// 难度修饰符
var difficultyModifiers = map[string]DifficultyModifier{
	"easy":   {HealthMod: 0.8, DamageMod: 0.8, RewardMod: 0.8},
	"normal": {HealthMod: 1.0, DamageMod: 1.0, RewardMod: 1.0},
	"hard":   {HealthMod: 1.2, DamageMod: 1.2, RewardMod: 1.5},
	"expert": {HealthMod: 1.5, DamageMod: 1.5, RewardMod: 2.0},
}

// 增益配置池
var buffConfigs = map[string][]BuffConfig{
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

// GetRandomBuffs 获取随机增益选项
func (s *DungeonService) GetRandomBuffs(floor int) ([]BuffOption, error) {
	// 计算概率
	commonChance := 0.7
	rareChance := 0.25
	epicChance := 0.05

	if floor%10 == 0 {
		commonChance = 0.5
		rareChance = 0.3
		epicChance = 0.2
	} else if floor%5 == 0 {
		commonChance = 0.5
		rareChance = 0.35
		epicChance = 0.15
	}

	_ = commonChance  // 标记为已使用（虽然当前实现中暂未直接使用）
	_ = rareChance
	_ = epicChance

	var options []BuffOption
	usedIds := make(map[string]bool)

	for len(options) < 3 {
		randVal := rand.Float64()
		var pool string

		if randVal < epicChance {
			pool = "epic"
		} else if randVal < epicChance+rareChance {
			pool = "rare"
		} else {
			pool = "common"
		}

		// 从池中随机选择
		configs := buffConfigs[pool]
		var available []BuffConfig

		for _, cfg := range configs {
			if !usedIds[cfg.ID] {
				available = append(available, cfg)
			}
		}

		if len(available) > 0 {
			idx := rand.Intn(len(available))
			cfg := available[idx]
			usedIds[cfg.ID] = true

			option := BuffOption{
				ID:          cfg.ID,
				Name:        cfg.Name,
				Description: cfg.Description,
				Type:        cfg.Type,
				Effect:      cfg.Effects,
			}
			options = append(options, option)
		} else {
			// 从所有池中选择
			var allAvailable []BuffConfig
			for _, poolConfigs := range buffConfigs {
				for _, cfg := range poolConfigs {
					if !usedIds[cfg.ID] {
						allAvailable = append(allAvailable, cfg)
					}
				}
			}

			if len(allAvailable) > 0 {
				idx := rand.Intn(len(allAvailable))
				cfg := allAvailable[idx]
				usedIds[cfg.ID] = true

				option := BuffOption{
					ID:          cfg.ID,
					Name:        cfg.Name,
					Description: cfg.Description,
					Type:        cfg.Type,
					Effect:      cfg.Effects,
				}
				options = append(options, option)
			} else {
				break
			}
		}
	}

	return options, nil
}

// SelectBuff 选择增益
func (s *DungeonService) SelectBuff(buffID string) (map[string]interface{}, error) {
	// 查找增益配置
	var selectedConfig *BuffConfig

	for _, poolConfigs := range buffConfigs {
		for _, cfg := range poolConfigs {
			if cfg.ID == buffID {
				selectedConfig = &cfg
				break
			}
		}
		if selectedConfig != nil {
			break
		}
	}

	if selectedConfig == nil {
		return nil, fmt.Errorf("增益不存在: %s", buffID)
	}

	return map[string]interface{}{
		"id":          selectedConfig.ID,
		"name":        selectedConfig.Name,
		"description": selectedConfig.Description,
		"effects":     selectedConfig.Effects,
	}, nil
}

// StartFight 开始战斗
func (s *DungeonService) StartFight(floor int, difficulty string) (*FightResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	modifier, ok := difficultyModifiers[difficulty]
	if !ok {
		return nil, fmt.Errorf("难度不存在: %s", difficulty)
	}

	// 计算敌人属性
	playerAttack := float64(user.Level) * 5 // 简化计算
	playerHealth := float64(user.Level) * 20

	_ = float64(user.Level) * 2 // playerDefense 暂未使用
	_ = math.Floor(playerHealth * modifier.HealthMod * (1 + float64(floor)*0.05)) // enemyHealth 暂未使用
	_ = math.Floor(playerAttack * modifier.DamageMod * (1 + float64(floor)*0.05))  // enemyDamage 暂未使用

	// 简化的战斗逻辑
	// 实际应该使用完整的战斗系统，这里仅为演示
	victory := rand.Float64() < 0.6 // 60%胜率

	result := &FightResult{
		Success: true,
		Victory: victory,
		Floor:   floor,
		Message: fmt.Sprintf("战斗%s！", map[bool]string{true: "胜利", false: "失败"}[victory]),
		Rewards: []interface{}{
			map[string]interface{}{
				"type":   "spirit_stone",
				"amount": int(100 * modifier.RewardMod),
			},
		},
	}

	// 保存战斗统计
	if victory {
		user.SpiritStones += int(100 * modifier.RewardMod)
		if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
			return nil, fmt.Errorf("更新用户灵石失败: %w", err)
		}
	}

	return result, nil
}

// EndDungeon 结束秘境
func (s *DungeonService) EndDungeon(floor int, victory bool) (map[string]interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	totalReward := int(floor * 50) // 基础奖励
	if victory {
		user.SpiritStones += totalReward
	}

	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, fmt.Errorf("更新用户数据失败: %w", err)
	}

	return map[string]interface{}{
		"floor":        floor,
		"totalReward":  totalReward,
		"victory":      victory,
		"spiritStones": user.SpiritStones,
	}, nil
}
