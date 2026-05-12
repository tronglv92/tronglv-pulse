package gorm

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/logger"
	"time"
)

func GetLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Error
	}
}

type SqlLogger struct {
	logLevel             logger.LogLevel
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
}

func NewSqlLogger(level logger.LogLevel, slowThreshold time.Duration, ignoreRecordNotFound bool) *SqlLogger {
	return &SqlLogger{
		logLevel:             level,
		slowThreshold:        slowThreshold,
		ignoreRecordNotFound: ignoreRecordNotFound,
	}
}

func (l *SqlLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *SqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		logx.WithContext(ctx).Infow(fmt.Sprintf(msg, data...))
	}
}

func (l *SqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		logx.WithContext(ctx).Infow(fmt.Sprintf(msg, data...), logx.Field("level", "warn"))
	}
}

func (l *SqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		logx.WithContext(ctx).Errorw(fmt.Sprintf(msg, data...), logx.Field("level", "error"))
	}
}

func (l *SqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []logx.LogField{
		logx.Field("sql", sql),
		logx.Field("duration_ms", float64(elapsed.Nanoseconds())/1e6),
		logx.Field("rows_affected", rows),
	}

	switch {
	case err != nil && (!l.ignoreRecordNotFound || err.Error() != "record not found"):
		if l.logLevel >= logger.Error {
			fields = append(fields, logx.Field("error", err.Error()))
			logx.WithContext(ctx).Errorw("SQL execution error", fields...)
		}
	case elapsed >= l.slowThreshold && l.logLevel >= logger.Warn:
		fields = append(fields, logx.Field("threshold_ms", float64(l.slowThreshold.Nanoseconds())/1e6))
		logx.WithContext(ctx).Sloww("Slow SQL query", fields...)
	case l.logLevel >= logger.Info:
		logx.WithContext(ctx).Infow("SQL query", fields...)
	}
}
