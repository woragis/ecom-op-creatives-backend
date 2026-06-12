package video

import (
	"context"
	"fmt"
)

func NewKling(apiKey, baseURL string) Provider {
	if baseURL == "" {
		baseURL = "https://api.klingai.com"
	}
	return newHTTP("kling", baseURL, apiKey, klingSubmit, klingPoll)
}

func klingSubmit(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error) {
	body := map[string]any{
		"prompt":       req.Prompt,
		"aspect_ratio": req.AspectRatio,
		"duration":     fmt.Sprintf("%d", req.DurationSec),
	}
	var resp struct {
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
		TaskID string `json:"task_id"`
	}
	if err := p.postJSON(ctx, "/v1/videos/text2video", body, &resp); err != nil {
		return nil, err
	}
	id := resp.Data.TaskID
	if id == "" {
		id = resp.TaskID
	}
	if id == "" {
		return nil, fmt.Errorf("kling: empty task_id")
	}
	return &Job{ID: id}, nil
}

func klingPoll(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error) {
	var resp struct {
		Data struct {
			TaskStatus string `json:"task_status"`
			TaskResult struct {
				Videos []struct {
					URL string `json:"url"`
				} `json:"videos"`
			} `json:"task_result"`
			TaskStatusMsg string `json:"task_status_msg"`
		} `json:"data"`
	}
	if err := p.getJSON(ctx, "/v1/videos/text2video/"+jobID, &resp); err != nil {
		return nil, err
	}
	switch resp.Data.TaskStatus {
	case "succeed", "completed", "success":
		if len(resp.Data.TaskResult.Videos) == 0 {
			return nil, fmt.Errorf("kling: no video url")
		}
		return &JobResult{Status: StatusCompleted, VideoURL: resp.Data.TaskResult.Videos[0].URL}, nil
	case "failed", "error":
		return &JobResult{Status: StatusFailed, Error: resp.Data.TaskStatusMsg}, nil
	default:
		return &JobResult{Status: StatusRunning}, nil
	}
}
