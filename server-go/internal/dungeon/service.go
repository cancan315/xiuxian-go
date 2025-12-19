package dungeon

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
)

// DungeonService 秘境服务
type DungeonService struct {
	userID        uint
	selectedBuffs []string               // 已选增益ID
	buffEffects   map[string]interface{} // 已选增益的效果应用
	rewardAmount  int                    // 战斗奖励金额
}

// NewDungeonService 创建秘境服务实例
func NewDungeonService(userID uint) *DungeonService {
	return &DungeonService{
		userID:        userID,
		selectedBuffs: []string{},
		buffEffects:   make(map[string]interface{}),
	}
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
		configs := BuffConfigPool[pool]
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
			for _, poolConfigs := range BuffConfigPool {
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

// RoundResult 单回合战斗结果
type RoundResult struct {
	Round        int      `json:"round"`
	PlayerHealth float64  `json:"playerHealth"`
	EnemyHealth  float64  `json:"enemyHealth"`
	Log          []string `json:"log"`
	BattleEnded  bool     `json:"battleEnded"`
	Victory      bool     `json:"victory"`
}

// StartFight 开始战斗 - 完全按照前端逻辑实现
func (s *DungeonService) StartFight(floor int, difficulty string) (*FightResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	modifier := GetDifficultyModifier(difficulty)

	// 解析基础属性JSON (attack, health, defense, speed)
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &baseAttrs); err != nil {
		return nil, fmt.Errorf("解析基础属性失败: %w", err)
	}

	// 解析战斗属性JSON (critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate)
	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(user.CombatAttributes, &combatAttrs); err != nil {
		return nil, fmt.Errorf("解析战斗属性失败: %w", err)
	}

	// 解析战斗抗性JSON
	var combatResist map[string]interface{}
	if err := json.Unmarshal(user.CombatResistance, &combatResist); err != nil {
		return nil, fmt.Errorf("解析战斗抗性失败: %w", err)
	}

	// 解析特殊属性JSON
	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(user.SpecialAttributes, &specialAttrs); err != nil {
		return nil, fmt.Errorf("解析特殊属性失败: %w", err)
	}

	// 从数据库中直接读取属性值
	var health, attack, defense, speed, critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate float64
	var critResist, comboResist, counterResist, stunResist, dodgeResist, vampireResist float64
	var healBoost, critDamageBoost, critDamageReduce, finalDamageBoost, finalDamageReduce, combatBoost, resistanceBoost float64

	// 从 baseAttributes 读取基础属性
	if v, ok := baseAttrs["health"].(float64); ok {
		health = v
	}
	if v, ok := baseAttrs["attack"].(float64); ok {
		attack = v
	}
	if v, ok := baseAttrs["defense"].(float64); ok {
		defense = v
	}
	if v, ok := baseAttrs["speed"].(float64); ok {
		speed = v
	}

	// 从 combatAttributes 读取战斗属性
	if v, ok := combatAttrs["critRate"].(float64); ok {
		critRate = v
	}
	if v, ok := combatAttrs["comboRate"].(float64); ok {
		comboRate = v
	}
	if v, ok := combatAttrs["counterRate"].(float64); ok {
		counterRate = v
	}
	if v, ok := combatAttrs["stunRate"].(float64); ok {
		stunRate = v
	}
	if v, ok := combatAttrs["dodgeRate"].(float64); ok {
		dodgeRate = v
	}
	if v, ok := combatAttrs["vampireRate"].(float64); ok {
		vampireRate = v
	}

	// 从 combatResistance 读取战斗抗性
	if v, ok := combatResist["critResist"].(float64); ok {
		critResist = v
	}
	if v, ok := combatResist["comboResist"].(float64); ok {
		comboResist = v
	}
	if v, ok := combatResist["counterResist"].(float64); ok {
		counterResist = v
	}
	if v, ok := combatResist["stunResist"].(float64); ok {
		stunResist = v
	}
	if v, ok := combatResist["dodgeResist"].(float64); ok {
		dodgeResist = v
	}
	if v, ok := combatResist["vampireResist"].(float64); ok {
		vampireResist = v
	}

	// 从 specialAttributes 读取特殊属性
	if v, ok := specialAttrs["healBoost"].(float64); ok {
		healBoost = v
	}
	if v, ok := specialAttrs["critDamageBoost"].(float64); ok {
		critDamageBoost = v
	}
	if v, ok := specialAttrs["critDamageReduce"].(float64); ok {
		critDamageReduce = v
	}
	if v, ok := specialAttrs["finalDamageBoost"].(float64); ok {
		finalDamageBoost = v
	}
	if v, ok := specialAttrs["finalDamageReduce"].(float64); ok {
		finalDamageReduce = v
	}
	if v, ok := specialAttrs["combatBoost"].(float64); ok {
		combatBoost = v
	}
	if v, ok := specialAttrs["resistanceBoost"].(float64); ok {
		resistanceBoost = v
	}

	// 初始化玩家战斗属性
	playerStats := &CombatStats{
		Health:            health,      // 当前生命值
		MaxHealth:         health,      // 最大生命值 = 生命值
		Damage:            attack,      // 伤害值 = attack
		Defense:           defense,     // 防御力
		Speed:             speed,       // 速度（决定先手）
		CritRate:          critRate,    // 暴击率
		ComboRate:         comboRate,   // 连击率
		CounterRate:       counterRate, // 反击率
		StunRate:          stunRate,    // 眩晕率
		DodgeRate:         dodgeRate,   // 闪避率
		VampireRate:       vampireRate, // 吸血率
		CritResist:        critResist,  // 抗暴击
		ComboResist:       comboResist, // 抗连击
		CounterResist:     counterResist,
		StunResist:        stunResist,
		DodgeResist:       dodgeResist,
		VampireResist:     vampireResist,
		HealBoost:         healBoost,
		CritDamageBoost:   critDamageBoost,
		CritDamageReduce:  critDamageReduce,
		FinalDamageBoost:  finalDamageBoost,
		FinalDamageReduce: finalDamageReduce,
		CombatBoost:       combatBoost,
		ResistanceBoost:   resistanceBoost,
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

	// 执行战斗模拟（完整的回合制逻辑）
	playerCurrentHealth := playerStats.MaxHealth
	enemyCurrentHealth := enemyStats.MaxHealth
	rounds := 0
	maxRounds := 100
	battleLog := []string{}

	for rounds < maxRounds && playerCurrentHealth > 0 && enemyCurrentHealth > 0 {
		rounds++

		// 根据速度决定先手和后手
		playerSpeed := playerStats.Speed * (1 + playerStats.CombatBoost)
		enemySpeed := enemyStats.Speed * (1 + enemyStats.CombatBoost)

		if playerSpeed >= enemySpeed {
			// ==================== 玩家先手 ====================
			// 第1步：玩家攻击敌人
			playerDmgResult := s.CalculateDamage(playerStats, enemyStats)
			enemyTakeDmgResult := s.TakeDamage(enemyStats, enemyCurrentHealth, playerDmgResult.Damage, playerStats)
			enemyCurrentHealth = enemyTakeDmgResult.CurrentHealth

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：玩家造成%.1f伤害", rounds, playerDmgResult.Damage)
			if playerDmgResult.IsCrit {
				logMsg += "(暴击)"
			}
			if playerDmgResult.IsCombo {
				logMsg += "(连击)"
			}
			if playerDmgResult.IsStun {
				logMsg += "(眩晕)"
			}
			battleLog = append(battleLog, logMsg)

			// 第2步：检查敌人是否死亡
			if enemyTakeDmgResult.IsDead {
				battleLog = append(battleLog, "敌人已被击败！")
				break
			}

			// 第3步：敌人反击（如果没被眩晕）
			if !playerDmgResult.IsStun {
				enemyDmgResult := s.CalculateDamage(enemyStats, playerStats)
				playerTakeDmgResult := s.TakeDamage(playerStats, playerCurrentHealth, enemyDmgResult.Damage, enemyStats)
				playerCurrentHealth = playerTakeDmgResult.CurrentHealth

				// 记录日志
				logMsg := fmt.Sprintf("第%d回合：敌人造成%.1f伤害", rounds, enemyDmgResult.Damage)
				if enemyDmgResult.IsCrit {
					logMsg += "(暴击)"
				}
				if enemyDmgResult.IsCombo {
					logMsg += "(连击)"
				}
				if playerTakeDmgResult.IsCounter {
					logMsg += "(玩家反击)"
				}
				battleLog = append(battleLog, logMsg)

				// 第4步：检查玩家是否死亡
				if playerTakeDmgResult.IsDead {
					battleLog = append(battleLog, "玩家已被击败！")
					break
				}
			} else {
				battleLog = append(battleLog, "敌人被眩晕，无法行动！")
			}
		} else {
			// ==================== 敌人先手 ====================
			// 第1步：敌人攻击玩家
			enemyDmgResult := s.CalculateDamage(enemyStats, playerStats)
			playerTakeDmgResult := s.TakeDamage(playerStats, playerCurrentHealth, enemyDmgResult.Damage, enemyStats)
			playerCurrentHealth = playerTakeDmgResult.CurrentHealth

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：敌人造成%.1f伤害", rounds, enemyDmgResult.Damage)
			if enemyDmgResult.IsCrit {
				logMsg += "(暴击)"
			}
			if enemyDmgResult.IsCombo {
				logMsg += "(连击)"
			}
			if enemyDmgResult.IsStun {
				logMsg += "(眩晕)"
			}
			battleLog = append(battleLog, logMsg)

			// 第2步：检查玩家是否死亡
			if playerTakeDmgResult.IsDead {
				battleLog = append(battleLog, "玩家已被击败！")
				break
			}

			// 第3步：玩家反击（如果没被眩晕）
			if !enemyDmgResult.IsStun {
				playerDmgResult := s.CalculateDamage(playerStats, enemyStats)
				enemyTakeDmgResult := s.TakeDamage(enemyStats, enemyCurrentHealth, playerDmgResult.Damage, playerStats)
				enemyCurrentHealth = enemyTakeDmgResult.CurrentHealth

				// 记录日志
				logMsg := fmt.Sprintf("第%d回合：玩家造成%.1f伤害", rounds, playerDmgResult.Damage)
				if playerDmgResult.IsCrit {
					logMsg += "(暴击)"
				}
				if playerDmgResult.IsCombo {
					logMsg += "(连击)"
				}
				if enemyTakeDmgResult.IsCounter {
					logMsg += "(敌人反击)"
				}
				battleLog = append(battleLog, logMsg)

				// 第4步：检查敌人是否死亡
				if enemyTakeDmgResult.IsDead {
					battleLog = append(battleLog, "敌人已被击败！")
					break
				}
			} else {
				battleLog = append(battleLog, "玩家被眩晕，无法行动！")
			}
		}
	}

	// 检查是否超出回合数限制
	if rounds >= maxRounds && playerCurrentHealth > 0 && enemyCurrentHealth > 0 {
		battleLog = append(battleLog, "战斗超出最大回合数，判定为失败！")
		playerCurrentHealth = 0 // 强制失败
	}

	// 判定胜负
	victory := playerCurrentHealth > 0

	// 计算奖励
	rewardAmount := int(100 * modifier.RewardMod * (1 + float64(floor)*0.1))
	if !victory {
		rewardAmount = int(float64(rewardAmount) * 0.3)
	}

	// 保存奖励䯸朋加在服务对象中，䭜EndDungeon使用
	s.rewardAmount = rewardAmount

	// 暴馨：事实上不立即更新数据库，每战斗结束后统一更新
	// 残禾EndDungeon来一次性更新玩家抬石

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

// StartFightStreaming 流式战斗 - 支持逐回合执行（用于每秒一回合的实时显示）
// 返回初始化的战斗状态和战斗实体，由调用者逐步调用 ExecuteFightRound
func (s *DungeonService) StartFightStreaming(floor int, difficulty string) (*CombatStats, *CombatStats, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, nil, fmt.Errorf("用户不存在: %w", err)
	}

	modifier := GetDifficultyModifier(difficulty)

	// 解析基础属性JSON (attack, health, defense, speed)
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &baseAttrs); err != nil {
		return nil, nil, fmt.Errorf("解析基础属性失败: %w", err)
	}

	// 解析战斗属性JSON (critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate)
	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(user.CombatAttributes, &combatAttrs); err != nil {
		return nil, nil, fmt.Errorf("解析战斗属性失败: %w", err)
	}

	// 解析战斗抗性JSON
	var combatResist map[string]interface{}
	if err := json.Unmarshal(user.CombatResistance, &combatResist); err != nil {
		return nil, nil, fmt.Errorf("解析战斗抗性失败: %w", err)
	}

	// 解析特殊属性JSON
	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(user.SpecialAttributes, &specialAttrs); err != nil {
		return nil, nil, fmt.Errorf("解析特殊属性失败: %w", err)
	}

	// 从数据库中直接读取属性值
	var health, attack, defense, speed, critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate float64
	var critResist, comboResist, counterResist, stunResist, dodgeResist, vampireResist float64
	var healBoost, critDamageBoost, critDamageReduce, finalDamageBoost, finalDamageReduce, combatBoost, resistanceBoost float64

	// 从 baseAttributes 读取基础属性
	if v, ok := baseAttrs["health"].(float64); ok {
		health = v
	}
	if v, ok := baseAttrs["attack"].(float64); ok {
		attack = v
	}
	if v, ok := baseAttrs["defense"].(float64); ok {
		defense = v
	}
	if v, ok := baseAttrs["speed"].(float64); ok {
		speed = v
	}

	// 从 combatAttributes 读取战斗属性
	if v, ok := combatAttrs["critRate"].(float64); ok {
		critRate = v
	}
	if v, ok := combatAttrs["comboRate"].(float64); ok {
		comboRate = v
	}
	if v, ok := combatAttrs["counterRate"].(float64); ok {
		counterRate = v
	}
	if v, ok := combatAttrs["stunRate"].(float64); ok {
		stunRate = v
	}
	if v, ok := combatAttrs["dodgeRate"].(float64); ok {
		dodgeRate = v
	}
	if v, ok := combatAttrs["vampireRate"].(float64); ok {
		vampireRate = v
	}

	// 从 combatResistance 读取战斗抗性
	if v, ok := combatResist["critResist"].(float64); ok {
		critResist = v
	}
	if v, ok := combatResist["comboResist"].(float64); ok {
		comboResist = v
	}
	if v, ok := combatResist["counterResist"].(float64); ok {
		counterResist = v
	}
	if v, ok := combatResist["stunResist"].(float64); ok {
		stunResist = v
	}
	if v, ok := combatResist["dodgeResist"].(float64); ok {
		dodgeResist = v
	}
	if v, ok := combatResist["vampireResist"].(float64); ok {
		vampireResist = v
	}

	// 从 specialAttributes 读取特殊属性
	if v, ok := specialAttrs["healBoost"].(float64); ok {
		healBoost = v
	}
	if v, ok := specialAttrs["critDamageBoost"].(float64); ok {
		critDamageBoost = v
	}
	if v, ok := specialAttrs["critDamageReduce"].(float64); ok {
		critDamageReduce = v
	}
	if v, ok := specialAttrs["finalDamageBoost"].(float64); ok {
		finalDamageBoost = v
	}
	if v, ok := specialAttrs["finalDamageReduce"].(float64); ok {
		finalDamageReduce = v
	}
	if v, ok := specialAttrs["combatBoost"].(float64); ok {
		combatBoost = v
	}
	if v, ok := specialAttrs["resistanceBoost"].(float64); ok {
		resistanceBoost = v
	}

	// 初始化玩家战斗属性
	playerStats := &CombatStats{
		Health:            health,      // 当前生命值
		MaxHealth:         health,      // 最大生命值 = 生命值
		Damage:            attack,      // 伤害值 = attack
		Defense:           defense,     // 防御力
		Speed:             speed,       // 速度（决定先手）
		CritRate:          critRate,    // 暴击率
		ComboRate:         comboRate,   // 连击率
		CounterRate:       counterRate, // 反击率
		StunRate:          stunRate,    // 眩晕率
		DodgeRate:         dodgeRate,   // 闪避率
		VampireRate:       vampireRate, // 吸血率
		CritResist:        critResist,  // 抗暴击
		ComboResist:       comboResist, // 抗连击
		CounterResist:     counterResist,
		StunResist:        stunResist,
		DodgeResist:       dodgeResist,
		VampireResist:     vampireResist,
		HealBoost:         healBoost,
		CritDamageBoost:   critDamageBoost,
		CritDamageReduce:  critDamageReduce,
		FinalDamageBoost:  finalDamageBoost,
		FinalDamageReduce: finalDamageReduce,
		CombatBoost:       combatBoost,
		ResistanceBoost:   resistanceBoost,
	}

	// 应用已选择的增益效果
	playerStats = s.ApplyBuffEffectsToStats(playerStats)

	// 应用难度修饰符
	playerStats.Health *= modifier.HealthMod
	playerStats.MaxHealth *= modifier.HealthMod
	playerStats.Damage *= modifier.DamageMod

	// 初始化敌人战斗属性（基于层数递增）
	enemyStats := &CombatStats{
		// 基础属性：随层数增长
		Health:    playerStats.Health * (1 + float64(floor)*0.05),                       // 敌人血量 = 玩家血量 * (1 + 层数*5%)
		MaxHealth: playerStats.Health * (1 + float64(floor)*0.05),                       // 敌人最大血量
		Damage:    playerStats.Damage * (1 + float64(floor)*0.05),                       // 敌人伤害 = 玩家伤害 * (1 + 层数*5%)
		Defense:   playerStats.Defense * modifier.DamageMod * (1 + float64(floor)*0.03), // 敌人防御 = 玩家防御 * 难度倍数 * (1 + 层数*3%)
		Speed:     playerStats.Speed * (1 + float64(floor)*0.02),                        // 敌人速度 = 玩家速度 * (1 + 层数*2%)
		// 战斗属性：随层数递增
		CritRate:    0.05 + float64(floor)*0.02, // 敌人暴击率 = 5% + 层数*2%
		ComboRate:   0.03 + float64(floor)*0.02, // 敌人连击率 = 3% + 层数*2%
		CounterRate: 0.03 + float64(floor)*0.02, // 敌人反击率 = 3% + 层数*2%
		StunRate:    0.02 + float64(floor)*0.01, // 敌人眩晕率 = 2% + 层数*1%
		DodgeRate:   0.05 + float64(floor)*0.02, // 敌人闪避率 = 5% + 层数*2%
		VampireRate: 0.02 + float64(floor)*0.01, // 敌人吸血率 = 2% + 层数*1%
		// 战斗抗性：随层数递增
		CritResist:    0.02 + float64(floor)*0.01, // 敌人抗暴击 = 2% + 层数*1%
		ComboResist:   0.02 + float64(floor)*0.01, // 敌人抗连击 = 2% + 层数*1%
		CounterResist: 0.02 + float64(floor)*0.01, // 敌人抗反击 = 2% + 层数*1%
		StunResist:    0.02 + float64(floor)*0.01, // 敌人抗眩晕 = 2% + 层数*1%
		DodgeResist:   0.02 + float64(floor)*0.01, // 敌人抗闪避 = 2% + 层数*1%
		VampireResist: 0.02 + float64(floor)*0.01, // 敌人抗吸血 = 2% + 层数*1%
		// 特殊属性：随层数递增
		HealBoost:         0.05 + float64(floor)*0.02, // 敌人治疗强化 = 5% + 层数*2%
		CritDamageBoost:   0.2 + float64(floor)*0.1,   // 敌人爆伤提升 = 20% + 层数*10%
		CritDamageReduce:  0.1 + float64(floor)*0.05,  // 敌人爆伤减免 = 10% + 层数*5%
		FinalDamageBoost:  0.05 + float64(floor)*0.02, // 敌人最终增伤 = 5% + 层数*2%
		FinalDamageReduce: 0.05 + float64(floor)*0.02, // 敌人最终减伤 = 5% + 层数*2%
		CombatBoost:       0.03 + float64(floor)*0.02, // 敌人战斗属性提升 = 3% + 层数*2%
		ResistanceBoost:   0.03 + float64(floor)*0.02, // 敌人抗性提升 = 3% + 层数*2%
	}

	return playerStats, enemyStats, nil
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

	for _, poolConfigs := range BuffConfigPool {
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
}}

// EndDungeon 结束秘境
// 次季：引入了rewardAmount字段，计算宁亭是她StartFight中计算的，这里翰撩直接返回
func (s *DungeonService) EndDungeon(floor int, victory bool) (map[string]interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 简接使用StartFight中计算的rewardAmount
	// 次程：EndDungeon仅达辂引入深深新邮存并欧说
	if victory && s.rewardAmount > 0 {
		user.SpiritStones += s.rewardAmount
	}

	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, fmt.Errorf("更新用户数据失败: %w", err)
	}

	return map[string]interface{}{
		"floor":        floor,
		"totalReward":  s.rewardAmount,
		"victory":      victory,
		"spiritStones": user.SpiritStones,
	}, nil
}

// SaveSessionToRedis 将秘境会话保存到Redis
func (s *DungeonService) SaveSessionToRedis(floor int, difficulty string) error {
	session := &DungeonSession{
		UserID:        s.userID,
		Floor:         floor,
		Difficulty:    difficulty,
		RefreshCount:  3,
		SelectedBuffs: s.selectedBuffs,
		PlayerBuffs:   s.buffEffects,
	}

	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// 清除成秘境会话，有效期60分钟
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// LoadSessionFromRedis 从 Redis 加载秘境会话
func (s *DungeonService) LoadSessionFromRedis() (*DungeonSession, error) {
	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err // 会话不存在
	}

	var session DungeonSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	// 恢复中断不丢失的增益效果
	s.selectedBuffs = session.SelectedBuffs
	s.buffEffects = session.PlayerBuffs

	return &session, nil
}

// ClearSessionFromRedis 清除Redis中的秘境会话
func (s *DungeonService) ClearSessionFromRedis() error {
	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	return redis.Client.Del(redis.Ctx, key).Err()
}
