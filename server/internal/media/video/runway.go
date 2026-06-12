package video

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type runwayConfig struct {
	textModel  string
	imageModel string
}

func NewRunway(apiKey, baseURL string, cfg runwayConfig) Provider {
	if baseURL == "" {
		baseURL = "https://api.dev.runwayml.com"
	}
	if cfg.textModel == "" {
		cfg.textModel = "gen4.5"
	}
	if cfg.imageModel == "" {
		cfg.imageModel = "gen4_turbo"
	}
	opts := cfg
	return newHTTP("runway", baseURL, apiKey,
		func(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error) {
			return runwaySubmit(ctx, p, req, opts)
		},
		runwayPoll,
	)
}

func runwayRatio(aspectRatio string) string {
	if aspectRatio == "9:16" || aspectRatio == "3:4" {
		return "720:1280"
	}
	return "1280:720"
}

func runwaySubmit(ctx context.Context, p *HTTPProvider, req SceneRequest, cfg runwayConfig) (*Job, error) {
	ratio := runwayRatio(req.AspectRatio)
	duration := clampDuration(req.DurationSec)

	if req.Mode == ModeImage2Video && req.ImageURL != "" {
		body := map[string]any{
			"promptImage": []map[string]string{
				{"uri": req.ImageURL, "position": "first"},
			},
			"promptText": req.Prompt,
			"model":      cfg.imageModel,
			"ratio":      ratio,
			"duration":   duration,
		}
		return runwayPostTask(ctx, p, "/v1/image_to_video", body)
	}

	body := map[string]any{
		"promptText": req.Prompt,
		"model":      cfg.textModel,
		"ratio":      ratio,
		"duration":   duration,
	}
	return runwayPostTask(ctx, p, "/v1/text_to_video", body)
}

func runwayPostTask(ctx context.Context, p *HTTPProvider, path string, body map[string]any) (*Job, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := p.postJSON(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, fmt.Errorf("runway: empty task id")
	}
	return &Job{ID: resp.ID}, nil
}

func runwayPoll(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error) {
	var resp struct {
		Status  string          `json:"status"`
		Output  json.RawMessage `json:"output"`
		Failure string          `json:"failure"`
	}
	if err := p.getJSON(ctx, "/v1/tasks/"+jobID, &resp); err != nil {
		return nil, err
	}
	switch strings.ToUpper(resp.Status) {
	case "SUCCEEDED", "COMPLETED":
		videoURL, err := parseRunwayOutputURLs(resp.Output)
		if err != nil {
			return nil, err
		}
		if videoURL == "" {
			return nil, fmt.Errorf("runway: no output url")
		}
		return &JobResult{Status: StatusCompleted, VideoURL: videoURL}, nil
	case "FAILED":
		return &JobResult{Status: StatusFailed, Error: resp.Failure}, nil
	default:
		return &JobResult{Status: StatusRunning}, nil
	}
}

func parseRunwayOutputURLs(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var urls []string
	if err := json.Unmarshal(raw, &urls); err == nil {
		if len(urls) > 0 {
			return urls[0], nil
		}
		return "", nil
	}
	var objects []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &objects); err != nil {
		return "", fmt.Errorf("runway: parse output: %w", err)
	}
	if len(objects) > 0 {
		return objects[0].URL, nil
	}
	return "", nil
}
