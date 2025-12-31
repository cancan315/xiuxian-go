package duel

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"github.com/gin-gonic/gin"
)

// GetBothPlayersAttributesForBattle 获取斗法双方的战斗属性数据（直接从users表）
// 用于PvP战斗初始化时获取双方完整的属性数据
func GetBothPlayersAttributesForBattle(playerID, opponentID int64) (playerData gin.H, opponentData gin.H, err error) {
	var playerUser, opponentUser models.User

	// 查询玩家信息
	if err := db.DB.First(&playerUser, playerID).Error; err != nil {
		return nil, nil, fmt.Errorf("玩家不存在: %w", err)
	}

	// 查询对手信息
	if err := db.DB.First(&opponentUser, opponentID).Error; err != nil {
		return nil, nil, fmt.Errorf("对手不存在: %w", err)
	}

	// 将玩家数据转换为战斗格式
	playerData = buildBattleAttributes(&playerUser)
	opponentData = buildBattleAttributes(&opponentUser)

	return playerData, opponentData, nil
}

// buildBattleAttributes 将User模型的JSON属性字段转换为gin.H格式的战斗属性
func buildBattleAttributes(user *models.User) gin.H {
	battleData := gin.H{
		"id":             user.ID,
		"playerName":     user.PlayerName,
		"level":          user.Level,
		"realm":          user.Realm,
		"cultivation":    user.Cultivation,
		"maxCultivation": user.MaxCultivation,
	}

	// 解析基础属性
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &baseAttrs); err == nil {
		battleData["baseAttributes"] = baseAttrs
	} else {
		battleData["baseAttributes"] = gin.H{
			"health":  0,
			"attack":  0,
			"defense": 0,
			"speed":   0,
		}
	}

	// 解析战斗属性
	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(user.CombatAttributes, &combatAttrs); err == nil {
		battleData["combatAttributes"] = combatAttrs
	} else {
		battleData["combatAttributes"] = gin.H{
			"critRate":    0,
			"comboRate":   0,
			"counterRate": 0,
			"stunRate":    0,
			"dodgeRate":   0,
			"vampireRate": 0,
		}
	}

	// 解析战斗抗性
	var combatResist map[string]interface{}
	if err := json.Unmarshal(user.CombatResistance, &combatResist); err == nil {
		battleData["combatResistance"] = combatResist
	} else {
		battleData["combatResistance"] = gin.H{
			"critResist":    0,
			"comboResist":   0,
			"counterResist": 0,
			"stunResist":    0,
			"dodgeResist":   0,
			"vampireResist": 0,
		}
	}

	// 解析特殊属性（如果需要）
	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(user.SpecialAttributes, &specialAttrs); err == nil {
		battleData["specialAttributes"] = specialAttrs
	}

	return battleData
}

// GetPlayerBattleData 获取玩家的战斗数据
// 欧旧函数，已改为使用GetBothPlayersAttributesForBattle
func GetPlayerBattleData(playerID int64) (gin.H, error) {
	var playerUser models.User

	// 查询玩家信息
	if err := db.DB.First(&playerUser, playerID).Error; err != nil {
		return nil, fmt.Errorf("玩家不存在: %w", err)
	}

	return buildBattleAttributes(&playerUser), nil
}

// GetDuelRecords 获取玩家的斗法战绵和统计
func GetDuelRecords(playerID int64, offset, limit int) ([]gin.H, gin.H, error) {
	// 查询战斗记录
	query := `
	SELECT 
		id,
		opponent_id as opponentId,
		opponent_name as opponentName,
		result,
		battle_type as battleType,
		rewards,
		created_at as time
	FROM battle_records
	WHERE player_id = ?
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`

	// 使用 GORM 的 Raw 方法执行原生 SQL 查询
	rows, err := db.DB.Raw(query, playerID, limit, offset).Rows()
	if err != nil {
		log.Printf("[Duel] 查询战斗记录失败: %v", err)
		return nil, nil, err
	}
	defer rows.Close()

	var records []gin.H
	for rows.Next() {
		record := gin.H{}
		var id interface{}
		var opponentId interface{}
		var opponentName interface{}
		var result interface{}
		var battleType interface{}
		var rewards interface{}
		var createdAt time.Time

		if err := rows.Scan(
			&id,
			&opponentId,
			&opponentName,
			&result,
			&battleType,
			&rewards,
			&createdAt,
		); err != nil {
			log.Printf("[Duel] 扫描战斗记录失败: %v", err)
			return nil, nil, err
		}

		// 将扫描结果赋值到 record 对象
		record["id"] = id
		record["opponentId"] = opponentId
		record["opponentName"] = opponentName
		record["result"] = result
		record["battleType"] = battleType
		record["rewards"] = rewards
		record["time"] = createdAt.Format("2006-01-02 15:04:05")
		records = append(records, record)
	}

	// 查询统计数据
	statsQuery := `
	SELECT 
		COUNT(*) as totalBattles,
		SUM(CASE WHEN result = '胜利' THEN 1 ELSE 0 END) as wins,
		SUM(CASE WHEN result = '失败' THEN 1 ELSE 0 END) as losses
	FROM battle_records
	WHERE player_id = ?
	`

	var totalBattles, wins, losses int
	// 使用 GORM 的 Raw 方法执行原生 SQL 查询
	err = db.DB.Raw(statsQuery, playerID).Row().Scan(&totalBattles, &wins, &losses)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[Duel] 查询战斗统计失败: %v", err)
		return nil, nil, err
	}

	// 计算胜率
	winRate := 0
	if totalBattles > 0 {
		winRate = (wins * 100) / totalBattles
	}

	stats := gin.H{
		"totalBattles":     totalBattles,
		"wins":             wins,
		"losses":           losses,
		"winRate":          winRate,
		"currentWinStreak": 0, // 可以根据需要计算连胜
		"maxWinStreak":     0, // 可以根据需要计算最高连胜
	}

	return records, stats, nil
}

// RecordBattleResult 记录战斗结果
func RecordBattleResult(battleRecord *models.BattleRecord) error {
	query := `
	INSERT INTO battle_records 
	(player_id, opponent_id, opponent_name, result, battle_type, rewards, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// 使用 GORM 的 Exec 方法执行原生 SQL
	err := db.DB.Exec(
		query,
		battleRecord.PlayerID,
		battleRecord.OpponentID,
		battleRecord.OpponentName,
		battleRecord.Result,
		battleRecord.BattleType,
		battleRecord.Rewards,
		time.Now(),
	).Error

	if err != nil {
		log.Printf("[Duel] 记录战斗结果失败: %v", err)
		return err
	}

	return nil
}

// ClaimBattleRewards 领取战斗奖励
func ClaimBattleRewards(playerID int64, rewards []string) error {
	// 这里可以根据奖励类型更新玩家数据
	// 例如：增加灵石、修为等
	for _, reward := range rewards {
		log.Printf("[Duel] 玩家 %d 获得奖励: %s", playerID, reward)
	}
	return nil
}
