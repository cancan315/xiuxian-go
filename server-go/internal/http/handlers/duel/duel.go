package duel

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/duel"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"

	"github.com/gin-gonic/gin"
)

// GetDuelSpiritCost 获取当前等级的斗法灵力消耗信息
// 对应 GET /api/duel/spirit-cost
func GetDuelSpiritCost(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	var user models.User
	if err := db.DB.Select("level, spirit").First(&user, userIDInt64).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取玩家信息失败",
			"error":   err.Error(),
		})
		return
	}

	duelCost := calculateDuelSpiritCost(user.Level)
	enoughSpirit := user.Spirit >= duelCost

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"level":          user.Level,
			"currentSpirit":  user.Spirit,
			"duelCost":       duelCost,
			"enoughSpirit":   enoughSpirit,
			"spiritRequired": duelCost - user.Spirit, // 还需要的灵力（如果不足）
		},
	})
}

// GetDuelOpponents 获取斗法对手列表（随机选择在线玩家）
// 对应 GET /api/duel/opponents
func GetDuelOpponents(c *gin.Context) {
	// 从context中获取当前用户ID
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 从数据库获取其他玩家列表（排除自己）
	opponents, err := db.GetDuelOpponents(userIDInt64, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取对手列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"opponents": opponents,
			"page":      page,
			"pageSize":  pageSize,
		},
	})
}

// GetPlayerBattleData 获取指定玩家的战斗数据
// 对应 GET /api/duel/player/:playerId/battle-data
func GetPlayerBattleData(c *gin.Context) {
	playerIDStr := c.Param("playerId")
	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的玩家ID",
		})
		return
	}

	// 从数据库获取玩家战斗数据
	battleData, err := db.GetPlayerBattleData(playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取玩家战斗数据失败",
			"error":   err.Error(),
		})
		return
	}

	if battleData == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "玩家不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    battleData,
	})
}

// GetBattleAttributes 获取斗法双方的完整战斗属性数据
// 对应 POST /api/duel/battle-attributes
// 请求体: { "playerID": 123, "opponentID": 456 }
func GetBattleAttributes(c *gin.Context) {
	log.Println("[Duel Handler] === GetBattleAttributes 请求开始 ===")

	var req struct {
		PlayerID   int64 `json:"playerID" binding:"required"`
		OpponentID int64 `json:"opponentID" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Duel Handler] JSON 绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[Duel Handler] 成功解析请求: PlayerID=%d, OpponentID=%d", req.PlayerID, req.OpponentID)

	// 从数据库获取双方的战斗属性
	playerData, opponentData, err := db.GetBothPlayersAttributesForBattle(req.PlayerID, req.OpponentID)
	if err != nil {
		log.Printf("[Duel Handler] 获取战斗属性失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取战斗属性失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[Duel Handler] 成功获取战斗属性，返回数据")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"playerData":   playerData,
			"opponentData": opponentData,
		},
	})
}

// GetDuelRecords 获取当前玩家的斗法战绩
// 对应 GET /api/duel/records
func GetDuelRecords(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 从数据库获取玩家的斗法记录和统计
	records, stats, err := db.GetDuelRecords(userIDInt64, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取战绩失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"records":  records,
			"stats":    stats,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// RecordBattleResult 记录战斗结果
// 对应 POST /api/duel/record-result
func RecordBattleResult(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	var req struct {
		OpponentID   int64  `json:"opponentId" binding:"required"`
		OpponentName string `json:"opponentName" binding:"required"`
		Result       string `json:"result" binding:"required,oneof=胜利 失败"`
		BattleType   string `json:"battleType" binding:"required,oneof=pvp pve"`
		Rewards      string `json:"rewards"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Duel Handler] RecordBattleResult JSON 绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[Duel Handler] 记录战斗结果: PlayerID=%d, OpponentID=%d, Result=%s, BattleType=%s",
		userIDInt64, req.OpponentID, req.Result, req.BattleType)

	// 记录战斗结果到数据库
	if err := db.RecordBattleResult(&models.BattleRecord{
		PlayerID:     userIDInt64,
		OpponentID:   req.OpponentID,
		OpponentName: req.OpponentName,
		Result:       req.Result,
		BattleType:   req.BattleType,
		Rewards:      req.Rewards,
	}); err != nil {
		log.Printf("[Duel Handler] 记录战斗结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "记录战斗结果失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[Duel Handler] 战斗结果已记录")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "战斗结果已记录",
	})
}

