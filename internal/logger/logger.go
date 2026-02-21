package logger

import (
	"log/slog"
	"os"
	"sync/atomic"
)

// loggerHolder stores the active logger atomically to avoid data races.
var loggerHolder atomic.Pointer[slog.Logger]

func init() {
	loggerHolder.Store(slog.Default())
}

// Log returns the package-level structured logger used throughout monigo.
var Log logAccessor

type logAccessor struct{}

// We expose Log as a struct so existing call sites (logger.Log.Info(...)) keep working
// via the __call__ pattern below. The actual pointer is read atomically each time.
func get() *slog.Logger { return loggerHolder.Load() }

func (logAccessor) Info(msg string, args ...any)  { get().Info(msg, args...) }
func (logAccessor) Warn(msg string, args ...any)  { get().Warn(msg, args...) }
func (logAccessor) Error(msg string, args ...any) { get().Error(msg, args...) }
func (logAccessor) Debug(msg string, args ...any) { get().Debug(msg, args...) }

// Init creates a new text handler logger at the given level.
func Init(level slog.Level) {
	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	loggerHolder.Store(l)
}

// SetLogger replaces the package logger with a user-provided slog.Logger.
func SetLogger(l *slog.Logger) {
	if l != nil {
		loggerHolder.Store(l)
	}
}

// Get returns the underlying *slog.Logger for callers that need it directly.
func Get() *slog.Logger {
	return get()
}
