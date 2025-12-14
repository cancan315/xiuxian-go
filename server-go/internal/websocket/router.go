package websocket

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterWebSocketRoutes 注册WebSocket路由
func RegisterWebSocketRoutes(router *gin.Engine, manager *ConnectionManager, logger *zap.Logger) {
	handler := NewWebSocketHandler(manager, logger)

	// WebSocket升级端点
	router.GET("/ws", handler.Upgrade)

	// WebSocket统计信息端点（可选）
	router.GET("/ws/stats", handler.Stats)
}

// InitializeHandlers 初始化所有事件处理器
func InitializeHandlers(manager *ConnectionManager, logger *zap.Logger) *Handlers {
	return &Handlers{
		Spirit:      NewSpiritHandler(manager, logger),
		Dungeon:     NewDungeonHandler(manager, logger),
		Leaderboard: NewLeaderboardHandler(manager, logger),
		Exploration: NewExplorationHandler(manager, logger),
	}
}

// Handlers 包含所有WebSocket事件处理器
type Handlers struct {
	Spirit      *SpiritHandler
	Dungeon     *DungeonHandler
	Leaderboard *LeaderboardHandler
	Exploration *ExplorationHandler
}
