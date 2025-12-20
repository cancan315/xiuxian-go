package dungeon

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/dungeon/battle"
	"xiuxian/server-go/internal/dungeon/battle/formula"
	"xiuxian/server-go/internal/dungeon/battle/resolver"
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

// toCombatStats 将 dungeon.CombatStats 转换为 battle.CombatStats
func toCombatStats(stats *CombatStats) *battle.CombatStats {
	return &battle.CombatStats{
		Health:            stats.Health,
		MaxHealth:         stats.MaxHealth,
		Damage:            stats.Damage,
		Defense:           stats.Defense,
		Speed:             stats.Speed,
		CritRate:          stats.CritRate,
		ComboRate:         stats.ComboRate,
		CounterRate:       stats.CounterRate,
		StunRate:          stats.StunRate,
		DodgeRate:         stats.DodgeRate,
		VampireRate:       stats.VampireRate,
		CritResist:        stats.CritResist,
		ComboResist:       stats.ComboResist,
		CounterResist:     stats.CounterResist,
		StunResist:        stats.StunResist,
		DodgeResist:       stats.DodgeResist,
		VampireResist:     stats.VampireResist,
		HealBoost:         stats.HealBoost,
		CritDamageBoost:   stats.CritDamageBoost,
		CritDamageReduce:  stats.CritDamageReduce,
		FinalDamageBoost:  stats.FinalDamageBoost,
		FinalDamageReduce: stats.FinalDamageReduce,
		CombatBoost:       stats.CombatBoost,
		ResistanceBoost:   stats.ResistanceBoost,
	}
}

// fromCombatStats 将 battle.CombatStats 转换为 dungeon.CombatStats
func fromCombatStats(stats *battle.CombatStats) *CombatStats {
	return &CombatStats{
		Health:            stats.Health,
		MaxHealth:         stats.MaxHealth,
		Damage:            stats.Damage,
		Defense:           stats.Defense,
		Speed:             stats.Speed,
		CritRate:          stats.CritRate,
		ComboRate:         stats.ComboRate,
		CounterRate:       stats.CounterRate,
		StunRate:          stats.StunRate,
		DodgeRate:         stats.DodgeRate,
		VampireRate:       stats.VampireRate,
		CritResist:        stats.CritResist,
		ComboResist:       stats.ComboResist,
		CounterResist:     stats.CounterResist,
		StunResist:        stats.StunResist,
		DodgeResist:       stats.DodgeResist,
		VampireResist:     stats.VampireResist,
		HealBoost:         stats.HealBoost,
		CritDamageBoost:   stats.CritDamageBoost,
		CritDamageReduce:  stats.CritDamageReduce,
		FinalDamageBoost:  stats.FinalDamageBoost,
		FinalDamageReduce: stats.FinalDamageReduce,
		CombatBoost:       stats.CombatBoost,
		ResistanceBoost:   stats.ResistanceBoost,
	}
}

// GetRandomBuffs 获取随机增益选项
func (s *DungeonService) GetRandomBuffs(floor int) ([]BuffOption, error) {
	// 计算概率 (common 通过 else 分支隐式处理)
	rareChance := 0.25
	epicChance := 0.05

	if floor%10 == 0 {
		rareChance = 0.3
		epicChance = 0.2
	} else if floor%5 == 0 {
		rareChance = 0.35
		epicChance = 0.15
	}

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

// parsePlayerAttributes 从数据库用户数据中解析玩家属性
func (s *DungeonService) parsePlayerAttributes(user *models.User) (*CombatStats, error) {
	// 解析基础属性JSON
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &baseAttrs); err != nil {
		return nil, fmt.Errorf("解析基础属性失败: %w", err)
	}

	// 解析战斗属性JSON
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
	var health, attack, defense, speed float64
	var critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate float64
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
	return &CombatStats{
		Health:            health,
		MaxHealth:         health,
		Damage:            attack,
		Defense:           defense,
		Speed:             speed,
		CritRate:          critRate,
		ComboRate:         comboRate,
		CounterRate:       counterRate,
		StunRate:          stunRate,
		DodgeRate:         dodgeRate,
		VampireRate:       vampireRate,
		CritResist:        critResist,
		ComboResist:       comboResist,
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
	}, nil
}

