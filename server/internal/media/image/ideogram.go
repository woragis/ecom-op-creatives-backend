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

type Ideogram struct {
	apiKey string
	hc     *http.Client
}

func NewIdeogram(apiKey string) *Ideogram {
	return &Ideogram{apiKey: apiKey, hc: &http.Client{Timeout: 120 * time.Second}}
}

func (i *Ideogram) ID() string { return "ideogram" }

func (i *Ideogram) Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error) {
	body, err := json.Marshal(map[string]any{
		"prompt":          req.Prompt,
		"aspect_ratio":    "9x16",
		"rendering_speed": "TURBO",
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.ideogram.ai/v1/ideogram-v3/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Api-Key", i.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := i.hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("ideogram http %d: %s", res.StatusCode, string(raw))
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
		return nil, fmt.Errorf("ideogram: empty image url")
	}
	data, err := Download(ctx, parsed.Data[0].URL)
	if err != nil {
		return nil, err
	}
	return &GenerateResult{Data: data}, nil
}
