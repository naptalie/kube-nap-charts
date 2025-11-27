// Package logger provides structured logging capabilities using Go's slog package.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// Level represents the level of logging.
type Level slog.Level

// Set of logging levels.
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// Logger represents a logger for logging information.
type Logger struct {
	handler   slog.Handler
	traceIDFn func(context.Context) string
}

// New constructs a new Logger.
func New(w io.Writer, minLevel Level, serviceName string, traceIDFn func(context.Context) string) *Logger {
	return NewWithHandler(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.Level(minLevel),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(time.Now().UTC().Format(time.RFC3339)),
				}
			}
			if a.Key == slog.LevelKey {
				return slog.Attr{
					Key:   "level",
					Value: a.Value,
				}
			}
			if a.Key == slog.MessageKey {
				return slog.Attr{
					Key:   "msg",
					Value: a.Value,
				}
			}
			return a
		},
	}), serviceName, traceIDFn)
}

// NewWithHandler constructs a new Logger with a custom handler.
func NewWithHandler(handler slog.Handler, serviceName string, traceIDFn func(context.Context) string) *Logger {
	// Add service name to all logs
	handler = handler.WithAttrs([]slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	})

	return &Logger{
		handler:   handler,
		traceIDFn: traceIDFn,
	}
}

// Debug logs at LevelDebug.
func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelDebug, 3, msg, args...)
}

// Info logs at LevelInfo.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelInfo, 3, msg, args...)
}

// Warn logs at LevelWarn.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelWarn, 3, msg, args...)
}

// Error logs at LevelError.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, slog.LevelError, 3, msg, args...)
}

func (log *Logger) write(ctx context.Context, level slog.Level, caller int, msg string, args ...any) {
	slogRec := slog.NewRecord(time.Now().UTC(), level, msg, uintptr(caller))

	// Add trace ID if available
	if log.traceIDFn != nil {
		if traceID := log.traceIDFn(ctx); traceID != "" {
			args = append(args, "trace_id", traceID)
		}
	}

	// Add source location
	if level >= slog.LevelError {
		_, file, line, ok := runtime.Caller(caller)
		if ok {
			args = append(args, "source", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	slogRec.Add(args...)

	log.handler.Handle(ctx, slogRec)
}

// NewStdLogger returns a standard library logger that writes to the Logger.
func NewStdLogger(log *Logger, level Level) *stdLogger {
	return &stdLogger{
		logger: log,
		level:  slog.Level(level),
	}
}

type stdLogger struct {
	logger *Logger
	level  slog.Level
}

func (l *stdLogger) Write(p []byte) (n int, err error) {
	ctx := context.Background()
	l.logger.write(ctx, l.level, 4, string(p))
	return len(p), nil
}
