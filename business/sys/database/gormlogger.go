// Package database provides support for access the database.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"

	glogger "gorm.io/gorm/logger"

	"github.com/rmsj/service/foundation/web"
)

// to be able to use the centralized logger with gorm, had to implement it's own interface.
type logger struct {
	zap                       *zap.SugaredLogger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	LogLevel                  glogger.LogLevel
}

func newLogger(log *zap.SugaredLogger) *logger {
	return &logger{zap: log}
}

func (l *logger) LogMode(level glogger.LogLevel) glogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Info {
		l.zap.Infow(msg, "traceid", web.GetTraceID(ctx), "info", append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Warn {
		l.zap.Warnw(msg, "traceid", web.GetTraceID(ctx), "warn", append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= glogger.Error {
		l.zap.Errorw(msg, "traceid", web.GetTraceID(ctx), "error", append([]interface{}{utils.FileWithLineNum()}, data...))
	}
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= glogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.zap.Errorw("record not found", "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql)
		} else {
			l.zap.Errorw("error", "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "rows", rows, "sql", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= glogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.zap.Warnw(slowLog, "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql)
		} else {
			l.zap.Warnw(slowLog, "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "rows", rows, "sql", sql)
		}
	case l.LogLevel == glogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.zap.Infow("info", "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql)
		} else {
			l.zap.Infow("info", "traceid", web.GetTraceID(ctx), "file", utils.FileWithLineNum(), "error", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql, "rows", rows)
		}
	}
}
