package config

import (
	"log/slog"
	"os"
	"sync"
)

var once sync.Once

func InitLogger(level string) {
	once.Do(func() {
		var slogLevel slog.Level

		switch level {
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

		opts := &slog.HandlerOptions{
			Level: slogLevel,
		}

		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	})
}
