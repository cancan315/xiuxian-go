package dungeon

import (
	"fmt"
	"math"

	"xiuxian/server-go/internal/dungeon/battle/formula"
	"xiuxian/server-go/internal/dungeon/battle/resolver"
)

// ExecuteRound 执行单个回合战斗 - 从Redis加载战斗状态，执行一回合，保存结果
func (s *DungeonService) ExecuteRound() (*RoundData, error) {
	// 从Redis加载战斗状态
	status, err := s.LoadBattleStatusFromRedis()
	if err != nil {
		return nil, fmt.Errorf("加载战斗状态失败: %w", err)
	}

	if status == nil {
		return nil, fmt.Errorf("战斗状态不存在")
	}

	// 检查战斗是否已结束
	if status.PlayerHealth <= 0 || status.EnemyHealth <= 0 {
		return nil, fmt.Errorf("战斗已结束")
	}

	// 执行一回合
	status.Round++
	const maxRounds = 100
	if status.Round > maxRounds {
		// 超出回合数，战斗失败
		status.PlayerHealth = 0
	}

	// 根据速度决定先手和后手
	roundLogs := []string{}
	playerSpeed := status.PlayerStats.Speed * (1 + status.PlayerStats.CombatBoost)
	enemySpeed := status.EnemyStats.Speed * (1 + status.EnemyStats.CombatBoost)

	if playerSpeed >= enemySpeed {
		// ==================== 玩家先手 ====================
		// 第1步：玩家攻击敌人
		playerDmgResult := formula.CalculateDamage(toCombatStats(status.PlayerStats), toCombatStats(status.EnemyStats))
		enemyTakeDmgResult := resolver.TakeDamage(toCombatStats(status.EnemyStats), status.EnemyHealth, playerDmgResult.TotalDamage, toCombatStats(status.PlayerStats))
		status.EnemyHealth = enemyTakeDmgResult.CurrentHealth

		// 吸血回复
		if playerDmgResult.IsVampire {
			status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
			if status.PlayerHealth > status.PlayerMaxHealth {
				status.PlayerHealth = status.PlayerMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", status.Round, playerDmgResult.BaseDamage)
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
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 第2步：检查敌人是否死亡
		if enemyTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, "敌人已被击败！")
			status.BattleLog = append(status.BattleLog, "敌人已被击败！")

			// 计算奖励
			modifier := GetDifficultyModifier(status.Difficulty)
			s.rewardAmount = int(100 * modifier.RewardMod * (1 + float64(status.Floor)*0.1))

			// 返回战斗结束的回合数据
			return &RoundData{
				Round:        status.Round,
				PlayerHealth: status.PlayerHealth,
				EnemyHealth:  math.Max(0, status.EnemyHealth),
				Logs:         roundLogs,
				BattleEnded:  true,
				Victory:      true,
				Rewards: []interface{}{
					map[string]interface{}{
						"type":   "spirit_stone",
						"amount": s.rewardAmount,
					},
				},
			}, nil
		}

		// 第3步：敌人回合（如果没被眩晕）
		if !playerDmgResult.IsStun {
			enemyDmgResult := formula.CalculateDamage(toCombatStats(status.EnemyStats), toCombatStats(status.PlayerStats))
			playerTakeDmgResult := resolver.TakeDamage(toCombatStats(status.PlayerStats), status.PlayerHealth, enemyDmgResult.TotalDamage, toCombatStats(status.EnemyStats))
			status.PlayerHealth = playerTakeDmgResult.CurrentHealth

			// 敌人吸血回复
			if enemyDmgResult.IsVampire {
				status.EnemyHealth = status.EnemyHealth + enemyDmgResult.VampireHeal
				if status.EnemyHealth > status.EnemyMaxHealth {
					status.EnemyHealth = status.EnemyMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", status.Round, enemyDmgResult.BaseDamage)
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
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 第4步：检查玩家是否死亡
			if playerTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, "玩家已被击败！")
				status.BattleLog = append(status.BattleLog, "玩家已被击败！")

				// 返回战斗结束的回合数据
				return &RoundData{
					Round:        status.Round,
					PlayerHealth: math.Max(0, status.PlayerHealth),
					EnemyHealth:  status.EnemyHealth,
					Logs:         roundLogs,
					BattleEnded:  true,
					Victory:      false,
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, "敌人被眩晕，无法行动！")
			status.BattleLog = append(status.BattleLog, "敌人被眩晕，无法行动！")
		}
	} else {
		// ==================== 敌人先手 ====================
		// 第1步：敌人攻击玩家
		enemyDmgResult := formula.CalculateDamage(toCombatStats(status.EnemyStats), toCombatStats(status.PlayerStats))
		playerTakeDmgResult := resolver.TakeDamage(toCombatStats(status.PlayerStats), status.PlayerHealth, enemyDmgResult.TotalDamage, toCombatStats(status.EnemyStats))
		status.PlayerHealth = playerTakeDmgResult.CurrentHealth

		// 敌人吸血回复
		if enemyDmgResult.IsVampire {
			status.EnemyHealth = status.EnemyHealth + enemyDmgResult.VampireHeal
			if status.EnemyHealth > status.EnemyMaxHealth {
				status.EnemyHealth = status.EnemyMaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", status.Round, enemyDmgResult.BaseDamage)
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
		roundLogs = append(roundLogs, logMsg)
		status.BattleLog = append(status.BattleLog, logMsg)

		// 第2步：检查玩家是否死亡
		if playerTakeDmgResult.IsDead {
			roundLogs = append(roundLogs, "玩家已被击败！")
			status.BattleLog = append(status.BattleLog, "玩家已被击败！")

			// 返回战斗结束的回合数据
			return &RoundData{
				Round:        status.Round,
				PlayerHealth: math.Max(0, status.PlayerHealth),
				EnemyHealth:  status.EnemyHealth,
				Logs:         roundLogs,
				BattleEnded:  true,
				Victory:      false,
			}, nil
		}

		// 第3步：玩家回合（如果没被眩晕）
		if !enemyDmgResult.IsStun {
			playerDmgResult := formula.CalculateDamage(toCombatStats(status.PlayerStats), toCombatStats(status.EnemyStats))
			enemyTakeDmgResult := resolver.TakeDamage(toCombatStats(status.EnemyStats), status.EnemyHealth, playerDmgResult.TotalDamage, toCombatStats(status.PlayerStats))
			status.EnemyHealth = enemyTakeDmgResult.CurrentHealth

			// 玩家吸血回复
			if playerDmgResult.IsVampire {
				status.PlayerHealth = status.PlayerHealth + playerDmgResult.VampireHeal
				if status.PlayerHealth > status.PlayerMaxHealth {
					status.PlayerHealth = status.PlayerMaxHealth
				}
			}

			// 记录日志
			logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", status.Round, playerDmgResult.BaseDamage)
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
			roundLogs = append(roundLogs, logMsg)
			status.BattleLog = append(status.BattleLog, logMsg)

			// 第4步：检查敌人是否死亡
			if enemyTakeDmgResult.IsDead {
				roundLogs = append(roundLogs, "敌人已被击败！")
				status.BattleLog = append(status.BattleLog, "敌人已被击败！")

				// 计算奖励
				modifier := GetDifficultyModifier(status.Difficulty)
				s.rewardAmount = int(100 * modifier.RewardMod * (1 + float64(status.Floor)*0.1))

				// 返回战斗结束的回合数据
				return &RoundData{
					Round:        status.Round,
					PlayerHealth: status.PlayerHealth,
					EnemyHealth:  math.Max(0, status.EnemyHealth),
					Logs:         roundLogs,
					BattleEnded:  true,
					Victory:      true,
					Rewards: []interface{}{
						map[string]interface{}{
							"type":   "spirit_stone",
							"amount": s.rewardAmount,
						},
					},
				}, nil
			}
		} else {
			roundLogs = append(roundLogs, "玩家被眩晕，无法行动！")
			status.BattleLog = append(status.BattleLog, "玩家被眩晕，无法行动！")
		}
	}

	// 检查是否超出回合数限制
	if status.Round >= 100 && status.PlayerHealth > 0 && status.EnemyHealth > 0 {
		roundLogs = append(roundLogs, "战斗超出最大回合数，判定为失败！")
		status.BattleLog = append(status.BattleLog, "战斗超出最大回合数，判定为失败！")
		status.PlayerHealth = 0 // 强制失败

		// 返回战斗结束的回合数据
		return &RoundData{
			Round:        status.Round,
			PlayerHealth: math.Max(0, status.PlayerHealth),
			EnemyHealth:  status.EnemyHealth,
			Logs:         roundLogs,
			BattleEnded:  true,
			Victory:      false,
		}, nil
	}

	// 保存更新的战斗状态到Redis
	if err := s.SaveBattleStatusToRedis(status); err != nil {
		// Redis保存失败不影响流程，仅记录日志
		fmt.Printf("保存战斗状态到Redis失败: %v\n", err)
	}

	// 返回本回合的结果（战斗继续）
	return &RoundData{
		Round:        status.Round,
		PlayerHealth: status.PlayerHealth,
		EnemyHealth:  status.EnemyHealth,
		Logs:         roundLogs,
		BattleEnded:  false,
	}, nil
}
