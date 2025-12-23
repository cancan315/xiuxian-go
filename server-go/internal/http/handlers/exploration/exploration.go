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
