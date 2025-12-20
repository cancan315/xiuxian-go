package engine

import (
	"fmt"

	"xiuxian/server-go/internal/dungeon/battle"
	"xiuxian/server-go/internal/dungeon/battle/formula"
	"xiuxian/server-go/internal/dungeon/battle/resolver"
)

// BattleEngine 战斗引擎
type BattleEngine struct {
	playerStats  *battle.CombatStats
	enemyStats   *battle.CombatStats
	battleLog    []string
	playerHealth float64
	enemyHealth  float64
	round        int
	maxRounds    int
}

// NewBattleEngine 创建战斗引擎
func NewBattleEngine(player, enemy *battle.CombatStats) *BattleEngine {
	return &BattleEngine{
		playerStats:  player,
		enemyStats:   enemy,
		battleLog:    []string{},
		playerHealth: player.MaxHealth,
		enemyHealth:  enemy.MaxHealth,
		round:        0,
		maxRounds:    100,
	}
}

// ExecuteRound 执行单个回合
func (e *BattleEngine) ExecuteRound() *battle.RoundResult {
	e.round++
	result := &battle.RoundResult{
		Round:        e.round,
		PlayerHealth: e.playerHealth,
		EnemyHealth:  e.enemyHealth,
		BattleEnded:  false,
		Victory:      false,
	}

	// 检查战斗是否已结束
	if e.playerHealth <= 0 || e.enemyHealth <= 0 || e.round > e.maxRounds {
		result.BattleEnded = true
		if e.round > e.maxRounds && e.playerHealth > 0 && e.enemyHealth > 0 {
			// 超时失败
			e.playerHealth = 0
		}
		result.Victory = e.playerHealth > 0
		result.PlayerHealth = e.playerHealth
		result.EnemyHealth = e.enemyHealth
		result.Log = e.battleLog
		return result
	}

	// 根据速度决定先手和后手
	playerSpeed := e.playerStats.Speed * (1 + e.playerStats.CombatBoost)
	enemySpeed := e.enemyStats.Speed * (1 + e.enemyStats.CombatBoost)

	if playerSpeed >= enemySpeed {
		e.executePlayerFirstRound()
	} else {
		e.executeEnemyFirstRound()
	}

	// 更新回合结果
	result.Round = e.round
	result.PlayerHealth = e.playerHealth
	result.EnemyHealth = e.enemyHealth
	result.Log = append([]string{}, e.battleLog...)
	result.BattleEnded = e.playerHealth <= 0 || e.enemyHealth <= 0
	if result.BattleEnded {
		result.Victory = e.playerHealth > 0
	}

	return result
}

