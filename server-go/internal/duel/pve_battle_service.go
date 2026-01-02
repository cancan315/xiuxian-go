package duel

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/dungeon/battle/formula"
	"xiuxian/server-go/internal/dungeon/battle/resolver"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
)

// PvEBattleService PvE 妖兽战斗服务
type PvEBattleService struct {
	playerID       int64
	monsterID      int
	difficulty     string // 怪物难度: normal, hard, boss
	rewardService  *RewardService
	monsterFactory *MonsterFactory // 妖兽工厂
}

// PvEBattleStatus PvE 战斗状态
type PvEBattleStatus struct {
	PlayerID         int64            `json:"player_id"`
	MonsterID        int              `json:"monster_id"`
	PlayerName       string           `json:"player_name"`
	MonsterName      string           `json:"monster_name"`
	Round            int              `json:"round"`
	PlayerHealth     float64          `json:"player_health"`
	PlayerMaxHealth  float64          `json:"player_max_health"`
	MonsterHealth    float64          `json:"monster_health"`
	MonsterMaxHealth float64          `json:"monster_max_health"`
	PlayerStats      *DuelCombatStats `json:"player_stats"`
	MonsterStats     *DuelCombatStats `json:"monster_stats"`
	BattleLog        []string         `json:"battle_log"`
}

// MonsterFactory 妖兽数据工厂
type MonsterFactory struct{}

// NewMonsterFactory 创建妖兽工厂
func NewMonsterFactory() *MonsterFactory {
	return &MonsterFactory{}
}

// GetMonsterBattleStats 获取妖兽的战斗属性
// 从妖兽配置中解析 JSON 数据并转换为 DuelCombatStats
func (mf *MonsterFactory) GetMonsterBattleStats(monster interface{}) (*DuelCombatStats, error) {
	stats := &DuelCombatStats{}

	// 尝试作为 map[string]interface{} 类型处理（来自 HTTP 接口）
	if monsterMap, ok := monster.(map[string]interface{}); ok {
		// 解析基础属性
		if baseAttrs, ok := monsterMap["baseAttributes"].(map[string]interface{}); ok {
			if val, ok := baseAttrs["health"].(float64); ok {
				stats.Health = val
			}
			if val, ok := baseAttrs["attack"].(float64); ok {
				stats.Attack = val
			}
			if val, ok := baseAttrs["defense"].(float64); ok {
				stats.Defense = val
			}
			if val, ok := baseAttrs["speed"].(float64); ok {
				stats.Speed = val
			}
		}

		// 解析战斗属性
		if combatAttrs, ok := monsterMap["combatAttributes"].(map[string]interface{}); ok {
			if val, ok := combatAttrs["critRate"].(float64); ok {
				stats.CritRate = val
			}
			if val, ok := combatAttrs["comboRate"].(float64); ok {
				stats.ComboRate = val
			}
			if val, ok := combatAttrs["counterRate"].(float64); ok {
				stats.CounterRate = val
			}
			if val, ok := combatAttrs["stunRate"].(float64); ok {
				stats.StunRate = val
			}
			if val, ok := combatAttrs["dodgeRate"].(float64); ok {
				stats.DodgeRate = val
			}
			if val, ok := combatAttrs["vampireRate"].(float64); ok {
				stats.VampireRate = val
			}
		}

		// 妖兽没有抗性属性，设为 0
		stats.CritResist = 0
		stats.ComboResist = 0
		stats.CounterResist = 0
		stats.StunResist = 0
		stats.DodgeResist = 0
		stats.VampireResist = 0

		return stats, nil
	}

	return nil, fmt.Errorf("无法转换妖兽数据")
}

// NewPvEBattleService 创建PvE战斗服务
// difficulty: 怪物难度 (normal, hard, boss)，由调用方从怪物配置中传入
func NewPvEBattleService(playerID int64, monsterID int, difficulty string) *PvEBattleService {
	// 如果未传入难度，默认为normal
	if difficulty == "" {
		difficulty = "normal"
	}

	return &PvEBattleService{
		playerID:       playerID,
		monsterID:      monsterID,
		difficulty:     difficulty,
		rewardService:  NewRewardService(DefaultRewardConfig()),
		monsterFactory: NewMonsterFactory(),
	}
}

