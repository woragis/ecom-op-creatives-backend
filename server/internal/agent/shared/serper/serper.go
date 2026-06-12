package serper

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

const serperURL = "https://google.serper.dev/search"

type HTTPClient struct {
	apiKey string
	hc     *http.Client
	gl     string
	hl     string
}

func NewHTTP(apiKey string) *HTTPClient {
	return &HTTPClient{
		apiKey: apiKey,
		hc:     &http.Client{Timeout: 15 * time.Second},
		gl:     "br",
		hl:     "pt-br",
	}
}

type serperRequest struct {
	Q   string `json:"q"`
	GL  string `json:"gl"`
	HL  string `json:"hl"`
	Num int    `json:"num"`
}

type serperResponse struct {
	Organic []Result `json:"organic"`
}

func (c *HTTPClient) Search(ctx context.Context, query string) (*SearchResponse, error) {
	started := time.Now()
	log := applog.FromContext(ctx).With("service", "serper", "operation", "search")
	log.Info("serper search", "query", query)

	body, err := json.Marshal(serperRequest{Q: query, GL: c.gl, HL: c.hl, Num: 5})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serperURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		log.Error("serper search failed",
			"status", res.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
			"body_preview", applog.Truncate(string(raw), 300),
		)
		return nil, fmt.Errorf("serper http %d: %s", res.StatusCode, string(raw))
	}

	var parsed serperResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	log.Info("serper search completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"results", len(parsed.Organic),
	)
	return &SearchResponse{Query: query, Organic: parsed.Organic}, nil
}
