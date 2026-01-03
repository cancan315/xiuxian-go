package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ZapGormLogger 自定义 GORM 日志器，使用 zap 作为底层实现
type ZapGormLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// NewZapGormLogger 创建 ZapGormLogger 实例
func NewZapGormLogger(zapLogger *zap.Logger) *ZapGormLogger {
	return &ZapGormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlogger.Info, // 默认 Info 级别，记录所有 SQL
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}
}

// LogMode 设置日志级别
func (l *ZapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 记录 Info 级别日志
func (l *ZapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.ZapLogger.Info(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Warn 记录 Warn 级别日志
func (l *ZapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Warn(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Error 记录 Error 级别日志
func (l *ZapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.ZapLogger.Error(fmt.Sprintf(msg, data...), zap.String("module", "gorm"))
	}
}

// Trace 记录 SQL 执行详情
func (l *ZapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建基础日志字段
	fields := []zap.Field{
		zap.String("module", "gorm"),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		// 忽略记录未找到错误（常见的查询场景）
		if l.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound) {
			l.ZapLogger.Debug("[DB] 记录未找到", fields...)
		} else {
			fields = append(fields, zap.Error(err))
			l.ZapLogger.Error("[DB] SQL执行错误", fields...)
		}

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		// 慢查询警告
		fields = append(fields, zap.String("slow_log", fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)))
		l.ZapLogger.Warn("[DB] 慢查询警告", fields...)

	case l.LogLevel >= gormlogger.Info:
		// 正常 SQL 记录
		l.ZapLogger.Debug("[DB] SQL执行", fields...)
	}
}

// SetSlowThreshold 设置慢查询阈值
func (l *ZapGormLogger) SetSlowThreshold(threshold time.Duration) *ZapGormLogger {
	l.SlowThreshold = threshold
	return l
}

// SetLogLevel 设置日志级别
func (l *ZapGormLogger) SetLogLevel(level gormlogger.LogLevel) *ZapGormLogger {
	l.LogLevel = level
	return l
}