// initEnemyStats 根据玩家属性和层数初始化敌人属性
func (s *DungeonService) initEnemyStats(playerStats *CombatStats, floor int, modifier *DifficultyModifier) *CombatStats {
	return &CombatStats{
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

	// 解析玩家属性
	playerStats, err := s.parsePlayerAttributes(&user)
	if err != nil {
		return nil, err
	}

	// 应用已选择的增益效果
	playerStats = fromCombatStats(resolver.ApplyBuffEffectsToStats(toCombatStats(playerStats), s.buffEffects))

	// 应用难度修饰符
	playerStats.Health *= modifier.HealthMod
	playerStats.MaxHealth *= modifier.HealthMod
	playerStats.Damage *= modifier.DamageMod

	// 初始化敌人属性
	enemyStats := s.initEnemyStats(playerStats, floor, &modifier)

	// 将玩家和敌人属性存入Redis
	if err := s.saveBattleStatsToRedis(playerStats, enemyStats, floor); err != nil {
		// Redis存储失败不影响战斗流程，仅记录日志
		fmt.Printf("保存战斗属性到Redis失败: %v\n", err)
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
			playerDmgResult := formula.CalculateDamage(toCombatStats(playerStats), toCombatStats(enemyStats))
			enemyTakeDmgResult := resolver.TakeDamage(toCombatStats(enemyStats), enemyCurrentHealth, playerDmgResult.TotalDamage, toCombatStats(playerStats))
			enemyCurrentHealth = enemyTakeDmgResult.CurrentHealth

			// 吸血回复
			if playerDmgResult.IsVampire {
				playerCurrentHealth = playerCurrentHealth + playerDmgResult.VampireHeal
				if playerCurrentHealth > playerStats.MaxHealth {
					playerCurrentHealth = playerStats.MaxHealth
				}
			}

			// 记录日志："伤害xxx，暴击伤害xxx，连击伤害xxx"
			logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", rounds, playerDmgResult.BaseDamage)
			if playerDmgResult.IsCrit {
				logMsg += fmt.Sprintf("，暴击伤害%.0f", playerDmgResult.CritDamage)
			}
			if playerDmgResult.IsCombo {
				logMsg += fmt.Sprintf("，连击伤害%.0f", playerDmgResult.ComboDamage)
			}
			if playerDmgResult.IsVampire {
				logMsg += fmt.Sprintf("，吸血回复%.0f", playerDmgResult.VampireHeal)
			}
			if playerDmgResult.IsStun {
				logMsg += "，敌人被眩晕一回合"
			}
			battleLog = append(battleLog, logMsg)

			// 第2步：检查敌人是否死亡
			if enemyTakeDmgResult.IsDead {
				battleLog = append(battleLog, "敌人已被击败！")
				break
			}

			// 第3步：敌人回合（如果没被眩晕）
			if !playerDmgResult.IsStun {
				enemyDmgResult := formula.CalculateDamage(toCombatStats(enemyStats), toCombatStats(playerStats))
				playerTakeDmgResult := resolver.TakeDamage(toCombatStats(playerStats), playerCurrentHealth, enemyDmgResult.TotalDamage, toCombatStats(enemyStats))
				playerCurrentHealth = playerTakeDmgResult.CurrentHealth

				// 敌人吸血回复
				if enemyDmgResult.IsVampire {
					enemyCurrentHealth = enemyCurrentHealth + enemyDmgResult.VampireHeal
					if enemyCurrentHealth > enemyStats.MaxHealth {
						enemyCurrentHealth = enemyStats.MaxHealth
					}
				}

				// 记录日志
				logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", rounds, enemyDmgResult.BaseDamage)
				if enemyDmgResult.IsCrit {
					logMsg += fmt.Sprintf("，暴击伤害%.0f", enemyDmgResult.CritDamage)
				}
				if enemyDmgResult.IsCombo {
					logMsg += fmt.Sprintf("，连击伤害%.0f", enemyDmgResult.ComboDamage)
				}
				if enemyDmgResult.IsVampire {
					logMsg += fmt.Sprintf("，敌人吸血回复%.0f", enemyDmgResult.VampireHeal)
				}
				if enemyDmgResult.IsStun {
					logMsg += "，玩家被眩晕一回合"
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
			enemyDmgResult := formula.CalculateDamage(toCombatStats(enemyStats), toCombatStats(playerStats))
			playerTakeDmgResult := resolver.TakeDamage(toCombatStats(playerStats), playerCurrentHealth, enemyDmgResult.TotalDamage, toCombatStats(enemyStats))
			playerCurrentHealth = playerTakeDmgResult.CurrentHealth

			// 敌人吸血回复
			if enemyDmgResult.IsVampire {
				enemyCurrentHealth = enemyCurrentHealth + enemyDmgResult.VampireHeal
				if enemyCurrentHealth > enemyStats.MaxHealth {
					enemyCurrentHealth = enemyStats.MaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", rounds, enemyDmgResult.BaseDamage)
			if enemyDmgResult.IsCrit {
				logMsg += fmt.Sprintf("，暴击伤害%.0f", enemyDmgResult.CritDamage)
			}
			if enemyDmgResult.IsCombo {
				logMsg += fmt.Sprintf("，连击伤害%.0f", enemyDmgResult.ComboDamage)
			}
			if enemyDmgResult.IsVampire {
				logMsg += fmt.Sprintf("，敌人吸血回复%.0f", enemyDmgResult.VampireHeal)
			}
			if enemyDmgResult.IsStun {
				logMsg += "，玩家被眩晕一回合"
			}
			battleLog = append(battleLog, logMsg)

			// 第2步：检查玩家是否死亡
			if playerTakeDmgResult.IsDead {
				battleLog = append(battleLog, "玩家已被击败！")
				break
			}

			// 第3步：玩家回合（如果没被眩晕）
			if !enemyDmgResult.IsStun {
				playerDmgResult := formula.CalculateDamage(toCombatStats(playerStats), toCombatStats(enemyStats))
				enemyTakeDmgResult := resolver.TakeDamage(toCombatStats(enemyStats), enemyCurrentHealth, playerDmgResult.TotalDamage, toCombatStats(playerStats))
				enemyCurrentHealth = enemyTakeDmgResult.CurrentHealth

				// 玩家吸血回复
				if playerDmgResult.IsVampire {
					playerCurrentHealth = playerCurrentHealth + playerDmgResult.VampireHeal
					if playerCurrentHealth > playerStats.MaxHealth {
						playerCurrentHealth = playerStats.MaxHealth
					}
				}

				// 记录日志
				logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", rounds, playerDmgResult.BaseDamage)
				if playerDmgResult.IsCrit {
					logMsg += fmt.Sprintf("，暴击伤害%.0f", playerDmgResult.CritDamage)
				}
				if playerDmgResult.IsCombo {
					logMsg += fmt.Sprintf("，连击伤害%.0f", playerDmgResult.ComboDamage)
				}
				if playerDmgResult.IsVampire {
					logMsg += fmt.Sprintf("，吸血回复%.0f", playerDmgResult.VampireHeal)
				}
				if playerDmgResult.IsStun {
					logMsg += "，敌人被眩晕一回合"
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

	// 解析玩家属性
	playerStats, err := s.parsePlayerAttributes(&user)
	if err != nil {
		return nil, nil, err
	}

	// 应用已选择的增益效果
	playerStats = fromCombatStats(resolver.ApplyBuffEffectsToStats(toCombatStats(playerStats), s.buffEffects))

	// 应用难度修饰符
	playerStats.Health *= modifier.HealthMod
	playerStats.MaxHealth *= modifier.HealthMod
	playerStats.Damage *= modifier.DamageMod

	// 初始化敌人属性
	enemyStats := s.initEnemyStats(playerStats, floor, &modifier)

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
}

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

// saveBattleStatsToRedis 将玩家和敌人属性存入Redis
func (s *DungeonService) saveBattleStatsToRedis(playerStats, enemyStats *CombatStats, floor int) error {
	// 玩家属性
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	playerData, err := json.Marshal(playerStats)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, playerKey, string(playerData), 60*time.Minute).Err(); err != nil {
		return err
	}

	// 敌人属性
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	enemyData, err := json.Marshal(enemyStats)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, enemyKey, string(enemyData), 60*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// LoadBattleStatsFromRedis 从Redis加载玩家和敌人属性
func (s *DungeonService) LoadBattleStatsFromRedis() (*CombatStats, *CombatStats, error) {
	// 加载玩家属性
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	playerData, err := redis.Client.Get(redis.Ctx, playerKey).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("加载玩家属性失败: %w", err)
	}
	var playerStats CombatStats
	if err := json.Unmarshal([]byte(playerData), &playerStats); err != nil {
		return nil, nil, err
	}

	// 加载敌人属性
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	enemyData, err := redis.Client.Get(redis.Ctx, enemyKey).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("加载敌人属性失败: %w", err)
	}
	var enemyStats CombatStats
	if err := json.Unmarshal([]byte(enemyData), &enemyStats); err != nil {
		return nil, nil, err
	}

	return &playerStats, &enemyStats, nil
}

// ClearBattleStatsFromRedis 清除Redis中的战斗属性
func (s *DungeonService) ClearBattleStatsFromRedis() error {
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	return redis.Client.Del(redis.Ctx, playerKey, enemyKey).Err()
}
