package main

import (
	"log"
	"net/http"
	"os"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"context"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/http/router"
	"xiuxian/server-go/internal/redis"
	"xiuxian/server-go/internal/spirit"
	"xiuxian/server-go/internal/websocket"
)

// LoggerMiddleware 创建一个中间件，将zap logger添加到gin上下文中
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("zap_logger", logger)
		c.Next()
	}
}

func main() {
	_ = godotenv.Load()

	if err := db.Init(); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	if err := redis.Init(); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

	// 初始化 zap 日志记录器，配置日志级别
	config := zap.NewProductionConfig()
	// 可以根据环境变量调整日志级别，如果没有设置则默认为 debug 级别
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, _ := config.Build()
	defer logger.Sync()

	r := gin.New()
	// 使用 gin-zap 替换 Gin 内置的日志功能
	r.Use(ginzap.Ginzap(logger, "2006-01-02T15:04:05.000Z0700", true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(LoggerMiddleware(logger))

	// 初始化WebSocket连接管理器
	wsManager := websocket.NewConnectionManager(logger)
	ctx := context.Background()
	wsManager.Start(ctx)

	// 初始化WebSocket事件处理器
	wsHandlers := websocket.InitializeHandlers(wsManager, logger)

	// 启动灵力增长后台任务
	spiritManager := spirit.NewSpiritGrowManager(logger, wsHandlers)
	spiritManager.Start()
	defer spiritManager.Stop()

	// ✅ 删除了译简化后的HeartbeatMonitor——玩家心跳是通过前端主动调用heartbeat API来维护的

	// 注册路由
	router.RegisterRoutes(r)

	// 注册WebSocket路由
	websocket.RegisterWebSocketRoutes(r, wsManager, logger)

	// 将WebSocket处理器注入到上下文中，供其他接口使用
	r.Use(func(c *gin.Context) {
		c.Set("ws_manager", wsManager)
		c.Set("ws_handlers", wsHandlers)
		c.Next()
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	log.Printf("Go server is running on %s", addr)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to run server: %v", err)
	}
}
