package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"skeleton-go/pkg/session"
)

// Fields attaches extra metadata to a log entry.
type Fields map[string]interface{}

// Logger writes structured JSON entries, including session metadata when present.
type Logger struct {
	writers []io.Writer
	mu      sync.Mutex
}

// Options configure how the logger writes to disk in addition to stdout.
type Options struct {
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

// New constructs a logger that writes to stdout and, if configured, a rotating file.
func New(opts Options) (*Logger, error) {
	writers := []io.Writer{os.Stdout}
	if opts.File != "" {
		if err := os.MkdirAll(filepath.Dir(opts.File), 0o755); err != nil {
			return nil, err
		}
		writers = append(writers, &lumberjack.Logger{
			Filename:   opts.File,
			MaxSize:    opts.MaxSizeMB,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAgeDays,
			Compress:   opts.Compress,
		})
	}

	return &Logger{writers: writers}, nil
}

// Info writes an informational log entry.
func (l *Logger) Info(ctx context.Context, message string, fields Fields) {
	l.write(ctx, "info", message, fields)
}

// Error writes an error log entry and captures the error text inside the fields map.
func (l *Logger) Error(ctx context.Context, message string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	l.write(ctx, "error", message, fields)
}

// DB records database-layer interactions, keeping them distinct for log filters.
func (l *Logger) DB(ctx context.Context, message string, fields Fields) {
	l.write(ctx, "db", message, fields)
}

func (l *Logger) write(ctx context.Context, level, message string, fields Fields) {
	if l == nil {
		return
	}

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level,
		"message":   message,
	}

	if sess := session.FromContext(ctx); sess != nil {
		entry["session_id"] = sess.ID
	}

	if len(fields) > 0 {
		entry["fields"] = fields
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, writer := range l.writers {
		writer.Write(append(payload, '\n'))
	}
}
