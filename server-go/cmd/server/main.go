package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

// buildLoggerWithLoki 创建带 Loki HTTP 输出的 zap logger
func buildLoggerWithLoki(config zap.Config, lokiURL string) (*zap.Logger, error) {
	// 创建 Loki HTTP 编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建 Loki HTTP 核心
	lokiCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		&lokiHTTPWriter{url: lokiURL},
		config.Level,
	)

	// 结合标准输出和 Loki 输出
	stdoutCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		config.Level,
	)

	// 合并多个 core
	core := zapcore.NewTee(lokiCore, stdoutCore)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// lokiHTTPWriter 通过 HTTP 推送日志到 Loki
type lokiHTTPWriter struct {
	url    string
	buffer [][]byte
}

func (w *lokiHTTPWriter) Write(p []byte) (n int, err error) {
	// 复制数据到缓冲区，避免外部修改
	copyData := make([]byte, len(p))
	copy(copyData, p)
	w.buffer = append(w.buffer, copyData)

	// 达到一定数量后或超时推送
	if len(w.buffer) >= 10 {
		_ = w.flush()
	}
	return len(p), nil
}

func (w *lokiHTTPWriter) Sync() error {
	return w.flush()
}

func (w *lokiHTTPWriter) flush() error {
	if len(w.buffer) == 0 {
		return nil
	}

	// 构建 Loki LogQL 格式的请求
	var lines [][]interface{}
	for _, entry := range w.buffer {
		// 时间戳（纳秒）和日志内容
		nowNano := time.Now().UnixNano()
		lines = append(lines, []interface{}{fmt.Sprintf("%d", nowNano), string(entry)})
	}

	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"job":  "xiuxian-server",
					"host": "localhost",
				},
				"values": lines,
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		w.url+"/loki/api/v1/push",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		// 推送失败不影响主程序
		fmt.Fprintf(os.Stderr, "failed to push logs to Loki: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	w.buffer = nil
	return nil
}

func main() {
	_ = godotenv.Load()

	// ✅ 先初始化 zap 日志记录器，配置日志级别
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

	// ✅ 集成 Loki 日志输出
	lokiURL := os.Getenv("LOKI_URL")
	var logger *zap.Logger
	var err error

	if lokiURL != "" {
		logger, err = buildLoggerWithLoki(config, lokiURL)
		if err != nil {
			log.Printf("failed to init loki logger, fallback to stdout: %v", err)
			logger, _ = config.Build()
		}
	} else {
		logger, _ = config.Build()
	}
	defer logger.Sync()

	// ✅ 使用 logger 初始化数据库（启用详细 SQL 日志）
	if err := db.InitWithLogger(logger); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	if err := redis.Init(); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

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
