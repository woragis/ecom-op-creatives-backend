package image

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Flux struct {
	apiKey  string
	baseURL string
	hc      *http.Client
}

func NewFlux(apiKey, baseURL string) *Flux {
	if baseURL == "" {
		baseURL = "https://api.bfl.ml"
	}
	return &Flux{apiKey: apiKey, baseURL: baseURL, hc: &http.Client{Timeout: 60 * time.Second}}
}

func (f *Flux) ID() string { return "flux" }

func (f *Flux) Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": req.Prompt,
		"width":  768,
		"height": 1344,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.baseURL+"/v1/flux-pro-1.1", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-Key", f.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := f.hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("flux http %d: %s", res.StatusCode, string(raw))
	}
	var submit struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &submit); err != nil {
		return nil, err
	}
	if submit.ID == "" {
		return nil, fmt.Errorf("flux: empty task id")
	}

	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		pollReq, err := http.NewRequestWithContext(ctx, http.MethodGet, f.baseURL+"/v1/get_result?id="+submit.ID, nil)
		if err != nil {
			return nil, err
		}
		pollReq.Header.Set("X-Key", f.apiKey)
		pollRes, err := f.hc.Do(pollReq)
		if err != nil {
			return nil, err
		}
		pollRaw, err := io.ReadAll(pollRes.Body)
		pollRes.Body.Close()
		if err != nil {
			return nil, err
		}
		if pollRes.StatusCode >= 400 {
			return nil, fmt.Errorf("flux poll http %d: %s", pollRes.StatusCode, string(pollRaw))
		}
		var result struct {
			Status string `json:"status"`
			Result struct {
				Sample string `json:"sample"`
			} `json:"result"`
		}
		if err := json.Unmarshal(pollRaw, &result); err != nil {
			return nil, err
		}
		switch result.Status {
		case "Ready":
			if result.Result.Sample == "" {
				return nil, fmt.Errorf("flux: empty sample url")
			}
			data, err := Download(ctx, result.Result.Sample)
			if err != nil {
				return nil, err
			}
			return &GenerateResult{Data: data}, nil
		case "Error", "Failed":
			return nil, fmt.Errorf("flux generation failed")
		default:
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}
	return nil, fmt.Errorf("flux: timed out")
}