// ClaimBattleRewards 领取战斗奖励
// 对应 POST /api/duel/claim-rewards
func ClaimBattleRewards(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	var req struct {
		Rewards []string `json:"rewards" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 处理奖励
	if err := db.ClaimBattleRewards(userIDInt64, req.Rewards); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "领取奖励失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "奖励已领取",
	})
}

// StartPvPBattle 开始PvP战斗
// 对应 POST /api/duel/start-pvp
func StartPvPBattle(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	// 检查每日斗法次数限制
	if err, remaining := checkDailyDuelLimit(userIDInt64); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success":   false,
			"message":   err.Error(),
			"remaining": remaining,
		})
		return
	}

	// 检查玩家灵力是否足够
	var user models.User
	if err := db.DB.First(&user, userIDInt64).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取玩家信息失败",
			"error":   err.Error(),
		})
		return
	}

	// 检查灵力是否足够
	enoughSpirit, duelCost, errMsg := checkDuelSpirit(&user)
	if !enoughSpirit {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":       false,
			"message":       errMsg,
			"currentSpirit": user.Spirit,
			"duelCost":      duelCost,
		})
		return
	}

	var req struct {
		OpponentID   int64       `json:"opponentId" binding:"required"`
		PlayerData   interface{} `json:"playerData" binding:"required"`
		OpponentData interface{} `json:"opponentData" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 创建PvP战斗服务
	battleService := duel.NewPvPBattleService(userIDInt64, req.OpponentID)

	// 开始战斗
	roundData, err := battleService.StartPvPBattle(req.PlayerData, req.OpponentData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "开始战斗失败",
			"error":   err.Error(),
		})
		return
	}

	// 战斗开始成功，扣除玩家灵力
	if _, err := deductDuelSpirit(userIDInt64, duelCost); err != nil {
		log.Printf("[Duel] 扣除玩家灵力失败: %v", err)
		// 扣除灵力失败不中断战斗流程，仅记录日志
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "战斗已开始",
		"data":    roundData,
	})
}

// ExecutePvPRound 执行PvP战斗回合
// 对应 POST /api/duel/execute-pvp-round
func ExecutePvPRound(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	var req struct {
		OpponentID int64 `json:"opponentId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 创建PvP战斗服务
	battleService := duel.NewPvPBattleService(userIDInt64, req.OpponentID)

	// 执行回合
	roundData, err := battleService.ExecutePvPRound()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "执行回合失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roundData,
	})
}

// EndPvPBattle 结束PvP战斗
// 对应 POST /api/duel/end-pvp
func EndPvPBattle(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userID := userIDInterface.(uint)
	userIDInt64 := int64(userID)

	var req struct {
		OpponentID int64 `json:"opponentId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 创建PvP战斗服务
	battleService := duel.NewPvPBattleService(userIDInt64, req.OpponentID)

	// 加载战斗状态
	status, err := battleService.LoadBattleStatusFromRedis()
	if err == nil && status != nil {
		// 战斗已结束，奖励已处理
		if status.PlayerHealth <= 0 || status.OpponentHealth <= 0 {
			// 注：奖励已在 ExecutePvPRound 中发放，这里无需重复处理
		}
	}

	// 清除战斗状态
	if err := battleService.ClearBattleStatusFromRedis(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清除战斗状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "战斗已结束",
	})
}

// ========== 斗法次数限制辅助函数 ==========

