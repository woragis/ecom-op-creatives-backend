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

type Dalle struct {
	apiKey string
	hc     *http.Client
}

func NewDalle(apiKey string) *Dalle {
	return &Dalle{apiKey: apiKey, hc: &http.Client{Timeout: 120 * time.Second}}
}

func (d *Dalle) ID() string { return "dalle" }

func (d *Dalle) Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error) {
	size := "1024x1792"
	if req.AspectRatio == "1:1" {
		size = "1024x1024"
	}
	body, err := json.Marshal(map[string]any{
		"model":   "dall-e-3",
		"prompt":  req.Prompt,
		"size":    size,
		"n":       1,
		"quality": "standard",
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := d.hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("dalle http %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 || parsed.Data[0].URL == "" {
		return nil, fmt.Errorf("dalle: empty image url")
	}
	data, err := Download(ctx, parsed.Data[0].URL)
	if err != nil {
		return nil, err
	}
	return &GenerateResult{Data: data}, nil
}
