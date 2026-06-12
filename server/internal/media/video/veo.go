package video

import (
	"context"
	"fmt"
)

func NewVeo(apiKey, baseURL string) Provider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	return newHTTP("veo", baseURL, apiKey, veoSubmit, veoPoll)
}

func veoSubmit(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error) {
	instance := map[string]any{"prompt": req.Prompt}
	if req.Mode == ModeImage2Video && req.ImageURL != "" {
		instance["image"] = map[string]any{"uri": req.ImageURL}
	}
	body := map[string]any{
		"instances": []map[string]any{instance},
		"parameters": map[string]any{
			"aspectRatio": req.AspectRatio,
			"durationSeconds": req.DurationSec,
		},
	}
	var resp struct {
		Name string `json:"name"`
	}
	path := "/v1beta/models/veo-2.0-generate-001:predictLongRunning"
	if err := p.postJSON(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	if resp.Name == "" {
		return nil, fmt.Errorf("veo: empty operation name")
	}
	return &Job{ID: resp.Name}, nil
}

func veoPoll(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error) {
	var resp struct {
		Done     bool `json:"done"`
		Error    struct {
			Message string `json:"message"`
		} `json:"error"`
		Response struct {
			GenerateVideoResponse struct {
				GeneratedSamples []struct {
					Video struct {
						URI string `json:"uri"`
					} `json:"video"`
				} `json:"generatedSamples"`
			} `json:"generateVideoResponse"`
		} `json:"response"`
	}
	path := "/v1beta/" + jobID
	if err := p.getJSON(ctx, path, &resp); err != nil {
		return nil, err
	}
	if resp.Error.Message != "" {
		return &JobResult{Status: StatusFailed, Error: resp.Error.Message}, nil
	}
	if !resp.Done {
		return &JobResult{Status: StatusRunning}, nil
	}
	samples := resp.Response.GenerateVideoResponse.GeneratedSamples
	if len(samples) == 0 || samples[0].Video.URI == "" {
		return nil, fmt.Errorf("veo: no video uri")
	}
	return &JobResult{Status: StatusCompleted, VideoURL: samples[0].Video.URI}, nil
}
