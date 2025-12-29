package exploration

import (
	"math"
	"net/http"

	"xiuxian/server-go/internal/db"
	explorationSvc "xiuxian/server-go/internal/exploration"
	"xiuxian/server-go/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 请求/响应类型别名，导出供处理器使用
type (
	ExplorationRequest  = explorationSvc.ExplorationRequest
	ExplorationResponse = explorationSvc.ExplorationResponse
	EventChoiceRequest  = explorationSvc.EventChoiceRequest
	EventChoiceResponse = explorationSvc.EventChoiceResponse
)

// StartExploration 开始探索
// POST /api/exploration/start
func StartExploration(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("StartExploration 入参",
		zap.Uint("userID", uid))

	service := explorationSvc.NewExplorationService(uid)

	events, log, err := service.StartExploration()
	if err != nil {
		zapLogger.Error("exploration failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		// 返回详细的错误信息给前端，包括修为检查、灵力检查等所有业务逻辑错误
		c.JSON(http.StatusBadRequest, ExplorationResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	zapLogger.Info("StartExploration 出参",
		zap.Uint("userID", uid),
		zap.Int("eventsCount", len(events)))

	c.JSON(http.StatusOK, ExplorationResponse{
		Success: true,
		Events:  events,
		Log:     log,
	})
}

// HandleEventChoice 处理事件选择
// POST /api/exploration/event-choice
func HandleEventChoice(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req EventChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	zapLogger.Info("HandleEventChoice 入参",
		zap.Uint("userID", uid),
		zap.String("eventType", req.EventType))

	service := explorationSvc.NewExplorationService(uid)
	rewards, err := service.HandleEventChoice(req.EventType, req.Choice)
	if err != nil {
		zapLogger.Error("handle event choice failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "处理失败", "error": err.Error()})
		return
	}

	zapLogger.Info("HandleEventChoice 出参",
		zap.Uint("userID", uid),
		zap.Any("rewards", rewards))

	c.JSON(http.StatusOK, EventChoiceResponse{
		Success: true,
		Rewards: rewards,
	})
}

// ========== 探索灵力消耗计算函数 ==========

// calculateExploreSpritCost 计算探索灵力消耗
// 公式: exploreCost = 288 * 1.2^(Level-1)
func calculateExploreSpritCost(level int) float64 {
	const baseCost = 288.0
	const costMultiplier = 1.2
	return baseCost * math.Pow(costMultiplier, float64(level-1))
}

// GetExploreSpritCost 获取当前等级的探索灵力消耗信息
// GET /api/exploration/spirit-cost
func GetExploreSpritCost(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var user models.User
	if err := db.DB.Select("level, spirit").First(&user, uid).Error; err != nil {
		zapLogger.Error("获取玩家信息失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取玩家信息失败",
			"error":   err.Error(),
		})
		return
	}

	explorecost := calculateExploreSpritCost(user.Level)
	enoughSpirit := user.Spirit >= explorecost

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"level":          user.Level,
			"currentSpirit":  user.Spirit,
			"exploreCost":    explorecost,
			"enoughSpirit":   enoughSpirit,
			"spiritRequired": math.Max(0, explorecost-user.Spirit), // 还需要的灵力（如果不足）
		},
	})
}
