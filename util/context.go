package util

import (
	"context"
	"log/slog"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l := ctx.Value(loggerKey); l == nil {
		slog.Warn("Using default slog as context does not have logger info")
		return slog.Default()
	} else {
		return l.(*slog.Logger)
	}
}
