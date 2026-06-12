package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
)

type Dalle struct {
	apiKey string
	model  string
	hc     *http.Client
}

func NewDalle(apiKey, model string) *Dalle {
	if strings.TrimSpace(model) == "" {
		model = "gpt-image-1-mini"
	}
	return &Dalle{apiKey: apiKey, model: model, hc: &http.Client{Timeout: 120 * time.Second}}
}

func (d *Dalle) ID() string { return "dalle" }

func (d *Dalle) Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error) {
	started := time.Now()
	log := applog.FromContext(ctx).With("service", "openai", "operation", "images.generations", "provider", "dalle", "model", d.model, "scene_id", req.SceneID, "role", req.Role)
	log.Info("openai image request",
		"prompt_preview", applog.Truncate(req.Prompt, 200),
		"aspect_ratio", req.AspectRatio,
	)

	body, err := json.Marshal(d.requestBody(req))
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
		log.Error("openai image failed",
			"status", res.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
			"body_preview", applog.Truncate(string(raw), 300),
		)
		return nil, fmt.Errorf("openai image http %d: %s", res.StatusCode, string(raw))
	}
	data, err := decodeImageResponse(ctx, raw)
	if err != nil {
		return nil, err
	}
	log.Info("openai image completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"bytes", len(data),
	)
	return &GenerateResult{Data: data}, nil
}

func (d *Dalle) requestBody(req ImageRequest) map[string]any {
	size := openAIImageSize(d.model, req.AspectRatio)
	if strings.HasPrefix(d.model, "dall-e") {
		return map[string]any{
			"model":   d.model,
			"prompt":  req.Prompt,
			"size":    size,
			"n":       1,
			"quality": "standard",
		}
	}
	return map[string]any{
		"model":   d.model,
		"prompt":  req.Prompt,
		"size":    size,
		"n":       1,
		"quality": "medium",
	}
}

func openAIImageSize(model, aspectRatio string) string {
	if strings.HasPrefix(model, "dall-e") {
		if aspectRatio == "1:1" {
			return "1024x1024"
		}
		return "1024x1792"
	}
	if aspectRatio == "1:1" {
		return "1024x1024"
	}
	return "1024x1536"
}

func decodeImageResponse(ctx context.Context, raw []byte) ([]byte, error) {
	var parsed struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("openai image: empty response")
	}
	item := parsed.Data[0]
	if item.B64JSON != "" {
		data, err := base64.StdEncoding.DecodeString(item.B64JSON)
		if err != nil {
			return nil, fmt.Errorf("openai image: decode b64: %w", err)
		}
		return data, nil
	}
	if item.URL != "" {
		return Download(ctx, item.URL)
	}
	return nil, fmt.Errorf("openai image: no url or b64_json in response")
}
