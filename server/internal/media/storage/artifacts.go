package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
)

func (l *Local) ArtifactsDir(runID string) string {
	return filepath.Join(l.RunDir(runID), "artifacts")
}

// WriteStepArtifact mirrors pipeline step output JSON to disk for backup and review.
func (l *Local) WriteStepArtifact(runID, stepType string, output []byte) error {
	if len(output) == 0 {
		output = []byte(`{}`)
	}
	dir := l.ArtifactsDir(runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	order := pipeline.StepOrder(stepType)
	name := fmt.Sprintf("%02d-%s.json", order, stepType)
	return os.WriteFile(filepath.Join(dir, name), output, 0o644)
}

// WriteStepErrorArtifact records a failed step attempt on disk.
func (l *Local) WriteStepErrorArtifact(runID, stepType, message string) error {
	dir := l.ArtifactsDir(runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	order := pipeline.StepOrder(stepType)
	body, err := json.Marshal(map[string]any{
		"stepType": stepType,
		"failed":   true,
		"error":    message,
		"at":       time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%02d-%s.error.json", order, stepType)
	return os.WriteFile(filepath.Join(dir, name), body, 0o644)
}

// WriteRunMeta writes a summary file after each step completes.
func (l *Local) WriteRunMeta(runID string, meta map[string]any) error {
	dir := l.ArtifactsDir(runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "run-meta.json"), body, 0o644)
}
