package online

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisc "xiuxian/server-go/internal/redis"
)

// helper 获取 Redis 客户端，便于未来扩展
func client() *redisv9.Client {
	return redisc.Client
}

// Login 标记玩家上线，对应 POST /api/online/login
// 会初始化玩家的灵力增长时间戳
func Login(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId"`
		IP       string `json:"ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 将playerId字符串转换为uint
	parsedID, err := strconv.ParseUint(req.PlayerID, 10, 32)
	if err != nil {
		zapLogger.Warn("无效的玩家ID", zap.String("playerID", req.PlayerID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的玩家ID"})
		return
	}
	playerID := uint(parsedID)

	// 获取玩家数据並初始化灵力增长时间戳
	var user models.User
	if err := db.DB.First(&user, playerID).Error; err != nil {
		zapLogger.Warn("玩家不存在", zap.String("playerID", req.PlayerID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "玩家不存在"})
		return
	}

	// ✅ 初始化灵力增长时间戳为当前时间
	// 注意：不论是第一次登录还是再次登录，都需要重置
	// 以避免離線期間的时間一次性計入灵力增长
	now := time.Now()
	if err := db.DB.Model(&user).Update("last_spirit_gain_time", now).Error; err != nil {
		zapLogger.Error("重置玩家灵力增长时间戳失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
	}

	zapLogger.Info("重置玩家灵力增长时间戳",
		zap.Uint("userID", user.ID),
		zap.Time("LastSpiritGainTime", now))

	rdb := client()
	loginTime := time.Now().UnixMilli()
	key := "player:online:" + req.PlayerID

	fields := map[string]interface{}{
		"playerId":      req.PlayerID,
		"loginTime":     strconv.FormatInt(loginTime, 10),
		"lastHeartbeat": strconv.FormatInt(loginTime, 10),
		"ip":            req.IP,
		"status":        "online",
	}

	if err := rdb.HSet(redisc.Ctx, key, fields).Err(); err != nil {
		zapLogger.Error("设置Redis数据失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.Expire(redisc.Ctx, key, 15*time.Second).Err(); err != nil {
		zapLogger.Error("Redis数据超时设置失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.SAdd(redisc.Ctx, "server:online:players", req.PlayerID).Err(); err != nil {
		zapLogger.Error("添加到在线玩家集合失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	zapLogger.Info("玩家上线成功",
		zap.Uint("userID", user.ID),
		zap.String("playerID", req.PlayerID))

	c.JSON(http.StatusOK, gin.H{"message": "成功标记为在线", "playerId": req.PlayerID, "loginTime": loginTime})
}

// Heartbeat 更新心跳，对应 POST /api/online/heartbeat
func Heartbeat(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId"`
	}

	// 添加详细的日志来诊断问题
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	rawData, _ := c.GetRawData()
	zapLogger.Info("收到心跳请求",
		zap.String("rawBody", string(rawData)),
		zap.Any("headers", c.Request.Header))

	// 重新设置body以便后续解析
	c.Request.Body = io.NopCloser(strings.NewReader(string(rawData)))

	if err := c.ShouldBindJSON(&req); err != nil {
		zapLogger.Error("心跳请求参数解析失败",
			zap.Error(err),
			zap.String("rawBody", string(rawData)))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误: " + err.Error()})
		return
	}

	if req.PlayerID == "" {
		zapLogger.Warn("心跳请求缺少玩家ID",
			zap.String("receivedPlayerId", req.PlayerID),
			zap.String("rawBody", string(rawData)))
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

	rdb := client()
	key := "player:online:" + req.PlayerID

	exists, err := rdb.Exists(redisc.Ctx, key).Result()
	if err != nil {
		zapLogger.Error("检查玩家在线状态失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if exists == 0 {
		zapLogger.Warn("玩家不在线",
			zap.String("playerID", req.PlayerID))
		c.JSON(http.StatusNotFound, gin.H{"error": "玩家不在线"})
		return
	}

	lastHeartbeat := time.Now().UnixMilli()
	if err := rdb.HSet(redisc.Ctx, key, "lastHeartbeat", strconv.FormatInt(lastHeartbeat, 10)).Err(); err != nil {
		zapLogger.Error("更新心跳时间失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.Expire(redisc.Ctx, key, 15*time.Second).Err(); err != nil {
		zapLogger.Error("更新Redis超时失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	zapLogger.Info("心跳更新成功",
		zap.String("playerID", req.PlayerID),
		zap.Int64("lastHeartbeat", lastHeartbeat))

	c.JSON(http.StatusOK, gin.H{"message": "心跳更新成功", "playerId": req.PlayerID, "lastHeartbeat": lastHeartbeat})
}

// Logout 标记玩家离线，对应 POST /api/online/logout
func Logout(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 将playerID从字符串转换为uint
	parsedID, err := strconv.ParseUint(req.PlayerID, 10, 32)
	if err != nil {
		zapLogger.Warn("无效的玩家ID", zap.String("playerID", req.PlayerID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的玩家ID"})
		return
	}
	playerID := uint(parsedID)

	rdb := client()
	key := "player:online:" + req.PlayerID

	exists, err := rdb.Exists(redisc.Ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "玩家不在线"})
		return
	}

	// ✅ 简化离线处理：只删除Redis key，依赖Redis TTL机制
	zapLogger.Info("开始离线处理",
		zap.String("playerID", req.PlayerID))

	// 直接删除Redis key
	if err := rdb.Del(redisc.Ctx, key).Err(); err != nil {
		zapLogger.Error("删除Redis数据失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	// 从在线玩家集合中移除玩家ID
	if err := rdb.SRem(redisc.Ctx, "server:online:players", req.PlayerID).Err(); err != nil {
		zapLogger.Error("从在线集合中移除失败",
			zap.String("playerID", req.PlayerID),
			zap.Error(err))
		// 不中断，继续执行
	}

	// ✅ 成功离线
	zapLogger.Info("玩家离线成功",
		zap.Uint("userID", playerID),
		zap.String("playerID", req.PlayerID))

	logoutTime := time.Now().UnixMilli()

	c.JSON(http.StatusOK, gin.H{"message": "成功标记为离线", "playerId": req.PlayerID, "logoutTime": logoutTime})
}

// GetOnlinePlayers 获取在线玩家列表，对应 GET /api/online/players
func GetOnlinePlayers(c *gin.Context) {
	rdb := client()

	ids, err := rdb.SMembers(redisc.Ctx, "server:online:players").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	players := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		key := "player:online:" + id
		data, err := rdb.HGetAll(redisc.Ctx, key).Result()
		if err != nil {
			continue
		}
		if status, ok := data["status"]; !ok || status != "online" {
			continue
		}
		loginTime, _ := strconv.ParseInt(data["loginTime"], 10, 64)
		lastHeartbeat, _ := strconv.ParseInt(data["lastHeartbeat"], 10, 64)
		players = append(players, gin.H{
			"playerId":      data["playerId"],
			"loginTime":     loginTime,
			"lastHeartbeat": lastHeartbeat,
			"ip":            data["ip"],
		})
	}

	c.JSON(http.StatusOK, gin.H{"players": players})
}

// GetPlayerOnlineStatus 获取指定玩家在线状态，对应 GET /api/online/player/:playerId
func GetPlayerOnlineStatus(c *gin.Context) {
	playerID := c.Param("playerId")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

	rdb := client()
	key := "player:online:" + playerID
	data, err := rdb.HGetAll(redisc.Ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "玩家不在线"})
		return
	}

	loginTime, _ := strconv.ParseInt(data["loginTime"], 10, 64)
	lastHeartbeat, _ := strconv.ParseInt(data["lastHeartbeat"], 10, 64)

	c.JSON(http.StatusOK, gin.H{
		"playerId":      data["playerId"],
		"status":        data["status"],
		"loginTime":     loginTime,
		"lastHeartbeat": lastHeartbeat,
		"ip":            data["ip"],
	})
}
