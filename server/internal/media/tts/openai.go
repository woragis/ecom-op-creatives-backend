package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openAISpeechURL = "https://api.openai.com/v1/audio/speech"

type OpenAI struct {
	apiKey string
	model  string
	voice  string
	hc     *http.Client
}

func NewOpenAI(apiKey, model, voice string) *OpenAI {
	if model == "" {
		model = "tts-1"
	}
	if voice == "" {
		voice = "alloy"
	}
	return &OpenAI{
		apiKey: apiKey,
		model:  model,
		voice:  voice,
		hc:     &http.Client{Timeout: 120 * time.Second},
	}
}

func (o *OpenAI) Synthesize(ctx context.Context, text string) ([]byte, error) {
	body, err := json.Marshal(map[string]string{
		"model":           o.model,
		"input":           text,
		"voice":           o.voice,
		"response_format": "mp3",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAISpeechURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := o.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("openai tts http %d: %s", res.StatusCode, string(raw))
	}
	return raw, nil
}
