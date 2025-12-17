package cultivation

import (
	"net/http"

	"xiuxian/server-go/internal/cache"
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

	// ✅ 注入灵力增长管理器以支持缓存灵力读取
	if spiritMgr, ok := c.Get("spirit_manager"); ok {
		if mgr, ok := spiritMgr.(interface{ GetPlayerSpiritFromCache(uint) float64 }); ok {
			service.SetSpiritGrowManager(mgr)
		}
	}

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

	// ✅ 异步删除修炼缓存（确保数据最新）
	go func() {
		if err := cache.DeleteCultivationDataCache(uid, zapLogger); err != nil {
			zapLogger.Warn("删除修炼缓存失败", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, resp)
}

// CultivateUntilBreakthrough 一键突破
// POST /api/cultivation/breakthrough
func CultivateUntilBreakthrough(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("CultivateUntilBreakthrough 入参",
		zap.Uint("userID", uid))

	service := cultivationSvc.NewCultivationService(uid)
	resp, err := service.CultivateUntilBreakthrough()
	if err != nil {
		zapLogger.Error("breakthrough failed",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "突破失败", "error": err.Error()})
		return
	}

	zapLogger.Info("CultivateUntilBreakthrough 出参",
		zap.Uint("userID", uid),
		zap.Bool("success", resp.Success))

	// ✅ 异步删除修炼缓存（确保数据最新）
	go func() {
		if err := cache.DeleteCultivationDataCache(uid, zapLogger); err != nil {
			zapLogger.Warn("删除修炼缓存失败", zap.Error(err))
		}
	}()

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

	// 优先从缓存获取修炼数据
	cachedData, err := cache.GetCultivationDataFromCache(uid, zapLogger)
	if err == nil && cachedData != nil {
		zapLogger.Debug("从缓存获取修炼数据", zap.Uint("userID", uid))
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    cachedData,
		})
		return
	}

	// 缓存不存在或获取失败，从数据库查询
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

	// 异步存储到缓存（不影响响应）
	go func() {
		if err := cache.SetCultivationDataCache(uid, data, zapLogger); err != nil {
			zapLogger.Warn("缓存修炼数据失败", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
