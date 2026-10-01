package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Init(level string) *slog.Logger {
	var programLevel = new(slog.LevelVar)

	switch strings.ToLower(level) {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
