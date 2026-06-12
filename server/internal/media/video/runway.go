package video

import (
	"context"
	"fmt"
)

func NewRunway(apiKey, baseURL string) Provider {
	if baseURL == "" {
		baseURL = "https://api.dev.runwayml.com"
	}
	return newHTTP("runway", baseURL, apiKey, runwaySubmit, runwayPoll)
}

func runwaySubmit(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error) {
	ratio := "720:1280"
	if req.AspectRatio == "9:16" {
		ratio = "720:1280"
	}
	body := map[string]any{
		"promptText": req.Prompt,
		"model":      "gen4_turbo",
		"ratio":      ratio,
		"duration":   req.DurationSec,
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := p.postJSON(ctx, "/v1/text_to_video", body, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, fmt.Errorf("runway: empty task id")
	}
	return &Job{ID: resp.ID}, nil
}

func runwayPoll(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error) {
	var resp struct {
		Status string `json:"status"`
		Output []struct {
			URL string `json:"url"`
		} `json:"output"`
		Failure string `json:"failure"`
	}
	if err := p.getJSON(ctx, "/v1/tasks/"+jobID, &resp); err != nil {
		return nil, err
	}
	switch resp.Status {
	case "SUCCEEDED", "succeeded", "completed":
		if len(resp.Output) == 0 || resp.Output[0].URL == "" {
			return nil, fmt.Errorf("runway: no output url")
		}
		return &JobResult{Status: StatusCompleted, VideoURL: resp.Output[0].URL}, nil
	case "FAILED", "failed":
		return &JobResult{Status: StatusFailed, Error: resp.Failure}, nil
	default:
		return &JobResult{Status: StatusRunning}, nil
	}
}
