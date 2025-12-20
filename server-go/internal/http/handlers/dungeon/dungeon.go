package dungeon

import (
	"net/http"
	"strconv"

	dungeonSvc "xiuxian/server-go/internal/dungeon"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 请求/响应类型别名
type (
	DungeonRequest = dungeonSvc.DungeonRequest
	BuffOption     = dungeonSvc.BuffOption
	FightResult    = dungeonSvc.FightResult
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

	// 验证difficulty
	validDifficulties := map[string]bool{"easy": true, "normal": true, "hard": true, "expert": true}
	if req.Difficulty == "" || !validDifficulties[req.Difficulty] {
		zapLogger.Warn("限制difficulty为nil或无效",
			zap.Uint("userID", uid),
			zap.String("difficulty", req.Difficulty))
		c.JSON(http.StatusBadRequest, gin.H{"message": "difficulty参数无效，必须是: easy, normal, hard, expert"})
		return
	}

	zapLogger.Info("StartDungeon 出参",
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
	floor, err := strconv.Atoi(floorStr)
	if err != nil || floor < 1 {
		zapLogger.Warn("Floor参数无效，使用默认值1",
			zap.String("floorStr", floorStr),
			zap.Error(err))
		floor = 1
	}

	zapLogger.Info("GetBuffOptions 入参",
		zap.Uint("userID", uuid),
		zap.Int("floor", floor))

	service := dungeonSvc.NewDungeonService(uuid)
	buffs, err := service.GetRandomBuffs(floor)
	if err != nil {
		zapLogger.Error("获取增益选项失败",
			zap.Uint("userID", uuid),
			zap.Int("floor", floor),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取增益失败", "error": err.Error()})
		return
	}

	zapLogger.Info("GetBuffOptions 出参",
		zap.Uint("userID", uuid),
		zap.Int("floor", floor),
		zap.Int("optionCount", len(buffs)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"floor":   floor,
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

	// 验证buffID
	if req.SelectedBuffID == "" {
		zapLogger.Warn("选择增益 - buffID为nil",
			zap.Uint("userID", uid))
		c.JSON(http.StatusBadRequest, gin.H{"message": "selectedBuffId参数必需"})
		return
	}

	service := dungeonSvc.NewDungeonService(uid)
	// 调用 SelectBuffAndApplyEffects 以便不仅返回增益信息，还要应用增益效果
	buff, err := service.SelectBuffAndApplyEffects(req.SelectedBuffID)
	if err != nil {
		zapLogger.Error("选择增益失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "选择增益失败", "error": err.Error()})
		return
	}

	zapLogger.Info("SelectBuff 出参",
		zap.Uint("userID", uid),
		zap.String("buffID", req.SelectedBuffID),
		zap.String("buffName", buff["name"].(string)))

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
		zap.Int("floor", req.Floor),
		zap.String("difficulty", req.Difficulty))

	// 验证difficulty
	if req.Difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "difficulty参数必需"})
		return
	}

	// 设置floor默认值
	floor := req.Floor
	if floor < 1 {
		floor = 1
	}

	service := dungeonSvc.NewDungeonService(uid)
	result, err := service.StartFight(floor, req.Difficulty)
	if err != nil {
		zapLogger.Error("战斗失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "战斗出错", "error": err.Error()})
		return
	}

	zapLogger.Info("StartFight 出参",
		zap.Uint("userID", uid),
		zap.Int("floor", floor),
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

// SaveBuff 保存已选增益
// POST /api/dungeon/save-buff
func SaveBuff(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req struct {
		Floor      int    `json:"floor"`      // 当前桂层
		Difficulty string `json:"difficulty"` // 难度
		BuffID     string `json:"buffId"`     // 选择的增益 ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	zapLogger.Info("SaveBuff 入参",
		zap.Uint("userID", uid),
		zap.String("buffID", req.BuffID),
		zap.Int("floor", req.Floor))

	// 验证buffID
	if req.BuffID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "buffId参数必需"})
		return
	}

	service := dungeonSvc.NewDungeonService(uid)

	// 尝试从 Redis 加载现有会话
	session, _ := service.LoadSessionFromRedis()

	// 更新增益选择
	buff, err := service.SelectBuffAndApplyEffects(req.BuffID)
	if err != nil {
		zapLogger.Error("保存增益失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存增益失败", "error": err.Error()})
		return
	}

	// 保存更新的会话到 Redis
	floor := req.Floor
	if session != nil {
		floor = session.Floor + 1 // 下一桂
	}
	if err := service.SaveSessionToRedis(floor, req.Difficulty); err != nil {
		zapLogger.Warn("保存秘境会话到 Redis 失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		// 继续执行，不中断流程
	}

	zapLogger.Info("SaveBuff 出参",
		zap.Uint("userID", uid),
		zap.String("buffID", req.BuffID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    buff,
		"message": "增益已保存",
	})
}

// LoadSession 加载磐境会话（用于断线重连）
// GET /api/dungeon/load-session
func LoadSession(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("LoadSession 入参",
		zap.Uint("userID", uid))

	service := dungeonSvc.NewDungeonService(uid)

	// 从 Redis 中加载会话
	session, err := service.LoadSessionFromRedis()
	if err != nil {
		// 会话不存在，返回空。前端需要检查 success 字段
		zapLogger.Info("没有找到秘境会话",
			zap.Uint("userID", uid))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不存在有效会话",
		})
		return
	}

	zapLogger.Info("LoadSession 出参",
		zap.Uint("userID", uid),
		zap.Int("floor", session.Floor))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"floor":         session.Floor,
			"difficulty":    session.Difficulty,
			"refreshCount":  session.RefreshCount,
			"selectedBuffs": session.SelectedBuffs,
			"playerBuffs":   session.PlayerBuffs,
		},
		"message": "秘境会话已恢复",
	})
}

// GetRoundData 获取单回合的战斗数据
// GET /api/dungeon/round-data
func GetRoundData(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("GetRoundData 入参",
		zap.Uint("userID", uid))

	service := dungeonSvc.NewDungeonService(uid)

	// 从Redis获取回合数据
	roundData, err := service.GetRoundDataFromRedis()
	if err != nil {
		// 回合数据不存在，返回失败
		zapLogger.Warn("GetRoundData - 回合数据不存在",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "回合数据不存在",
		})
		return
	}

	zapLogger.Info("GetRoundData 出参",
		zap.Uint("userID", uid),
		zap.Int("round", roundData.Round),
		zap.Bool("battleEnded", roundData.BattleEnded))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roundData,
	})
}

// ExecuteRound 执行单个回合战斗
// POST /api/dungeon/execute-round
func ExecuteRound(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("ExecuteRound 入参",
		zap.Uint("userID", uid))

	service := dungeonSvc.NewDungeonService(uid)

	// 执行一回合
	roundData, err := service.ExecuteRound()
	if err != nil {
		zapLogger.Error("ExecuteRound - 执行失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "执行回合失败",
			"error":   err.Error(),
		})
		return
	}

	zapLogger.Info("ExecuteRound 出参",
		zap.Uint("userID", uid),
		zap.Int("round", roundData.Round),
		zap.Bool("battleEnded", roundData.BattleEnded))

	// 保存回合数据到Redis以供前端获取
	if err := service.SaveRoundDataToRedis(roundData); err != nil {
		zapLogger.Warn("ExecuteRound - 保存回合数据失败",
			zap.Uint("userID", uid),
			zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roundData,
	})
}
