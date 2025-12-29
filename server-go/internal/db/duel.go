package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"xiuxian/server-go/internal/models"

	"github.com/gin-gonic/gin"
)

// GetDuelOpponents 获取可以挑战的道友列表
// 返回除了当前玩家外的所有玩家列表
func GetDuelOpponents(currentUserID int64, offset, limit int) ([]gin.H, error) {
	query := `
	SELECT 
		u.id,
		u.player_name as playerName,
		u.level,
		u.realm,
		u.cultivation,
		u.max_cultivation as maxCultivation,
		u.spirit_stones as spiritStones,
		u.base_attributes,
		u.combat_attributes,
		u.combat_resistance
	FROM users u
	WHERE u.id != ?
	ORDER BY u.id
	LIMIT ? OFFSET ?
	`

	// 使用 GORM 的 Raw 方法执行原生 SQL 查询
	rows, err := DB.Raw(query, currentUserID, limit, offset).Rows()
	if err != nil {
		log.Printf("[Duel] 查询对手列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var opponents []gin.H

	for rows.Next() {
		opponent := gin.H{}
		var id interface{}
		var playerName interface{}
		var level int
		var realm interface{}
		var cultivation, maxCultivation interface{}
		var spiritStones interface{}
		var baseAttributesJSON, combatAttributesJSON, combatResistanceJSON []byte

		if err := rows.Scan(
			&id,
			&playerName,
			&level,
			&realm,
			&cultivation,
			&maxCultivation,
			&spiritStones,
			&baseAttributesJSON,
			&combatAttributesJSON,
			&combatResistanceJSON,
		); err != nil {
			log.Printf("[Duel] 扫描对手数据失败: %v", err)
			return nil, err
		}

		opponent["id"] = id
		opponent["name"] = playerName
		opponent["level"] = level
		opponent["realm"] = realm
		opponent["cultivation"] = cultivation
		opponent["maxCultivation"] = maxCultivation
		opponent["spiritStones"] = spiritStones

		// 解析 JSON 属性
		var baseAttrs map[string]interface{}
		if err := json.Unmarshal(baseAttributesJSON, &baseAttrs); err != nil {
			baseAttrs = make(map[string]interface{})
		}
		opponent["baseAttributes"] = baseAttrs

		var combatAttrs map[string]interface{}
		if err := json.Unmarshal(combatAttributesJSON, &combatAttrs); err != nil {
			combatAttrs = make(map[string]interface{})
		}
		opponent["combatAttributes"] = combatAttrs

		var combatRes map[string]interface{}
		if err := json.Unmarshal(combatResistanceJSON, &combatRes); err != nil {
			combatRes = make(map[string]interface{})
		}
		opponent["combatResistance"] = combatRes

		opponents = append(opponents, opponent)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[Duel] 遍历结果集失败: %v", err)
		return nil, err
	}

	return opponents, nil
}

// GetPlayerBattleData 获取玩家的战斗数据
func GetPlayerBattleData(playerID int64) (gin.H, error) {
	query := `
	SELECT 
		u.id,
		u.player_name as playerName,
		u.level,
		u.realm,
		u.cultivation,
		u.max_cultivation as maxCultivation,
		u.base_attributes,
		u.combat_attributes,
		u.combat_resistance,
		u.special_attributes
	FROM users u
	WHERE u.id = ?
	`

	battleData := gin.H{}
	var id int64
	var playerName string
	var level int
	var realm string
	var cultivation, maxCultivation sql.NullFloat64
	var baseAttributesJSON, combatAttributesJSON, combatResistanceJSON, specialAttributesJSON []byte

	row := DB.Raw(query, playerID).Row()
	err := row.Scan(
		&id,
		&playerName,
		&level,
		&realm,
		&cultivation,
		&maxCultivation,
		&baseAttributesJSON,
		&combatAttributesJSON,
		&combatResistanceJSON,
		&specialAttributesJSON,
	)

	// 将扫描的值赋给 map
	battleData["id"] = id
	battleData["playerName"] = playerName
	battleData["level"] = level
	battleData["realm"] = realm
	battleData["cultivation"] = cultivation.Float64
	battleData["maxCultivation"] = maxCultivation.Float64

	if err == sql.ErrNoRows {
		return nil, errors.New("玩家不存在")
	}
	if err != nil {
		log.Printf("[Duel] 获取玩家战斗数据失败: %v", err)
		return nil, err
	}

	// 解析 JSON 属性
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(baseAttributesJSON, &baseAttrs); err != nil {
		baseAttrs = make(map[string]interface{})
	}
	battleData["baseAttributes"] = baseAttrs

	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(combatAttributesJSON, &combatAttrs); err != nil {
		combatAttrs = make(map[string]interface{})
	}
	battleData["combatAttributes"] = combatAttrs

	var combatRes map[string]interface{}
	if err := json.Unmarshal(combatResistanceJSON, &combatRes); err != nil {
		combatRes = make(map[string]interface{})
	}
	battleData["combatResistance"] = combatRes

	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(specialAttributesJSON, &specialAttrs); err != nil {
		specialAttrs = make(map[string]interface{})
	}
	battleData["specialAttributes"] = specialAttrs

	return battleData, nil
}

// GetDuelRecords 获取玩家的斗法战绩和统计
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

	rows, err := DB.Raw(query, playerID, limit, offset).Rows()
	if err != nil {
		log.Printf("[Duel] 查询战斗记录失败: %v", err)
		return nil, nil, err
	}
	defer rows.Close()

	var records []gin.H
	for rows.Next() {
		record := gin.H{}
		var id int64
		var opponentId int64
		var opponent, result, battleType, rewards string
		var createdAt time.Time

		if err := rows.Scan(
			&id,
			&opponentId,
			&opponent,
			&result,
			&battleType,
			&rewards,
			&createdAt,
		); err != nil {
			log.Printf("[Duel] 扫描战斗记录失败: %v", err)
			return nil, nil, err
		}

		record["id"] = id
		record["opponentId"] = opponentId
		record["opponent"] = opponent
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

	var totalBattles, wins, losses int64
	row := DB.Raw(statsQuery, playerID).Row()
	err = row.Scan(&totalBattles, &wins, &losses)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[Duel] 查询战斗统计失败: %v", err)
		return nil, nil, err
	}

	// 计算胜率
	winRate := int64(0)
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
	log.Printf("[Duel DB] 准备记录战斗结果: PlayerID=%d, OpponentID=%d, Result=%s, BattleType=%s, OpponentName=%s",
		battleRecord.PlayerID,
		battleRecord.OpponentID,
		battleRecord.Result,
		battleRecord.BattleType,
		battleRecord.OpponentName)

	query := `
    INSERT INTO battle_records 
    (player_id, opponent_id, opponent_name, result, battle_type, rewards, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	result := DB.Exec(
		query,
		battleRecord.PlayerID,
		battleRecord.OpponentID,
		battleRecord.OpponentName,
		battleRecord.Result,
		battleRecord.BattleType,
		battleRecord.Rewards,
		time.Now(),
	)

	if result.Error != nil {
		log.Printf("[Duel DB] 记录战斗结果失败，错误详情: %v", result.Error)
		log.Printf("[Duel DB] SQL 执行日志: %s", result.Statement.SQL.String())
		return result.Error
	}

	log.Printf("[Duel DB] 战斗结果已成功记录，受影响行数: %d", result.RowsAffected)
	return nil
}

// GetBothPlayersAttributesForBattle 获取两个玩家的战斗属性数据
func GetBothPlayersAttributesForBattle(playerID, opponentID int64) (gin.H, gin.H, error) {
	playerData, err := GetPlayerBattleData(playerID)
	if err != nil {
		log.Printf("[Duel] 获取玩家 %d 的战斗数据失败: %v", playerID, err)
		return nil, nil, err
	}

	opponentData, err := GetPlayerBattleData(opponentID)
	if err != nil {
		log.Printf("[Duel] 获取对手 %d 的战斗数据失败: %v", opponentID, err)
		return nil, nil, err
	}

	return playerData, opponentData, nil
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
