package online

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"

	redisc "xiuxian/server-go/internal/redis"
)

// helper 获取 Redis 客户端，便于未来扩展
func client() *redisv9.Client {
	return redisc.Client
}

// Login 标记玩家上线，对应 POST /api/online/login
func Login(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId"`
		IP       string `json:"ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.Expire(redisc.Ctx, key, 24*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.SAdd(redisc.Ctx, "server:online:players", req.PlayerID).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成功标记为在线", "playerId": req.PlayerID, "loginTime": loginTime})
}

// Heartbeat 更新心跳，对应 POST /api/online/heartbeat
func Heartbeat(c *gin.Context) {
	var req struct {
		PlayerID string `json:"playerId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少玩家ID"})
		return
	}

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

	lastHeartbeat := time.Now().UnixMilli()
	if err := rdb.HSet(redisc.Ctx, key, "lastHeartbeat", strconv.FormatInt(lastHeartbeat, 10)).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.Expire(redisc.Ctx, key, 24*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

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

	logoutTime := time.Now().UnixMilli()
	if err := rdb.HSet(redisc.Ctx, key, map[string]interface{}{
		"status":     "offline",
		"logoutTime": strconv.FormatInt(logoutTime, 10),
	}).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}
	if err := rdb.SRem(redisc.Ctx, "server:online:players", req.PlayerID).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

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