// StartPvEBattle 开始 PvE 战斗
func (s *PvEBattleService) StartPvEBattle(playerData interface{}, monsterData interface{}) (*PvPRoundData, error) {
	// 获取玩家信息
	var player models.User
	if err := db.DB.First(&player, s.playerID).Error; err != nil {
		return nil, fmt.Errorf("玩家不存在: %w", err)
	}

	// 转换玩家属性
	playerStats := convertGinHToStats(playerData)

	// 转换妖兽属性
	monsterStats, err := s.monsterFactory.GetMonsterBattleStats(monsterData)
	if err != nil {
		return nil, fmt.Errorf("妖兽数据转换失败: %w", err)
	}

	// 获取妖兽名称
	monsterName := "未知妖兽"
	if monsterMap, ok := monsterData.(map[string]interface{}); ok {
		if name, ok := monsterMap["name"].(string); ok {
			monsterName = name
		}
	}

	// 创建战斗状态并保存到 Redis
	battleStatus := &PvEBattleStatus{
		PlayerID:         s.playerID,
		MonsterID:        s.monsterID,
		PlayerName:       player.PlayerName,
		MonsterName:      monsterName,
		Round:            0,
		PlayerHealth:     playerStats.Health,
		PlayerMaxHealth:  playerStats.Health,
		MonsterHealth:    monsterStats.Health,
		MonsterMaxHealth: monsterStats.Health,
		PlayerStats:      playerStats,
		MonsterStats:     monsterStats,
		BattleLog:        []string{},
	}

	if err := s.SaveBattleStatusToRedis(battleStatus); err != nil {
		log.Printf("[PvE] 保存战斗状态到 Redis 失败: %v\n", err)
	}

	return &PvPRoundData{
		Round:          0,
		PlayerHealth:   playerStats.Health,
		OpponentHealth: monsterStats.Health,
		Logs:           []string{"战斗已初始化，准备开始！"},
		BattleEnded:    false,
	}, nil
}

