package database

import (
	"context"
	"fmt"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"skeleton-go/pkg/logger"
)

type gormLogger struct {
	log        *logger.Logger
	logQueries bool
	level      gormlogger.LogLevel
}

func newGormLogger(log *logger.Logger, logQueries bool) gormlogger.Interface {
	return &gormLogger{log: log, logQueries: logQueries, level: gormlogger.Warn}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cloned := *l
	cloned.level = level
	return &cloned
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(gormlogger.Info) {
		return
	}
	l.logDB(ctx, "gorm.info", msg, data...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(gormlogger.Warn) {
		return
	}
	l.logDB(ctx, "gorm.warn", msg, data...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if !l.shouldLog(gormlogger.Error) {
		return
	}
	l.logDB(ctx, "gorm.error", msg, data...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.log == nil {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := logger.Fields{
		"query":       sql,
		"duration_ms": float64(elapsed.Microseconds()) / 1000.0,
	}
	if rows != -1 {
		fields["rows"] = rows
	}

	switch {
	case err != nil && l.shouldLog(gormlogger.Error):
		l.log.Error(ctx, "gorm.trace", err, fields)
	case l.logQueries && l.shouldLog(gormlogger.Info):
		l.log.DB(ctx, "gorm.query", fields)
	}
}

func (l *gormLogger) shouldLog(level gormlogger.LogLevel) bool {
	return l.log != nil && l.level >= level
}

func (l *gormLogger) logDB(ctx context.Context, key, msg string, data ...interface{}) {
	if l.log == nil {
		return
	}
	l.log.DB(ctx, key, logger.Fields{"message": fmt.Sprintf(msg, data...)})
}
