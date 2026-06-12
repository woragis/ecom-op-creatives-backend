package video

import "context"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type SceneRequest struct {
	SceneID     string
	Prompt      string
	DurationSec int
	AspectRatio string
}

type Job struct {
	ID string
}

type JobResult struct {
	Status   Status
	VideoURL string
	Error    string
}

type Provider interface {
	ID() string
	Submit(ctx context.Context, req SceneRequest) (*Job, error)
	Poll(ctx context.Context, jobID string) (*JobResult, error)
}

type Clip struct {
	SceneID   string `json:"sceneId"`
	PublicURL string `json:"publicUrl"`
	FilePath  string `json:"filePath"`
	JobID     string `json:"jobId"`
	Provider  string `json:"provider"`
}

type StepOutput struct {
	Provider string `json:"provider"`
	Skipped  bool   `json:"skipped"`
	Reason   string `json:"reason,omitempty"`
	Clips    []Clip `json:"clips"`
}
