package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
)

type HTTPProvider struct {
	id      string
	baseURL string
	apiKey  string
	hc      *http.Client
	submit  func(ctx context.Context, p *HTTPProvider, req SceneRequest) (*Job, error)
	poll    func(ctx context.Context, p *HTTPProvider, jobID string) (*JobResult, error)
}

func newHTTP(id, baseURL, apiKey string,
	submit func(context.Context, *HTTPProvider, SceneRequest) (*Job, error),
	poll func(context.Context, *HTTPProvider, string) (*JobResult, error),
) *HTTPProvider {
	return &HTTPProvider{
		id: id, baseURL: baseURL, apiKey: apiKey,
		hc: &http.Client{Timeout: 60 * time.Second},
		submit: submit, poll: poll,
	}
}

func (p *HTTPProvider) ID() string { return p.id }

func (p *HTTPProvider) Submit(ctx context.Context, req SceneRequest) (*Job, error) {
	return p.submit(ctx, p, req)
}

func (p *HTTPProvider) Poll(ctx context.Context, jobID string) (*JobResult, error) {
	return p.poll(ctx, p, jobID)
}

func (p *HTTPProvider) postJSON(ctx context.Context, path string, body any, out any) error {
	started := time.Now()
	log := applog.FromContext(ctx).With("service", p.id, "operation", "http.post", "path", path)
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	p.setHeaders(req)
	res, err := p.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		log.Error("video provider post failed",
			"status", res.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
			"body_preview", applog.Truncate(string(raw), 300),
		)
		return fmt.Errorf("%s post %s: %d %s", p.id, path, res.StatusCode, string(raw))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return err
		}
	}
	log.Debug("video provider post ok", "duration_ms", time.Since(started).Milliseconds())
	return nil
}

func (p *HTTPProvider) getJSON(ctx context.Context, path string, out any) error {
	started := time.Now()
	log := applog.FromContext(ctx).With("service", p.id, "operation", "http.get", "path", path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
	if err != nil {
		return err
	}
	p.setHeaders(req)
	res, err := p.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		log.Error("video provider get failed",
			"status", res.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
			"body_preview", applog.Truncate(string(raw), 300),
		)
		return fmt.Errorf("%s get %s: %d %s", p.id, path, res.StatusCode, string(raw))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return err
	}
	log.Debug("video provider get ok", "duration_ms", time.Since(started).Milliseconds())
	return nil
}

func (p *HTTPProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	if p.id == "runway" {
		req.Header.Set("X-Runway-Version", "2024-11-06")
	}
}
