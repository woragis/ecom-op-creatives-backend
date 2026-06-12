package video

import (
	"context"
	"fmt"
)

func NewLuma(apiKey, baseURL string) Provider {
	if baseURL == "" {
		baseURL = "https://api.lumalabs.ai"
	}
	return newHTTP("luma", baseURL, apiKey, lumaSubmit, lumaPoll)
}

func lumaSubmit(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error) {
	body := map[string]any{
		"prompt":       req.Prompt,
		"aspect_ratio": req.AspectRatio,
	}
	if req.Mode == ModeImage2Video && req.ImageURL != "" {
		body["keyframes"] = map[string]any{
			"frame0": map[string]any{
				"type": "image",
				"url":  req.ImageURL,
			},
		}
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := p.postJSON(ctx, "/dream-machine/v1/generations", body, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, fmt.Errorf("luma: empty generation id")
	}
	return &Job{ID: resp.ID}, nil
}

func lumaPoll(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error) {
	var resp struct {
		State string `json:"state"`
		Assets struct {
			Video string `json:"video"`
		} `json:"assets"`
		FailureReason string `json:"failure_reason"`
	}
	if err := p.getJSON(ctx, "/dream-machine/v1/generations/"+jobID, &resp); err != nil {
		return nil, err
	}
	switch resp.State {
	case "completed":
		if resp.Assets.Video == "" {
			return nil, fmt.Errorf("luma: no video url")
		}
		return &JobResult{Status: StatusCompleted, VideoURL: resp.Assets.Video}, nil
	case "failed":
		return &JobResult{Status: StatusFailed, Error: resp.FailureReason}, nil
	default:
		return &JobResult{Status: StatusRunning}, nil
	}
}
