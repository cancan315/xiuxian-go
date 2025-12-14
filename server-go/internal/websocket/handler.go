package websocket

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该更严格地检查Origin
		return true
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	manager *ConnectionManager
	logger  *zap.Logger
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(manager *ConnectionManager, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
		logger:  logger,
	}
}

// Upgrade 升级HTTP连接为WebSocket
func (h *WebSocketHandler) Upgrade(c *gin.Context) {
	// 从请求头获取userID和token
	userIDStr := c.Query("userId")
	token := c.Query("token")

	if userIDStr == "" || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少userId或token"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的userId"})
		return
	}

	// 这里应该验证token（与HTTP认证中间件一致）
	// 为了简化，这里假设token已在中间件中验证
	username := c.GetString("username")
	if username == "" {
		username = "玩家" + userIDStr
	}

	// 升级连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	// 注册连接
	h.manager.RegisterClient(uint(userID), username, conn)
	h.logger.Info("WebSocket连接成功",
		zap.Uint("userID", uint(userID)),
		zap.String("username", username))
}

// Stats 获取连接统计
func (h *WebSocketHandler) Stats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"onlineCount": h.manager.GetOnlineCount(),
		"timestamp":   getCurrentTime(),
	})
}

// 辅助函数
func getCurrentTime() int64 {
	return time.Now().Unix()
}

// ExtractUserID 从Authorization头提取userID
func ExtractUserID(authHeader string) (uint, error) {
	// 假设格式为: "Bearer TOKEN"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, nil
	}

	// 这里应该解析JWT token获取userID
	// 为了简化示例，返回0
	return 0, nil
}
