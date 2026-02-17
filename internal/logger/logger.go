package logger

import (
	"log/slog"
	"os"
)

// Log is the package-level structured logger used throughout monigo.
// Defaults to slog.Default(). Override via SetLogger() or Init().
var Log *slog.Logger = slog.Default()

// Init creates a new text handler logger at the given level.
func Init(level slog.Level) {
	Log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

// SetLogger replaces the package logger with a user-provided slog.Logger.
func SetLogger(l *slog.Logger) {
	if l != nil {
		Log = l
	}
}
