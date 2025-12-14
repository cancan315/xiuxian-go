package dungeon

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// DungeonService 秘境服务
type DungeonService struct {
	userID        uint
	selectedBuffs []string               // 已选增益ID
	buffEffects   map[string]interface{} // 已选增益的效果应用
}

// NewDungeonService 创建秘境服务实例
func NewDungeonService(userID uint) *DungeonService {
	return &DungeonService{
		userID:        userID,
		selectedBuffs: []string{},
		buffEffects:   make(map[string]interface{}),
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

	_ = commonChance // 标记为已使用（虽然当前实现中暂未直接使用）
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

// CalculateDamage 计算伤害 - 对应前端 calculateDamage 方法
func (s *DungeonService) CalculateDamage(attacker *CombatStats, defender *CombatStats) *DamageResult {
	result := &DamageResult{}

	// 基础伤害 = damage * (1 + combatBoost)
	damage := math.Abs(attacker.Damage * (1 + attacker.CombatBoost))

	// 暴击判定：finalCritRate = MAX(0, MIN(0.8, critRate * (1 + combatBoost) - target.critResist * (1 + target.resistanceBoost)))
	finalCritRate := math.Max(0, math.Min(0.8,
		attacker.CritRate*(1+attacker.CombatBoost)-defender.CritResist*(1+defender.ResistanceBoost)))
	if rand.Float64() < finalCritRate {
		damage *= 1.5 + attacker.CritDamageBoost
		result.IsCrit = true
	}

	// 连击判定：finalComboRate = MAX(0, MIN(0.8, comboRate * (1 + combatBoost) - target.comboResist))
	finalComboRate := math.Max(0, math.Min(0.8,
		attacker.ComboRate*(1+attacker.CombatBoost)-defender.ComboResist))
	if rand.Float64() < finalComboRate {
		damage *= 1.3
		result.IsCombo = true
	}

	// 吸血判定：finalVampireRate = MAX(0, MIN(0.8, vampireRate * (1 + combatBoost) - target.vampireResist))
	finalVampireRate := math.Max(0, math.Min(0.8,
		attacker.VampireRate*(1+attacker.CombatBoost)-defender.VampireResist))
	if rand.Float64() < finalVampireRate {
		result.IsVampire = true
	}

	// 眩晕判定：finalStunRate = MAX(0, MIN(0.8, stunRate * (1 + combatBoost) - target.stunResist))
	finalStunRate := math.Max(0, math.Min(0.8,
		attacker.StunRate*(1+attacker.CombatBoost)-defender.StunResist))
	if rand.Float64() < finalStunRate {
		result.IsStun = true
	}

	// 最终伤害 = damage * (1 + finalDamageBoost)
	damage *= 1 + attacker.FinalDamageBoost

	result.Damage = math.Abs(damage)
	return result
}

// CalculateDamageReduction 计算伤害减免 - 对应前端 calculateDamageReduction 方法
func (s *DungeonService) CalculateDamageReduction(defender *CombatStats, incomingDamage float64, attacker *CombatStats) float64 {
	damage := math.Abs(incomingDamage)

	// 防御减伤 = damage * 100 / (100 + defense * (1 + combatBoost))
	effectiveDefense := defender.Defense * (1 + defender.CombatBoost)
	damage *= 100 / (100 + effectiveDefense)

	// 如果是暴击伤害，应用暴击伤害减免 - damage *= (1 - critDamageReduce)
	// 注：前端在这里检查attackerStats.isCrit，我们直接在这里判断
	if attacker != nil && attacker.CritRate > 0 {
		damage *= 1 - defender.CritDamageReduce
	}

	// 最终减伤 = damage *= (1 - finalDamageReduce)
	damage *= 1 - defender.FinalDamageReduce

	return math.Abs(damage)
}

// TakeDamage 被伤害 - 对应前端 takeDamage 方法
func (s *DungeonService) TakeDamage(defender *CombatStats, currentHealth float64, incomingDamage float64, source *CombatStats) *TakeDamageResult {
	result := &TakeDamageResult{}

	// 闪避判定：actualDodgeRate = MAX(0, MIN(0.8, dodgeRate - (source ? source.dodgeResist : 0)))
	var actualDodgeRate float64
	if source != nil {
		actualDodgeRate = math.Max(0, math.Min(0.8, defender.DodgeRate-source.DodgeResist))
	} else {
		actualDodgeRate = math.Max(0, math.Min(0.8, defender.DodgeRate))
	}

	// 闪避成功
	if rand.Float64() < actualDodgeRate {
		result.Dodged = true
		result.Damage = 0
		result.CurrentHealth = currentHealth
		return result
	}

	// 计算实际伤害
	reducedDamage := s.CalculateDamageReduction(defender, incomingDamage, source)
	newHealth := math.Max(0, currentHealth-reducedDamage)

	// 反击判定：finalCounterRate = MAX(0, MIN(0.8, counterRate - source.counterResist))
	if source != nil {
		finalCounterRate := math.Max(0, math.Min(0.8, defender.CounterRate-source.CounterResist))
		if rand.Float64() < finalCounterRate {
			result.IsCounter = true
		}
	}

	result.Dodged = false
	result.Damage = reducedDamage
	result.CurrentHealth = newHealth
	result.IsDead = newHealth <= 0

	return result
}

// ApplyBuffEffectsToStats 将已选增益效果应用到战斗属性
func (s *DungeonService) ApplyBuffEffectsToStats(stats *CombatStats) *CombatStats {
	for key, value := range s.buffEffects {
		if val, ok := value.(float64); ok {
			switch key {
			case "health":
				stats.Health *= (1 + val)
				stats.MaxHealth *= (1 + val)
			case "finalDamageBoost", "damage":
				stats.Damage *= (1 + val)
			case "defense":
				stats.Defense *= (1 + val)
			case "speed":
				stats.Speed *= (1 + val)
			case "critRate":
				stats.CritRate += val
			case "dodgeRate":
				stats.DodgeRate += val
			case "comboRate":
				stats.ComboRate += val
			case "vampireRate":
				stats.VampireRate += val
			case "stunRate":
				stats.StunRate += val
			case "critDamageBoost":
				stats.CritDamageBoost += val
			case "finalDamageReduce":
				stats.FinalDamageReduce += val
			case "combatBoost":
				stats.CombatBoost += val
			case "resistanceBoost":
				stats.ResistanceBoost += val
			}
		}
	}
	return stats
}

// StartFight 开始战斗 - 完全按照前端逻辑实现
func (s *DungeonService) StartFight(floor int, difficulty string) (*FightResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	modifier, ok := difficultyModifiers[difficulty]
	if !ok {
		return nil, fmt.Errorf("难度不存在: %s", difficulty)
	}

	// 解析战斗属性JSON
	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(user.CombatAttributes, &combatAttrs); err != nil {
		return nil, fmt.Errorf("解析战斗属性失败: %w", err)
	}

	// 解析特殊属性JSON
	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(user.SpecialAttributes, &specialAttrs); err != nil {
		return nil, fmt.Errorf("解析特殊属性失败: %w", err)
	}

	// 从数据库获取玩家战斗属性
	playerStats := &CombatStats{
		Health:            getFloatValue(combatAttrs, "health", float64(user.Level)*20),
		MaxHealth:         getFloatValue(combatAttrs, "maxHealth", float64(user.Level)*20),
		Damage:            getFloatValue(combatAttrs, "damage", float64(user.Level)*5),
		Defense:           getFloatValue(combatAttrs, "defense", float64(user.Level)*2),
		Speed:             getFloatValue(combatAttrs, "speed", float64(user.Level)),
		CritRate:          getFloatValue(combatAttrs, "critRate", 0.05),
		ComboRate:         getFloatValue(combatAttrs, "comboRate", 0.0),
		CounterRate:       getFloatValue(combatAttrs, "counterRate", 0.0),
		StunRate:          getFloatValue(combatAttrs, "stunRate", 0.0),
		DodgeRate:         getFloatValue(combatAttrs, "dodgeRate", 0.05),
		VampireRate:       getFloatValue(combatAttrs, "vampireRate", 0.0),
		CritResist:        getFloatValue(combatAttrs, "critResist", 0.0),
		ComboResist:       getFloatValue(combatAttrs, "comboResist", 0.0),
		CounterResist:     getFloatValue(combatAttrs, "counterResist", 0.0),
		StunResist:        getFloatValue(combatAttrs, "stunResist", 0.0),
		DodgeResist:       getFloatValue(combatAttrs, "dodgeResist", 0.0),
		VampireResist:     getFloatValue(combatAttrs, "vampireResist", 0.0),
		HealBoost:         getFloatValue(specialAttrs, "healBoost", 0.0),
		CritDamageBoost:   getFloatValue(specialAttrs, "critDamageBoost", 0.5),
		CritDamageReduce:  getFloatValue(specialAttrs, "critDamageReduce", 0.0),
		FinalDamageBoost:  getFloatValue(specialAttrs, "finalDamageBoost", 0.0),
		FinalDamageReduce: getFloatValue(specialAttrs, "finalDamageReduce", 0.0),
		CombatBoost:       getFloatValue(specialAttrs, "combatBoost", 0.0),
		ResistanceBoost:   getFloatValue(specialAttrs, "resistanceBoost", 0.0),
	}

	// 应用已选择的增益效果
	playerStats = s.ApplyBuffEffectsToStats(playerStats)

	// 应用难度修饰符
	playerStats.Health *= modifier.HealthMod
	playerStats.MaxHealth *= modifier.HealthMod
	playerStats.Damage *= modifier.DamageMod

	// 初始化敌人战斗属性（基于层数递增）
	enemyStats := &CombatStats{
		Health:            playerStats.Health * (1 + float64(floor)*0.05),
		MaxHealth:         playerStats.Health * (1 + float64(floor)*0.05),
		Damage:            playerStats.Damage * (1 + float64(floor)*0.05),
		Defense:           playerStats.Defense * modifier.DamageMod * (1 + float64(floor)*0.03),
		Speed:             playerStats.Speed * (1 + float64(floor)*0.02),
		CritRate:          0.05 + float64(floor)*0.02,
		ComboRate:         0.03 + float64(floor)*0.02,
		CounterRate:       0.03 + float64(floor)*0.02,
		StunRate:          0.02 + float64(floor)*0.01,
		DodgeRate:         0.05 + float64(floor)*0.02,
		VampireRate:       0.02 + float64(floor)*0.01,
		CritResist:        0.02 + float64(floor)*0.01,
		ComboResist:       0.02 + float64(floor)*0.01,
		CounterResist:     0.02 + float64(floor)*0.01,
		StunResist:        0.02 + float64(floor)*0.01,
		DodgeResist:       0.02 + float64(floor)*0.01,
		VampireResist:     0.02 + float64(floor)*0.01,
		HealBoost:         0.05 + float64(floor)*0.02,
		CritDamageBoost:   0.2 + float64(floor)*0.1,
		CritDamageReduce:  0.1 + float64(floor)*0.05,
		FinalDamageBoost:  0.05 + float64(floor)*0.02,
		FinalDamageReduce: 0.05 + float64(floor)*0.02,
		CombatBoost:       0.03 + float64(floor)*0.02,
		ResistanceBoost:   0.03 + float64(floor)*0.02,
	}

	// 执行战斗模拟
	playerCurrentHealth := playerStats.MaxHealth
	enemyCurrentHealth := enemyStats.MaxHealth
	rounds := 0
	maxRounds := 100
	battleLog := []string{}

	for rounds < maxRounds && playerCurrentHealth > 0 && enemyCurrentHealth > 0 {
		rounds++

		// 根据速度决定攻击顺序
		playerSpeed := playerStats.Speed * (1 + playerStats.CombatBoost)
		enemySpeed := enemyStats.Speed * (1 + enemyStats.CombatBoost)

		if playerSpeed >= enemySpeed {
			// 玩家先手
			dmgResult := s.CalculateDamage(playerStats, enemyStats)
			playerTakeDmgResult := s.TakeDamage(enemyStats, enemyCurrentHealth, dmgResult.Damage, playerStats)
			enemyCurrentHealth = playerTakeDmgResult.CurrentHealth

			battleLog = append(battleLog, fmt.Sprintf("第%d回合：玩家造成%.1f伤害", rounds, dmgResult.Damage))

			if playerTakeDmgResult.IsDead {
				break
			}

			// 敌人反击（如果没被眩晕）
			if !dmgResult.IsStun {
				dmgResult2 := s.CalculateDamage(enemyStats, playerStats)
				enemyTakeDmgResult := s.TakeDamage(playerStats, playerCurrentHealth, dmgResult2.Damage, enemyStats)
				playerCurrentHealth = enemyTakeDmgResult.CurrentHealth

				battleLog = append(battleLog, fmt.Sprintf("第%d回合：敌人造成%.1f伤害", rounds, dmgResult2.Damage))

				if enemyTakeDmgResult.IsDead {
					break
				}
			}
		} else {
			// 敌人先手
			dmgResult := s.CalculateDamage(enemyStats, playerStats)
			enemyTakeDmgResult := s.TakeDamage(playerStats, playerCurrentHealth, dmgResult.Damage, enemyStats)
			playerCurrentHealth = enemyTakeDmgResult.CurrentHealth

			battleLog = append(battleLog, fmt.Sprintf("第%d回合：敌人造成%.1f伤害", rounds, dmgResult.Damage))

			if enemyTakeDmgResult.IsDead {
				break
			}

			// 玩家反击（如果没被眩晕）
			if !dmgResult.IsStun {
				dmgResult2 := s.CalculateDamage(playerStats, enemyStats)
				playerTakeDmgResult := s.TakeDamage(enemyStats, enemyCurrentHealth, dmgResult2.Damage, playerStats)
				enemyCurrentHealth = playerTakeDmgResult.CurrentHealth

				battleLog = append(battleLog, fmt.Sprintf("第%d回合：玩家造成%.1f伤害", rounds, dmgResult2.Damage))

				if playerTakeDmgResult.IsDead {
					break
				}
			}
		}
	}

	// 判定胜负
	victory := playerCurrentHealth > 0

	// 计算奖励
	rewardAmount := int(100 * modifier.RewardMod * (1 + float64(floor)*0.1))
	if !victory {
		rewardAmount = int(float64(rewardAmount) * 0.3)
	}

	// 更新玩家灵石
	user.SpiritStones += rewardAmount
	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, fmt.Errorf("更新用户灵石失败: %w", err)
	}

	result := &FightResult{
		Success: true,
		Victory: victory,
		Floor:   floor,
		Message: fmt.Sprintf("战斗%s！\n玩家血量: %.1f/%.1f\n敌人血量: %.1f/%.1f\n回合数: %d\n战斗日志: %v",
			map[bool]string{true: "胜利", false: "失败"}[victory],
			math.Max(0, playerCurrentHealth), playerStats.MaxHealth,
			math.Max(0, enemyCurrentHealth), enemyStats.MaxHealth,
			rounds,
			battleLog),
		Rewards: []interface{}{
			map[string]interface{}{
				"type":   "spirit_stone",
				"amount": rewardAmount,
			},
		},
	}

	return result, nil
}

// SelectBuffAndApplyEffects 选择增益并记录效果
func (s *DungeonService) SelectBuffAndApplyEffects(buffID string) (map[string]interface{}, error) {
	buff, err := s.SelectBuff(buffID)
	if err != nil {
		return nil, err
	}

	// 记录已选增益
	s.selectedBuffs = append(s.selectedBuffs, buffID)

	// 提取效果并应用
	if effects, ok := buff["effects"].(map[string]interface{}); ok {
		for k, v := range effects {
			s.buffEffects[k] = v
		}
	}

	return buff, nil
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

// getFloatValue 从map中安全获取float64值，如果不存在则返回默认值
func getFloatValue(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
		if i, ok := val.(int); ok {
			return float64(i)
		}
		if s, ok := val.(string); ok {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
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
