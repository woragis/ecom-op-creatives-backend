package applog

import (
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

var root *slog.Logger

func Init() {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOG_FORMAT")), "text") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	root = slog.New(handler)
}

func L() *slog.Logger {
	if root == nil {
		Init()
	}
	return root
}

func ForRun(runID uuid.UUID) *slog.Logger {
	return L().With("run_id", runID.String())
}

func ForStep(runID uuid.UUID, stepType string) *slog.Logger {
	return ForRun(runID).With("step", stepType)
}
