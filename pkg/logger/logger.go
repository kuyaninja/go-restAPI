package logger

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"skeleton-go/pkg/session"
)

// Fields attaches extra metadata to a log entry.
type Fields map[string]interface{}

// Logger wraps a zap.Logger with convenience helpers.
type Logger struct {
	base *zap.Logger
}

// Options configure how the logger writes to disk in addition to stdout.
type Options struct {
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

const dbLogLevel zapcore.Level = zapcore.FatalLevel + 1

// New constructs a zap-backed logger that writes to stdout and, if configured, a rotating file.
func New(opts Options) (*Logger, error) {
	encoder := zapcore.NewJSONEncoder(encoderConfig())

	cores := []zapcore.Core{zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.LevelEnablerFunc(allLevelsEnabled))}

	if opts.File != "" {
		if err := os.MkdirAll(filepath.Dir(opts.File), 0o755); err != nil {
			return nil, err
		}
		lumberjackWriter := &lumberjack.Logger{
			Filename:   opts.File,
			MaxSize:    opts.MaxSizeMB,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAgeDays,
			Compress:   opts.Compress,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(lumberjackWriter), zap.LevelEnablerFunc(allLevelsEnabled)))
	}

	base := zap.New(zapcore.NewTee(cores...))
	return &Logger{base: base}, nil
}

// Info writes an informational log entry.
func (l *Logger) Info(ctx context.Context, message string, fields Fields) {
	l.write(ctx, zapcore.InfoLevel, message, fields)
}

// Error writes an error log entry and captures the error text inside the fields map.
func (l *Logger) Error(ctx context.Context, message string, err error, fields Fields) {
	if err != nil {
		if fields == nil {
			fields = Fields{}
		}
		fields["error"] = err.Error()
	}
	l.write(ctx, zapcore.ErrorLevel, message, fields)
}

// DB records database-layer interactions, keeping them distinct for log filters.
func (l *Logger) DB(ctx context.Context, message string, fields Fields) {
	l.write(ctx, dbLogLevel, message, fields)
}

func (l *Logger) write(ctx context.Context, level zapcore.Level, message string, fields Fields) {
	if l == nil || l.base == nil {
		return
	}

	ce := l.base.Check(level, message)
	if ce == nil {
		return
	}

	zapFields := make([]zap.Field, 0, len(fields)+2)
	if sess := session.FromContext(ctx); sess != nil {
		zapFields = append(zapFields, zap.String("session_id", sess.ID))
	}
	if len(fields) > 0 {
		zapFields = append(zapFields, zap.Any("fields", fields))
	}
	ce.Write(zapFields...)
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "message",
		CallerKey:      "",
		StacktraceKey:  "",
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeLevel:    encodeLevel,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
}

func encodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == dbLogLevel {
		enc.AppendString("db")
		return
	}
	zapcore.LowercaseLevelEncoder(level, enc)
}

func allLevelsEnabled(zapcore.Level) bool {
	return true
}

// Sync flushes buffered logs.
func (l *Logger) Sync() error {
	if l == nil || l.base == nil {
		return nil
	}
	return l.base.Sync()
}

// NewTestLogger exposes a zap logger backed by the provided writer. Used in tests.
func newTestLogger(writer io.Writer) *Logger {
	encoder := zapcore.NewJSONEncoder(encoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zap.LevelEnablerFunc(allLevelsEnabled))
	return &Logger{base: zap.New(core)}
}

var errNilLogger = errors.New("logger not initialized")

// Base exposes the underlying zap logger for advanced use.
func (l *Logger) Base() (*zap.Logger, error) {
	if l == nil || l.base == nil {
		return nil, errNilLogger
	}
	return l.base, nil
}
