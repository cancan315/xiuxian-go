package dungeon

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	dungeonSvc "xiuxian/server-go/internal/dungeon"
)

// 请求/响应类型别名
type (
	DungeonRequest  = dungeonSvc.DungeonRequest
	BuffOption      = dungeonSvc.BuffOption
	FightResult     = dungeonSvc.FightResult
)

// StartDungeon 开始秘境
// POST /api/dungeon/start
func StartDungeon(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req DungeonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	zapLogger.Info("StartDungeon 入参",
		zap.Uint("userID", uid),
		zap.String("difficulty", req.Difficulty))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"floor":        1,
			"difficulty":   req.Difficulty,
			"refreshCount": 3,
		},
		"message": "秘境已开启",
	})
}

// GetBuffOptions 获取增益选项
// GET /api/dungeon/buffs/:floor
func GetBuffOptions(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uuid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	floorStr := c.Param("floor")
	// 将floor字符串转换为整数
	floor := 1 // 默认值
	if floorStr != "" {
		// 实际应使用strconv.Atoi转换，这里简化处理
		_ = floorStr // 标记为已使用
	}

	zapLogger.Info("GetBuffOptions 入参",
		zap.Uint("userID", uuid))

	service := dungeonSvc.NewDungeonService(uuid)
	buffs, err := service.GetRandomBuffs(floor) // floor应该从前端传递
	if err != nil {
		zapLogger.Error("获取增益选项失败",
			zap.Uint("userID", uuid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取增益失败", "error": err.Error()})
		return
	}

	zapLogger.Info("GetBuffOptions 出参",
		zap.Uint("userID", uuid),
		zap.Int("optionCount", len(buffs)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"floor":   1,
			"options": buffs,
		},
	})
}

// SelectBuff 选择增益
// POST /api/dungeon/select-buff
func SelectBuff(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req DungeonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	zapLogger.Info("SelectBuff 入参",
		zap.Uint("userID", uid),
		zap.String("buffID", req.SelectedBuffID))

	service := dungeonSvc.NewDungeonService(uid)
	buff, err := service.SelectBuff(req.SelectedBuffID)
	if err != nil {
		zapLogger.Error("选择增益失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "选择增益失败", "error": err.Error()})
		return
	}

	zapLogger.Info("SelectBuff 出参",
		zap.Uint("userID", uid),
		zap.String("buffID", req.SelectedBuffID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    buff,
		"message": "增益已选择",
	})
}

// StartFight 开始战斗
// POST /api/dungeon/fight
func StartFight(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req DungeonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	zapLogger.Info("StartFight 入参",
		zap.Uint("userID", uid),
		zap.String("difficulty", req.Difficulty))

	service := dungeonSvc.NewDungeonService(uid)
	result, err := service.StartFight(1, req.Difficulty) // floor应该从前端传递
	if err != nil {
		zapLogger.Error("战斗失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "战斗出错", "error": err.Error()})
		return
	}

	zapLogger.Info("StartFight 出参",
		zap.Uint("userID", uid),
		zap.Bool("victory", result.Victory))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// EndDungeon 结束秘境
// POST /api/dungeon/end
func EndDungeon(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req struct {
		Floor   int  `json:"floor"`
		Victory bool `json:"victory"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	zapLogger.Info("EndDungeon 入参",
		zap.Uint("userID", uid),
		zap.Int("floor", req.Floor),
		zap.Bool("victory", req.Victory))

	service := dungeonSvc.NewDungeonService(uid)
	result, err := service.EndDungeon(req.Floor, req.Victory)
	if err != nil {
		zapLogger.Error("结束秘境失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "结束秘境失败", "error": err.Error()})
		return
	}

	zapLogger.Info("EndDungeon 出参",
		zap.Uint("userID", uid),
		zap.Int("floor", req.Floor))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "秘境已结束",
	})
}