// checkDailyDuelLimit 检查每日斗法次数限制
// 返回 (error, remaining) - 如果error为nil表示通过检查，remaining为剩余次数
func checkDailyDuelLimit(userID int64) (error, int) {
	const maxDailyDuels = 20

	// 获取今天的日期字符串（格式: 2025-12-29）
	today := time.Now().Format("2006-01-02")
	duelCountKey := "duel:daily:" + today + ":" + strconv.FormatInt(userID, 10)

	// 从Redis获取今天已经斗法的次数
	countStr, err := redis.Client.Get(redis.Ctx, duelCountKey).Result()

	var duelCount int
	if err != nil {
		// Redis中不存在该键，说明是新的一天
		duelCount = 0
	} else {
		duelCount, _ = strconv.Atoi(countStr)
	}

	// 检查是否超过限制
	if duelCount >= maxDailyDuels {
		log.Printf("[Duel] 玩家 %d 今日斗法次数已满(%d/%d)", userID, duelCount, maxDailyDuels)
		return fmt.Errorf("今日斗法次数已达上限(20/20)，请明天再来！"), 0
	}

	// 增加计数
	newCount := duelCount + 1
	remaining := maxDailyDuels - newCount

	// 设置Redis键值，过期时间为今天剩余的秒数
	timeUntilMidnight := getTimeUntilMidnight()
	redis.Client.Set(redis.Ctx, duelCountKey, newCount, timeUntilMidnight)

	log.Printf("[Duel] 玩家 %d 斗法次数: %d/%d, 剩余: %d", userID, newCount, maxDailyDuels, remaining)

	return nil, remaining
}

// getTimeUntilMidnight 获取距离今天00:00:00的时间差
func getTimeUntilMidnight() time.Duration {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
	return midnight.Sub(now)
}

// getDailyDuelCount 获取玩家今日斗法次数
func getDailyDuelCount(userID int64) int {
	today := time.Now().Format("2006-01-02")
	duelCountKey := "duel:daily:" + today + ":" + strconv.FormatInt(userID, 10)

	countStr, err := redis.Client.Get(redis.Ctx, duelCountKey).Result()
	if err != nil {
		return 0
	}

	count, _ := strconv.Atoi(countStr)
	return count
}

// ========== 斗法灵力消耗计算函数 ==========

// calculateDuelSpritCost 计算斗法灵力消耗
// 公式: duelCost = 1440 * 1.2^(Level-1)
func calculateDuelSpiritCost(level int) float64 {
	const baseCost = 1440.0
	const costMultiplier = 1.2
	return baseCost * math.Pow(costMultiplier, float64(level-1))
}

// checkDuelSpirit 检查玩家灵力是否足够进行斗法
// 返回 (足够, 需要消耗的灵力, 错误消息)
func checkDuelSpirit(user *models.User) (bool, float64, string) {
	duelCost := calculateDuelSpiritCost(user.Level)
	if user.Spirit < duelCost {
		requiredMore := duelCost - user.Spirit
		errMsg := fmt.Sprintf("灵力不足！当前灵力: %.0f，斗法消耗: %.0f，还差: %.0f",
			user.Spirit, duelCost, requiredMore)
		return false, duelCost, errMsg
	}
	return true, duelCost, ""
}

// deductDuelSpirit 扣除玩家斗法灵力消耗
// 返回 (更新后的灵力, 错误)
func deductDuelSpirit(userID int64, duelCost float64) (float64, error) {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return 0, fmt.Errorf("获取玩家信息失败: %w", err)
	}

	// 扣除灵力
	user.Spirit -= duelCost

	// 确保灵力不为负
	if user.Spirit < 0 {
		user.Spirit = 0
	}

	// 保存到数据库
	if err := db.DB.Model(&user).Update("spirit", user.Spirit).Error; err != nil {
		return 0, fmt.Errorf("扣除灵力失败: %w", err)
	}

	log.Printf("[Duel] 玩家 %d 消耗灵力 %.0f，剩余灵力: %.0f",
		userID, duelCost, user.Spirit)

	return user.Spirit, nil
}
