package cultivation

import (
	"net/http"

	cultivationSvc "xiuxian/server-go/internal/cultivation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 请求/响应类型别名
type (
	CultivationRequest      = cultivationSvc.CultivationRequest
	CultivationResponse     = cultivationSvc.CultivationResponse
	AutoCultivationResponse = cultivationSvc.AutoCultivationResponse
	CultivationData         = cultivationSvc.CultivationData
)

// SingleCultivate 单次打坐修炼
// POST /api/cultivation/single
func SingleCultivate(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("SingleCultivate 入参",
		zap.Uint("userID", uid))

	service := cultivationSvc.NewCultivationService(uid)

	resp, err := service.SingleCultivate()
	if err != nil {
		zapLogger.Error("cultivation failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "修炼失败", "error": err.Error()})
		return
	}

	zapLogger.Info("SingleCultivate 出参",
		zap.Uint("userID", uid),
		zap.Float64("cultivationGain", resp.CultivationGain))

	c.JSON(http.StatusOK, resp)
}

// GetCultivationData 获取修炼数据
// GET /api/cultivation/data
func GetCultivationData(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("GetCultivationData 入参",
		zap.Uint("userID", uid))

	// 从数据库查询修炼数据
	service := cultivationSvc.NewCultivationService(uid)
	data, err := service.GetCultivationData()
	if err != nil {
		zapLogger.Error("get cultivation data failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取数据失败", "error": err.Error()})
		return
	}

	zapLogger.Info("GetCultivationData 出参",
		zap.Uint("userID", uid),
		zap.Int("level", data.Level),
		zap.String("realm", data.Realm))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// UseFormation 使用聚灵阵
// POST /api/cultivation/formation
func UseFormation(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("UseFormation 入参",
		zap.Uint("userID", uid))

	service := cultivationSvc.NewCultivationService(uid)

	resp, err := service.UseFormation()
	if err != nil {
		zapLogger.Error("formation failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "聚灵阵使用失败", "error": err.Error()})
		return
	}

	zapLogger.Info("UseFormation 出参",
		zap.Uint("userID", uid),
		zap.Float64("cultivationGain", resp.CultivationGain),
		zap.Int("stoneCost", resp.StoneCost))

	c.JSON(http.StatusOK, resp)
}
