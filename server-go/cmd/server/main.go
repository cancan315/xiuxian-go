package main

import (
	"log"
	"net/http"
	"os"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/http/handlers/online"
	"xiuxian/server-go/internal/http/router"
	"xiuxian/server-go/internal/redis"
	"xiuxian/server-go/internal/spirit"
	"xiuxian/server-go/internal/tasks"

	"github.com/gin-contrib/gzip"
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
	// 使用 gzip 中间件压缩响应数据
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	// 使用 gin-zap 替换 Gin 内置的日志功能
	r.Use(ginzap.Ginzap(logger, "2006-01-02T15:04:05.000Z0700", true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(LoggerMiddleware(logger))

	// 启动灵力增长后台任务
	spiritManager := spirit.NewSpiritGrowManager(logger)
	spiritManager.Start()
	defer spiritManager.Stop()

	// ✅ 启动心跳监控任务（清理超时玩家和战斗数据）
	online.StartHeartbeatMonitor(logger)

	// ✅ 启动后台定期同步任务（装备和灵宠资源）
	tasks.InitTasks(logger)

	// ✅ 将灵力增长管理器注入到上下文中（必须在注册路由之前）
	r.Use(func(c *gin.Context) {
		c.Set("spirit_manager", spiritManager)
		c.Next()
	})

	// 注册路由
	router.RegisterRoutes(r)

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
