package applog

import (
	"context"
	"log/slog"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/runctx"
)

func FromContext(ctx context.Context) *slog.Logger {
	l := L()
	if runID, stepID, stepType, ok := runctx.From(ctx); ok {
		if runID != "" {
			l = l.With("run_id", runID)
		}
		if stepType != "" {
			l = l.With("step", stepType)
		}
		if stepID != "" {
			l = l.With("step_id", stepID)
		}
	}
	return l
}

func Truncate(s string, max int) string {
	if max <= 0 {
		max = 200
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
