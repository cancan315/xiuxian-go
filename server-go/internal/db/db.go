package db

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接（无日志版本，向后兼容）
func Init() error {
	return InitWithLogger(nil)
}

// InitWithLogger 初始化数据库连接，支持传入 zap logger
// 如果传入 logger，将启用详细的 SQL 日志记录
func InitWithLogger(zapLogger *zap.Logger) error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "xiuxian_db"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "xiuxian_user"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "xiuxian_password"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)

	// 配置 GORM
	config := &gorm.Config{}

	// 如果提供了 zap logger，使用自定义的 GORM logger
	if zapLogger != nil {
		// 根据环境变量设置日志级别
		logLevel := gormlogger.Info // 默认记录所有 SQL
		if os.Getenv("DB_LOG_LEVEL") == "warn" {
			logLevel = gormlogger.Warn
		} else if os.Getenv("DB_LOG_LEVEL") == "error" {
			logLevel = gormlogger.Error
		} else if os.Getenv("DB_LOG_LEVEL") == "silent" {
			logLevel = gormlogger.Silent
		}

		gormLogger := NewZapGormLogger(zapLogger).SetLogLevel(logLevel)
		config.Logger = gormLogger

		zapLogger.Info("[DB] 数据库日志已启用",
			zap.String("host", host),
			zap.String("port", port),
			zap.String("database", name),
			zap.Int("logLevel", int(logLevel)))
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return err
	}

	return nil
}
