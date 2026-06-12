package elevenlabs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	apiKey  string
	voiceID string
	mock    bool
	hc      *http.Client
}

func New(apiKey, voiceID string, mock bool) *Client {
	return &Client{
		apiKey:  apiKey,
		voiceID: voiceID,
		mock:    mock,
		hc:      &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if c.mock || c.apiKey == "" {
		return mockMP3(), nil
	}
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", c.voiceID)
	body := []byte(fmt.Sprintf(`{"text":%q,"model_id":"eleven_multilingual_v2"}`, text))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("elevenlabs http %d: %s", res.StatusCode, string(b))
	}
	return io.ReadAll(res.Body)
}

// mockMP3 returns minimal bytes so pipeline can proceed without API.
func mockMP3() []byte {
	if b, err := os.ReadFile("testdata/mock.mp3"); err == nil {
		return b
	}
	// ID3 header stub — not playable in all players but enough for pipeline tests
	return []byte("MOCK_MP3_PHASE1")
}
