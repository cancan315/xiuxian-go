package duel

import (
	"encoding/json"
	"fmt"
	"log"
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

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// PvPBattleService 斗法战斗服务
type PvPBattleService struct {
	playerID      int64
	opponentID    int64
	rewardService *RewardService
}

// NewPvPBattleService 创建斗法战斗服务
func NewPvPBattleService(playerID, opponentID int64) *PvPBattleService {
	return &PvPBattleService{
		playerID:      playerID,
		opponentID:    opponentID,
		rewardService: NewRewardService(DefaultRewardConfig()),
	}
}

// PvPBattleStatus 斗法战斗状态
type PvPBattleStatus struct {
	PlayerID          int64            `json:"player_id"`
	OpponentID        int64            `json:"opponent_id"`
	PlayerName        string           `json:"player_name"`
	OpponentName      string           `json:"opponent_name"`
	Round             int              `json:"round"`
	PlayerHealth      float64          `json:"player_health"`
	PlayerMaxHealth   float64          `json:"player_max_health"`
	OpponentHealth    float64          `json:"opponent_health"`
	OpponentMaxHealth float64          `json:"opponent_max_health"`
	PlayerStats       *DuelCombatStats `json:"player_stats"`
	OpponentStats     *DuelCombatStats `json:"opponent_stats"`
	BattleLog         []string         `json:"battle_log"`
}

// DuelCombatStats 斗法战斗属性（从gin.H映射而来）
type DuelCombatStats struct {
	Health        float64 `json:"health"`
	Attack        float64 `json:"attack"`
	Defense       float64 `json:"defense"`
	Speed         float64 `json:"speed"`
	CritRate      float64 `json:"crit_rate"`
	ComboRate     float64 `json:"combo_rate"`
	CounterRate   float64 `json:"counter_rate"`
	StunRate      float64 `json:"stun_rate"`
	DodgeRate     float64 `json:"dodge_rate"`
	VampireRate   float64 `json:"vampire_rate"`
	CritResist    float64 `json:"crit_resist"`
	ComboResist   float64 `json:"combo_resist"`
	CounterResist float64 `json:"counter_resist"`
	StunResist    float64 `json:"stun_resist"`
	DodgeResist   float64 `json:"dodge_resist"`
	VampireResist float64 `json:"vampire_resist"`
}

// PvPRoundData 斗法单回合数据
type PvPRoundData struct {
	Round          int           `json:"round"`
	PlayerHealth   float64       `json:"player_health"`
	OpponentHealth float64       `json:"opponent_health"`
	Logs           []string      `json:"logs"`
	BattleEnded    bool          `json:"battle_ended"`
	Victory        bool          `json:"victory"`
	Rewards        []interface{} `json:"rewards,omitempty"`
}

// convertGinHToStats 将gin.H转换为DuelCombatStats
func convertGinHToStats(data interface{}) *DuelCombatStats {
	stats := &DuelCombatStats{}

	if ginH, ok := data.(map[string]interface{}); ok {
		if v, ok := ginH["baseAttributes"].(map[string]interface{}); ok {
			if val, ok := v["health"].(float64); ok {
				stats.Health = val
			}
			if val, ok := v["attack"].(float64); ok {
				stats.Attack = val
			}
			if val, ok := v["defense"].(float64); ok {
				stats.Defense = val
			}
			if val, ok := v["speed"].(float64); ok {
				stats.Speed = val
			}
		}

		if v, ok := ginH["combatAttributes"].(map[string]interface{}); ok {
			if val, ok := v["critRate"].(float64); ok {
				stats.CritRate = val
			}
			if val, ok := v["comboRate"].(float64); ok {
				stats.ComboRate = val
			}
			if val, ok := v["counterRate"].(float64); ok {
				stats.CounterRate = val
			}
			if val, ok := v["stunRate"].(float64); ok {
				stats.StunRate = val
			}
			if val, ok := v["dodgeRate"].(float64); ok {
				stats.DodgeRate = val
			}
			if val, ok := v["vampireRate"].(float64); ok {
				stats.VampireRate = val
			}
		}

		if v, ok := ginH["combatResistance"].(map[string]interface{}); ok {
			if val, ok := v["critResist"].(float64); ok {
				stats.CritResist = val
			}
			if val, ok := v["comboResist"].(float64); ok {
				stats.ComboResist = val
			}
			if val, ok := v["counterResist"].(float64); ok {
				stats.CounterResist = val
			}
			if val, ok := v["stunResist"].(float64); ok {
				stats.StunResist = val
			}
			if val, ok := v["dodgeResist"].(float64); ok {
				stats.DodgeResist = val
			}
			if val, ok := v["vampireResist"].(float64); ok {
				stats.VampireResist = val
			}
		}
	}

	return stats
}

// convertDuelStatsToBattleStats 将DuelCombatStats转换为battle.CombatStats
func convertDuelStatsToBattleStats(stats *DuelCombatStats) *battle.CombatStats {
	return &battle.CombatStats{
		Health:            stats.Health,
		MaxHealth:         stats.Health,
		Damage:            stats.Attack,
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
		HealBoost:         0.05, // 默认值
		CritDamageBoost:   0.2,  // 默认值
		CritDamageReduce:  0.1,  // 默认值
		FinalDamageBoost:  0.05, // 默认值
		FinalDamageReduce: 0.05, // 默认值
		CombatBoost:       0.03, // 默认值
		ResistanceBoost:   0.03, // 默认值
	}
}

// StartPvPBattle 开始斗法战斗
func (s *PvPBattleService) StartPvPBattle(playerData, opponentData interface{}) (*PvPRoundData, error) {
	// 获取玩家信息
	var player models.User
	if err := db.DB.First(&player, s.playerID).Error; err != nil {
		return nil, fmt.Errorf("玩家不存在: %w", err)
	}

	// 获取对手信息
	var opponent models.User
	if err := db.DB.First(&opponent, s.opponentID).Error; err != nil {
		return nil, fmt.Errorf("对手不存在: %w", err)
	}

	// 转换玩家属性
	playerStats := convertGinHToStats(playerData)
	opponentStats := convertGinHToStats(opponentData)

	// 创建战斗状态并保存到Redis
	battleStatus := &PvPBattleStatus{
		PlayerID:          s.playerID,
		OpponentID:        s.opponentID,
		PlayerName:        player.PlayerName,
		OpponentName:      opponent.PlayerName,
		Round:             0,
		PlayerHealth:      playerStats.Health,
		PlayerMaxHealth:   playerStats.Health,
		OpponentHealth:    opponentStats.Health,
		OpponentMaxHealth: opponentStats.Health,
		PlayerStats:       playerStats,
		OpponentStats:     opponentStats,
		BattleLog:         []string{},
	}

	if err := s.SaveBattleStatusToRedis(battleStatus); err != nil {
		fmt.Printf("保存战斗状态到Redis失败: %v\n", err)
	}

	return &PvPRoundData{
		Round:          0,
		PlayerHealth:   playerStats.Health,
		OpponentHealth: opponentStats.Health,
		Logs:           []string{"战斗已初始化，准备开始！"},
		BattleEnded:    false,
	}, nil
}

// ExecutePvPRound 执行斗法战斗回合
func (s *PvPBattleService) ExecutePvPRound() (*PvPRoundData, error) {
	// 从Redis加载战斗状态
	status, err := s.LoadBattleStatusFromRedis()
	if err != nil {
		return nil, fmt.Errorf("加载战斗状态失败: %w", err)
	}

	if status == nil {
		return nil, fmt.Errorf("战斗状态不存在")
	}

	// 检查战斗是否已结束
	if status.PlayerHealth <= 0 || status.OpponentHealth <= 0 {
		return nil, fmt.Errorf("战斗已结束")
	}

	// 回合间隔：1秒延迟，增强战斗的实时感
	// 检查上一回合的执行时间
	lastRoundKey := fmt.Sprintf("pvp:battle:lastround:%d:%d", s.playerID, s.opponentID)
	lastRoundTime, err := redis.Client.Get(redis.Ctx, lastRoundKey).Int64()
	if err == nil && lastRoundTime > 0 {
		elapsed := time.Now().UnixMilli() - lastRoundTime
		const roundInterval = 1000 // 1秒（毫秒）
		if elapsed < roundInterval {
			waitTime := time.Duration(roundInterval-elapsed) * time.Millisecond
			fmt.Printf("[PvP] 回合%d等待中... 剩余等待时间: %v\n", status.Round+1, waitTime)
			time.Sleep(waitTime)
		}
	}

	// 记录本回合的执行时间
	redis.Client.Set(redis.Ctx, lastRoundKey, time.Now().UnixMilli(), 60*time.Minute)

	// 执行一回合
	status.Round++
	const maxRounds = 100
	if status.Round > maxRounds {
		status.PlayerHealth = 0
	}

	// 根据速度决定先手和后手
	roundLogs := []string{}
	playerSpeed := status.PlayerStats.Speed
	opponentSpeed := status.OpponentStats.Speed

	if playerSpeed >= opponentSpeed {
		// ==================== 玩家先手 ====================
		playerDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.PlayerStats), convertDuelStatsToBattleStats(status.OpponentStats))
		opponentTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.OpponentStats), status.OpponentHealth, playerDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.PlayerStats))
		status.OpponentHealth = opponentTakeDmgResult.CurrentHealth

		// 吸血回复
		if playerDmgResult.IsVampire {
			status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
			if status.PlayerHealth > status.PlayerMaxHealth {
				status.PlayerHealth = status.PlayerMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.PlayerName, status.OpponentName, playerDmgResult.BaseDamage)
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
			logMsg += "，对手被眩晕一回合"
		}
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 检查对手是否死亡
		if opponentTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.OpponentName, status.PlayerName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

			// 获取玩家信息以获取等级
			var player models.User
			if err := db.DB.First(&player, s.playerID).Error; err != nil {
				log.Printf("[Duel] 获取玩家等级失败: %v", err)
			}

			// 计算并发放奖励
			baseRewards := s.rewardService.CalculateRewards(status, player.Level)
			finalRewards := s.rewardService.ApplyRewardMultiplier(baseRewards)

			if err := s.rewardService.GrantRewardsToPlayer(s.playerID, finalRewards); err != nil {
				log.Printf("[Duel] 发放奖励失败: %v", err)
			}

			// 清除回合时间标记
			redis.Client.Del(redis.Ctx, lastRoundKey)

			return &PvPRoundData{
				Round:          status.Round,
				PlayerHealth:   status.PlayerHealth,
				OpponentHealth: math.Max(0, status.OpponentHealth),
				Logs:           roundLogs,
				BattleEnded:    true,
				Victory:        true,
				Rewards: []interface{}{
					map[string]interface{}{
						"type":   "spirit_stone",
						"amount": finalRewards.SpiritStones,
					},
					map[string]interface{}{
						"type":   "cultivation",
						"amount": finalRewards.Cultivation,
					},
				},
			}, nil
		}

		// 对手回合（如果没被眩晕）
		if !playerDmgResult.IsStun {
			opponentDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.OpponentStats), convertDuelStatsToBattleStats(status.PlayerStats))
			playerTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.PlayerStats), status.PlayerHealth, opponentDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.OpponentStats))
			status.PlayerHealth = playerTakeDmgResult.CurrentHealth

			// 对手吸血回复
			if opponentDmgResult.IsVampire {
				status.OpponentHealth = status.OpponentHealth + opponentDmgResult.VampireHeal
				if status.OpponentHealth > status.OpponentMaxHealth {
					status.OpponentHealth = status.OpponentMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.OpponentName, status.PlayerName, opponentDmgResult.BaseDamage)
			if opponentDmgResult.IsCrit {
				logMsg += fmt.Sprintf("，暴击伤害%.0f", opponentDmgResult.CritDamage)
			}
			if opponentDmgResult.IsCombo {
				logMsg += fmt.Sprintf("，连击伤害%.0f", opponentDmgResult.ComboDamage)
			}
			if opponentDmgResult.IsVampire {
				logMsg += fmt.Sprintf("，吸血回复%.0f", opponentDmgResult.VampireHeal)
			}
			if opponentDmgResult.IsStun {
				logMsg += "，玩家被眩晕一回合"
			}
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 检查玩家是否死亡
			if playerTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.PlayerName, status.OpponentName))
				status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

				// 清除回合时间标记
				redis.Client.Del(redis.Ctx, lastRoundKey)

				return &PvPRoundData{
					Round:          status.Round,
					PlayerHealth:   math.Max(0, status.PlayerHealth),
					OpponentHealth: status.OpponentHealth,
					Logs:           roundLogs,
					BattleEnded:    true,
					Victory:        false,
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, fmt.Sprintf("%s被眩晕，无法行动！", status.OpponentName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])
		}
	} else {
		// ==================== 对手先手 ====================
		opponentDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.OpponentStats), convertDuelStatsToBattleStats(status.PlayerStats))
		playerTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.PlayerStats), status.PlayerHealth, opponentDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.OpponentStats))
		status.PlayerHealth = playerTakeDmgResult.CurrentHealth

		// 对手吸血回复
		if opponentDmgResult.IsVampire {
			status.OpponentHealth = status.OpponentHealth + opponentDmgResult.VampireHeal
			if status.OpponentHealth > status.OpponentMaxHealth {
				status.OpponentHealth = status.OpponentMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.OpponentName, status.PlayerName, opponentDmgResult.BaseDamage)
		if opponentDmgResult.IsCrit {
			logMsg += fmt.Sprintf("，暴击伤害%.0f", opponentDmgResult.CritDamage)
		}
		if opponentDmgResult.IsCombo {
			logMsg += fmt.Sprintf("，连击伤害%.0f", opponentDmgResult.ComboDamage)
		}
		if opponentDmgResult.IsVampire {
			logMsg += fmt.Sprintf("，吸血回复%.0f", opponentDmgResult.VampireHeal)
		}
		if opponentDmgResult.IsStun {
			logMsg += "，玩家被眩晕一回合"
		}
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 检查玩家是否死亡
		if playerTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.PlayerName, status.OpponentName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

			// 清除回合时间标记
			redis.Client.Del(redis.Ctx, lastRoundKey)

			return &PvPRoundData{
				Round:          status.Round,
				PlayerHealth:   math.Max(0, status.PlayerHealth),
				OpponentHealth: status.OpponentHealth,
				Logs:           roundLogs,
				BattleEnded:    true,
				Victory:        false,
			}, nil
		}

		// 玩家回合（如果没被眩晕）
		if !opponentDmgResult.IsStun {
			playerDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.PlayerStats), convertDuelStatsToBattleStats(status.OpponentStats))
			opponentTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.OpponentStats), status.OpponentHealth, playerDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.PlayerStats))
			status.OpponentHealth = opponentTakeDmgResult.CurrentHealth

			// 玩家吸血回复
			if playerDmgResult.IsVampire {
				status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
				if status.PlayerHealth > status.PlayerMaxHealth {
					status.PlayerHealth = status.PlayerMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.PlayerName, status.OpponentName, playerDmgResult.BaseDamage)
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
				logMsg += "，对手被眩晕一回合"
			}
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 检查对手是否死亡
			if opponentTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.OpponentName, status.PlayerName))
				status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

				// 计算奖励
				var winnerPlayer models.User
				if err := db.DB.First(&winnerPlayer, s.playerID).Error; err != nil {
					log.Printf("[Duel] 获取玩家信息失败: %v", err)
				}

				// 使用RewardService计算并发放奖励
				baseRewards := s.rewardService.CalculateRewards(status, winnerPlayer.Level)
				finalRewards := s.rewardService.ApplyRewardMultiplier(baseRewards)

				if err := s.rewardService.GrantRewardsToPlayer(s.playerID, finalRewards); err != nil {
					log.Printf("[Duel] 发放奖励失败: %v", err)
				}

				// 清除回合时间标记
				redis.Client.Del(redis.Ctx, lastRoundKey)

				return &PvPRoundData{
					Round:          status.Round,
					PlayerHealth:   status.PlayerHealth,
					OpponentHealth: math.Max(0, status.OpponentHealth),
					Logs:           roundLogs,
					BattleEnded:    true,
					Victory:        true,
					Rewards: []interface{}{
						map[string]interface{}{
							"type":   "spirit_stone",
							"amount": finalRewards.SpiritStones,
						},
						map[string]interface{}{
							"type":   "cultivation",
							"amount": finalRewards.Cultivation,
						},
					},
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, fmt.Sprintf("%s被眩晕，无法行动！", status.PlayerName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])
		}
	}

	// 检查是否超出回合数限制
	if status.Round >= 100 && status.PlayerHealth > 0 && status.OpponentHealth > 0 {
		roundLogs = append(roundLogs, "战斗超出最大回合数，判定为失败！")
		status.BattleLog = append(status.BattleLog, "战斗超出最大回合数，判定为失败！")
		status.PlayerHealth = 0

		// 清除回合时间标记
		redis.Client.Del(redis.Ctx, lastRoundKey)

		return &PvPRoundData{
			Round:          status.Round,
			PlayerHealth:   math.Max(0, status.PlayerHealth),
			OpponentHealth: status.OpponentHealth,
			Logs:           roundLogs,
			BattleEnded:    true,
			Victory:        false,
		}, nil
	}

	// 保存更新的战斗状态到Redis
	if err := s.SaveBattleStatusToRedis(status); err != nil {
		fmt.Printf("保存战斗状态到Redis失败: %v\n", err)
	}

	// 返回本回合的结果（战斗继续）
	return &PvPRoundData{
		Round:          status.Round,
		PlayerHealth:   status.PlayerHealth,
		OpponentHealth: status.OpponentHealth,
		Logs:           roundLogs,
		BattleEnded:    false,
	}, nil
}

// SaveBattleStatusToRedis 将战斗状态保存到Redis
func (s *PvPBattleService) SaveBattleStatusToRedis(status *PvPBattleStatus) error {
	key := fmt.Sprintf("pvp:battle:status:%d:%d", s.playerID, s.opponentID)
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

// LoadBattleStatusFromRedis 从Redis加载战斗状态
func (s *PvPBattleService) LoadBattleStatusFromRedis() (*PvPBattleStatus, error) {
	key := fmt.Sprintf("pvp:battle:status:%d:%d", s.playerID, s.opponentID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var status PvPBattleStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ClearBattleStatusFromRedis 清除Redis中的战斗状态
func (s *PvPBattleService) ClearBattleStatusFromRedis() error {
	key := fmt.Sprintf("pvp:battle:status:%d:%d", s.playerID, s.opponentID)
	return redis.Client.Del(redis.Ctx, key).Err()
}
