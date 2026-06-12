package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Stability struct {
	apiKey string
	hc     *http.Client
}

func NewStability(apiKey string) *Stability {
	return &Stability{apiKey: apiKey, hc: &http.Client{Timeout: 120 * time.Second}}
}

func (s *Stability) ID() string { return "stability" }

func (s *Stability) Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("prompt", req.Prompt)
	_ = w.WriteField("aspect_ratio", "9:16")
	_ = w.WriteField("output_format", "png")
	if err := w.Close(); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stability.ai/v2beta/stable-image/generate/sd3", &body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Accept", "image/*")
	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	res, err := s.hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("stability http %d: %s", res.StatusCode, string(data))
	}
	return &GenerateResult{Data: data}, nil
}
