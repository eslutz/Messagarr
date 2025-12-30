package logger

import (
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

// Init initializes the logger with default settings
func Init() {
	env := os.Getenv("ENV")
	var handler slog.Handler

	if env == "production" {
		// JSON handler for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Text handler for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// SetLevel updates the log level
func SetLevel(level string) {
	env := os.Getenv("ENV")
	var handler slog.Handler
	var slogLevel slog.Level

	// Parse log level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Override with LOG_LEVEL env var if set
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		switch strings.ToLower(envLevel) {
		case "debug":
			slogLevel = slog.LevelDebug
		case "info":
			slogLevel = slog.LevelInfo
		case "warn":
			slogLevel = slog.LevelWarn
		case "error":
			slogLevel = slog.LevelError
		}
	}

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		})
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}