// ExecutePvERound 执行 PvE 战斗回合
func (s *PvEBattleService) ExecutePvERound() (*PvPRoundData, error) {
	// 从 Redis 加载战斗状态
	status, err := s.LoadBattleStatusFromRedis()
	if err != nil {
		return nil, fmt.Errorf("加载战斗状态失败: %w", err)
	}

	if status == nil {
		return nil, fmt.Errorf("战斗状态不存在")
	}

	// 检查战斗是否已结束
	if status.PlayerHealth <= 0 || status.MonsterHealth <= 0 {
		return nil, fmt.Errorf("战斗已结束")
	}

	// 回合间隔：1秒延迟
	lastRoundKey := fmt.Sprintf("pve:battle:lastround:%d:%d", s.playerID, s.monsterID)
	lastRoundTime, err := redis.Client.Get(redis.Ctx, lastRoundKey).Int64()
	if err == nil && lastRoundTime > 0 {
		elapsed := time.Now().UnixMilli() - lastRoundTime
		const roundInterval = 1000 // 1秒（毫秒）
		if elapsed < roundInterval {
			waitTime := time.Duration(roundInterval-elapsed) * time.Millisecond
			log.Printf("[PvE] 回合 %d 等待中... 剩余等待时间: %v\n", status.Round+1, waitTime)
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
	monsterSpeed := status.MonsterStats.Speed

	if playerSpeed >= monsterSpeed {
		// ==================== 玩家先手 ====================
		playerDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.PlayerStats), convertDuelStatsToBattleStats(status.MonsterStats))
		monsterTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.MonsterStats), status.MonsterHealth, playerDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.PlayerStats))
		status.MonsterHealth = monsterTakeDmgResult.CurrentHealth

		// 吸血回复
		if playerDmgResult.IsVampire {
			status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
			if status.PlayerHealth > status.PlayerMaxHealth {
				status.PlayerHealth = status.PlayerMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.PlayerName, status.MonsterName, playerDmgResult.BaseDamage)
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
			logMsg += "，妖兽被眩晕一回合"
		}
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 检查妖兽是否死亡
		if monsterTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.MonsterName, status.PlayerName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

			// 检查是普通妖兽还是除魔卫道（通过ID区分：101+为除魔卫道）
			var rewardItems []interface{}
			if s.monsterID >= 101 {
				// 除魔卫道奖励：灵石、修为、丹方残页
				var user models.User
				if err := db.DB.First(&user, s.playerID).Error; err == nil {
					demonRewards := s.rewardService.CalculateRewardsForDemonSlaying(status, user.Level, s.difficulty)
					if demonRewards != nil {
						if err := s.rewardService.GrantDemonSlayingRewardsToPlayer(s.playerID, demonRewards); err != nil {
							log.Printf("[PvE] 发放除魔卫道奖励失败: %v", err)
						}
						// 拆分为多个奖励项，以便前端正确显示
						// 灵石奖励
						rewardItems = append(rewardItems, map[string]interface{}{
							"type":   "spirit_stone",
							"amount": demonRewards.SpiritStones,
						})
						// 修为奖励
						rewardItems = append(rewardItems, map[string]interface{}{
							"type":   "cultivation",
							"amount": demonRewards.Cultivation,
						})
						// 丹方残页奖励（如果有）
						if demonRewards.PillFragmentName != "" {
							rewardItems = append(rewardItems, map[string]interface{}{
								"type":  "pill_fragment",
								"name":  demonRewards.PillFragmentName,
								"count": 1,
							})
						}
					}
				}
			} else {
				// 普通妖兽奖励：仅灵草
				awardRewards := s.rewardService.CalculateRewardsForPvE(status, 0, s.difficulty)
				if awardRewards != nil {
					if err := s.rewardService.GrantPvERewardsToPlayer(s.playerID, awardRewards); err != nil {
						log.Printf("[PvE] 发放奖励失败: %v", err)
					}
					rewardItems = append(rewardItems, map[string]interface{}{
						"type":    "herb",
						"herbId":  awardRewards.HerbID,
						"name":    awardRewards.Name,
						"count":   awardRewards.Count,
						"quality": awardRewards.Quality,
					})
				}
			}

			// 清除回合时间标记
			redis.Client.Del(redis.Ctx, lastRoundKey)

			return &PvPRoundData{
				Round:          status.Round,
				PlayerHealth:   status.PlayerHealth,
				OpponentHealth: math.Max(0, status.MonsterHealth),
				Logs:           roundLogs,
				BattleEnded:    true,
				Victory:        true,
				Rewards:        rewardItems,
			}, nil
		}

		// 妖兽回合（如果没被眩晕）
		if !playerDmgResult.IsStun {
			monsterDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.MonsterStats), convertDuelStatsToBattleStats(status.PlayerStats))
			playerTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.PlayerStats), status.PlayerHealth, monsterDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.MonsterStats))
			status.PlayerHealth = playerTakeDmgResult.CurrentHealth

			// 妖兽吸血回复
			if monsterDmgResult.IsVampire {
				status.MonsterHealth = status.MonsterHealth + monsterDmgResult.VampireHeal
				if status.MonsterHealth > status.MonsterMaxHealth {
					status.MonsterHealth = status.MonsterMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.MonsterName, status.PlayerName, monsterDmgResult.BaseDamage)
			if monsterDmgResult.IsCrit {
				logMsg += fmt.Sprintf("，暴击伤害%.0f", monsterDmgResult.CritDamage)
			}
			if monsterDmgResult.IsCombo {
				logMsg += fmt.Sprintf("，连击伤害%.0f", monsterDmgResult.ComboDamage)
			}
			if monsterDmgResult.IsVampire {
				logMsg += fmt.Sprintf("，吸血回复%.0f", monsterDmgResult.VampireHeal)
			}
			if monsterDmgResult.IsStun {
				logMsg += "，玩家被眩晕一回合"
			}
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 检查玩家是否死亡
			if playerTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s战胜！", status.PlayerName, status.MonsterName))
				status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

				// 清除回合时间标记
				redis.Client.Del(redis.Ctx, lastRoundKey)

				return &PvPRoundData{
					Round:          status.Round,
					PlayerHealth:   math.Max(0, status.PlayerHealth),
					OpponentHealth: status.MonsterHealth,
					Logs:           roundLogs,
					BattleEnded:    true,
					Victory:        false,
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, fmt.Sprintf("%s被眩晕，无法行动！", status.MonsterName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])
		}
	} else {
		// ==================== 妖兽先手 ====================
		monsterDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.MonsterStats), convertDuelStatsToBattleStats(status.PlayerStats))
		playerTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.PlayerStats), status.PlayerHealth, monsterDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.MonsterStats))
		status.PlayerHealth = playerTakeDmgResult.CurrentHealth

		// 妖兽吸血回复
		if monsterDmgResult.IsVampire {
			status.MonsterHealth = status.MonsterHealth + monsterDmgResult.VampireHeal
			if status.MonsterHealth > status.MonsterMaxHealth {
				status.MonsterHealth = status.MonsterMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.MonsterName, status.PlayerName, monsterDmgResult.BaseDamage)
		if monsterDmgResult.IsCrit {
			logMsg += fmt.Sprintf("，暴击伤害%.0f", monsterDmgResult.CritDamage)
		}
		if monsterDmgResult.IsCombo {
			logMsg += fmt.Sprintf("，连击伤害%.0f", monsterDmgResult.ComboDamage)
		}
		if monsterDmgResult.IsVampire {
			logMsg += fmt.Sprintf("，吸血回复%.0f", monsterDmgResult.VampireHeal)
		}
		if monsterDmgResult.IsStun {
			logMsg += "，玩家被眩晕一回合"
		}
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 检查玩家是否死亡
		if playerTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s战胜！", status.PlayerName, status.MonsterName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

			// 清除回合时间标记
			redis.Client.Del(redis.Ctx, lastRoundKey)

			return &PvPRoundData{
				Round:          status.Round,
				PlayerHealth:   math.Max(0, status.PlayerHealth),
				OpponentHealth: status.MonsterHealth,
				Logs:           roundLogs,
				BattleEnded:    true,
				Victory:        false,
			}, nil
		}

		// 玩家回合（如果没被眩晕）
		if !monsterDmgResult.IsStun {
			playerDmgResult := formula.CalculateDamage(convertDuelStatsToBattleStats(status.PlayerStats), convertDuelStatsToBattleStats(status.MonsterStats))
			monsterTakeDmgResult := resolver.TakeDamage(convertDuelStatsToBattleStats(status.MonsterStats), status.MonsterHealth, playerDmgResult.TotalDamage, convertDuelStatsToBattleStats(status.PlayerStats))
			status.MonsterHealth = monsterTakeDmgResult.CurrentHealth

			// 玩家吸血回复
			if playerDmgResult.IsVampire {
				status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
				if status.PlayerHealth > status.PlayerMaxHealth {
					status.PlayerHealth = status.PlayerMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：%s对%s造成伤害%.0f", status.Round, status.PlayerName, status.MonsterName, playerDmgResult.BaseDamage)
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
				logMsg += "，妖兽被眩晕一回合"
			}
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 检查妖兽是否死亡
			if monsterTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, fmt.Sprintf("%s已被击败！%s获得胜利！", status.MonsterName, status.PlayerName))
				status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])

				// 计算并发放奖励
				awardRewards := s.rewardService.CalculateRewardsForPvE(status, 0, s.difficulty) // playerLevel 不用于灵草奖励

				if awardRewards != nil {
					if err := s.rewardService.GrantPvERewardsToPlayer(s.playerID, awardRewards); err != nil {
						log.Printf("[PvE] 发放奖励失败: %v", err)
					}
				}

				// 清除回合时间标记
				redis.Client.Del(redis.Ctx, lastRoundKey)

				var rewardItems []interface{}
				if awardRewards != nil {
					rewardItems = append(rewardItems, map[string]interface{}{
						"type":    "herb",
						"herbId":  awardRewards.HerbID,
						"name":    awardRewards.Name,
						"count":   awardRewards.Count,
						"quality": awardRewards.Quality,
					})
				}

				return &PvPRoundData{
					Round:          status.Round,
					PlayerHealth:   status.PlayerHealth,
					OpponentHealth: math.Max(0, status.MonsterHealth),
					Logs:           roundLogs,
					BattleEnded:    true,
					Victory:        true,
					Rewards:        rewardItems,
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, fmt.Sprintf("%s被眩晕，无法行动！", status.PlayerName))
			status.BattleLog = append(status.BattleLog, roundLogs[len(roundLogs)-1])
		}
	}

	// 检查是否超出回合数限制
	if status.Round >= 100 && status.PlayerHealth > 0 && status.MonsterHealth > 0 {
		roundLogs = append(roundLogs, "战斗超出最大回合数，判定为失败！")
		status.BattleLog = append(status.BattleLog, "战斗超出最大回合数，判定为失败！")
		status.PlayerHealth = 0

		// 清除回合时间标记
		redis.Client.Del(redis.Ctx, lastRoundKey)

		return &PvPRoundData{
			Round:          status.Round,
			PlayerHealth:   math.Max(0, status.PlayerHealth),
			OpponentHealth: status.MonsterHealth,
			Logs:           roundLogs,
			BattleEnded:    true,
			Victory:        false,
		}, nil
	}

	// 保存更新的战斗状态到 Redis
	if err := s.SaveBattleStatusToRedis(status); err != nil {
		log.Printf("[PvE] 保存战斗状态到 Redis 失败: %v\n", err)
	}

	// 返回本回合的结果（战斗继续）
	return &PvPRoundData{
		Round:          status.Round,
		PlayerHealth:   status.PlayerHealth,
		OpponentHealth: status.MonsterHealth,
		Logs:           roundLogs,
		BattleEnded:    false,
	}, nil
}

// SaveBattleStatusToRedis 将战斗状态保存到 Redis
func (s *PvEBattleService) SaveBattleStatusToRedis(status *PvEBattleStatus) error {
	key := fmt.Sprintf("pve:battle:status:%d:%d", s.playerID, s.monsterID)
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

// LoadBattleStatusFromRedis 从 Redis 加载战斗状态
func (s *PvEBattleService) LoadBattleStatusFromRedis() (*PvEBattleStatus, error) {
	key := fmt.Sprintf("pve:battle:status:%d:%d", s.playerID, s.monsterID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var status PvEBattleStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ClearBattleStatusFromRedis 清除 Redis 中的战斗状态
func (s *PvEBattleService) ClearBattleStatusFromRedis() error {
	key := fmt.Sprintf("pve:battle:status:%d:%d", s.playerID, s.monsterID)
	return redis.Client.Del(redis.Ctx, key).Err()
}