// executePlayerFirstRound 玩家先手回合
func (e *BattleEngine) executePlayerFirstRound() {
	// 第1步：玩家攻击敌人
	playerDmgResult := formula.CalculateDamage(e.playerStats, e.enemyStats)
	enemyTakeDmgResult := resolver.TakeDamage(e.enemyStats, e.enemyHealth, playerDmgResult.TotalDamage, e.playerStats)
	e.enemyHealth = enemyTakeDmgResult.CurrentHealth

	// 吸血回复
	if playerDmgResult.IsVampire {
		e.playerHealth = e.playerHealth + playerDmgResult.VampireHeal
		if e.playerHealth > e.playerStats.MaxHealth {
			e.playerHealth = e.playerStats.MaxHealth
		}
	}

	// 记录日志："伤害xxx，暴击伤害xxx，连击伤害xxx"
	logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", e.round, playerDmgResult.BaseDamage)
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
	e.battleLog = append(e.battleLog, logMsg)

	// 第2步：检查敌人是否死亡
	if enemyTakeDmgResult.IsDead {
		e.battleLog = append(e.battleLog, "敌人已被击败！")
		return
	}

	// 第3步：敌人回合（如果没被眩晕）
	if !playerDmgResult.IsStun {
		enemyDmgResult := formula.CalculateDamage(e.enemyStats, e.playerStats)
		playerTakeDmgResult := resolver.TakeDamage(e.playerStats, e.playerHealth, enemyDmgResult.TotalDamage, e.enemyStats)
		e.playerHealth = playerTakeDmgResult.CurrentHealth

		// 敌人吸血回复
		if enemyDmgResult.IsVampire {
			e.enemyHealth = e.enemyHealth + enemyDmgResult.VampireHeal
			if e.enemyHealth > e.enemyStats.MaxHealth {
				e.enemyHealth = e.enemyStats.MaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", e.round, enemyDmgResult.BaseDamage)
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
		e.battleLog = append(e.battleLog, logMsg)

		// 第4步：检查玩家是否死亡
		if playerTakeDmgResult.IsDead {
			e.battleLog = append(e.battleLog, "玩家已被击败！")
		}
	} else {
		e.battleLog = append(e.battleLog, "敌人被眩晕，无法行动！")
	}
}

// executeEnemyFirstRound 敌人先手回合
func (e *BattleEngine) executeEnemyFirstRound() {
	// 第1步：敌人攻击玩家
	enemyDmgResult := formula.CalculateDamage(e.enemyStats, e.playerStats)
	playerTakeDmgResult := resolver.TakeDamage(e.playerStats, e.playerHealth, enemyDmgResult.TotalDamage, e.enemyStats)
	e.playerHealth = playerTakeDmgResult.CurrentHealth

	// 敌人吸血回复
	if enemyDmgResult.IsVampire {
		e.enemyHealth = e.enemyHealth + enemyDmgResult.VampireHeal
		if e.enemyHealth > e.enemyStats.MaxHealth {
			e.enemyHealth = e.enemyStats.MaxHealth
		}
	}

	// 记录日志
	logMsg := fmt.Sprintf("第%d回合：敌人对玩家造成伤害%.0f", e.round, enemyDmgResult.BaseDamage)
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
	e.battleLog = append(e.battleLog, logMsg)

	// 第2步：检查玩家是否死亡
	if playerTakeDmgResult.IsDead {
		e.battleLog = append(e.battleLog, "玩家已被击败！")
		return
	}

	// 第3步：玩家回合（如果没被眩晕）
	if !enemyDmgResult.IsStun {
		playerDmgResult := formula.CalculateDamage(e.playerStats, e.enemyStats)
		enemyTakeDmgResult := resolver.TakeDamage(e.enemyStats, e.enemyHealth, playerDmgResult.TotalDamage, e.playerStats)
		e.enemyHealth = enemyTakeDmgResult.CurrentHealth

		// 玩家吸血回复
		if playerDmgResult.IsVampire {
			e.playerHealth = e.playerHealth + playerDmgResult.VampireHeal
			if e.playerHealth > e.playerStats.MaxHealth {
				e.playerHealth = e.playerStats.MaxHealth
			}
		}

		// 记录日志
		logMsg := fmt.Sprintf("第%d回合：玩家对敌人造成伤害%.0f", e.round, playerDmgResult.BaseDamage)
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
		e.battleLog = append(e.battleLog, logMsg)

		// 第4步：检查敌人是否死亡
		if enemyTakeDmgResult.IsDead {
			e.battleLog = append(e.battleLog, "敌人已被击败！")
		}
	} else {
		e.battleLog = append(e.battleLog, "玩家被眩晕，无法行动！")
	}
}

// GetBattleLog 获取战斗日志
func (e *BattleEngine) GetBattleLog() []string {
	return e.battleLog
}

// IsFinished 战斗是否结束
func (e *BattleEngine) IsFinished() bool {
	return e.playerHealth <= 0 || e.enemyHealth <= 0 || e.round >= e.maxRounds
}

// GetFinalResult 获取最终结果
func (e *BattleEngine) GetFinalResult() (victory bool, playerHealth, enemyHealth float64) {
	if e.round >= e.maxRounds && e.playerHealth > 0 && e.enemyHealth > 0 {
		// 超时失败
		playerHealth = 0
	} else {
		playerHealth = e.playerHealth
	}
	enemyHealth = e.enemyHealth
	victory = playerHealth > 0
	return
}
