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
	path := "/v1/videos/text2video"
	body := map[string]any{
		"prompt":       req.Prompt,
		"aspect_ratio": req.AspectRatio,
		"duration":     fmt.Sprintf("%d", req.DurationSec),
	}
	if req.Mode == ModeImage2Video && req.ImageURL != "" {
		path = "/v1/videos/image2video"
		body["image"] = req.ImageURL
	}
	var resp struct {
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
		TaskID string `json:"task_id"`
	}
	if err := p.postJSON(ctx, path, body, &resp); err != nil {
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
	err := p.getJSON(ctx, "/v1/videos/text2video/"+jobID, &resp)
	if err != nil {
		if err2 := p.getJSON(ctx, "/v1/videos/image2video/"+jobID, &resp); err2 != nil {
			return nil, err
		}
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
