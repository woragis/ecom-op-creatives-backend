package runctx

import "context"

type values struct {
	RunID    string
	StepID   string
	StepType string
}

type key struct{}

func With(ctx context.Context, runID, stepID, stepType string) context.Context {
	return context.WithValue(ctx, key{}, values{
		RunID:    runID,
		StepID:   stepID,
		StepType: stepType,
	})
}

func From(ctx context.Context) (runID, stepID, stepType string, ok bool) {
	v, ok := ctx.Value(key{}).(values)
	if !ok {
		return "", "", "", false
	}
	return v.RunID, v.StepID, v.StepType, true
}
