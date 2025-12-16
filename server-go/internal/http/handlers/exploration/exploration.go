package exploration

import (
	"net/http"

	explorationSvc "xiuxian/server-go/internal/exploration"

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

	// 先检查灵力是否满足（每次探索消耗100点灵力）
	if err := service.CheckSpiritCost(); err != nil {
		if err.Error() == explorationSvc.ErrInsufficientSpirit.Error() {
			zapLogger.Warn("exploration failed: insufficient spirit",
				zap.Uint("userID", uid))
			c.JSON(http.StatusBadRequest, ExplorationResponse{
				Success: false,
				Error:   "探索失败灵力不足",
			})
			return
		}
		// 其他错误
		zapLogger.Error("check spirit cost failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "探索失败", "error": err.Error()})
		return
	}

	events, log, err := service.StartExploration()
	if err != nil {
		zapLogger.Error("exploration failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "探索失败", "error": err.Error()})
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
