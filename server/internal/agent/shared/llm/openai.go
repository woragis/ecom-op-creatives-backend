package llm

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

const openAIURL = "https://api.openai.com/v1/chat/completions"

type OpenAI struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAI {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAI{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

type openAIRequest struct {
	Model          string          `json:"model"`
	Messages       []openAIMessage `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *OpenAI) CompleteJSON(ctx context.Context, system, user string) ([]byte, error) {
	started := time.Now()
	log := applog.FromContext(ctx).With("service", "openai", "operation", "chat.completions", "model", o.model)
	log.Debug("llm request",
		"system_preview", applog.Truncate(system, 120),
		"user_preview", applog.Truncate(user, 240),
		"system_chars", len(system),
		"user_chars", len(user),
	)

	reqBody := openAIRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
		Temperature:    0.7,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		log.Error("llm request failed",
			"status", res.StatusCode,
			"duration_ms", time.Since(started).Milliseconds(),
			"body_preview", applog.Truncate(string(raw), 300),
		)
		return nil, fmt.Errorf("openai http %d: %s", res.StatusCode, string(raw))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil {
		log.Error("llm api error", "message", parsed.Error.Message, "duration_ms", time.Since(started).Milliseconds())
		return nil, fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response")
	}
	log.Info("llm request completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"prompt_tokens", parsed.Usage.PromptTokens,
		"completion_tokens", parsed.Usage.CompletionTokens,
		"total_tokens", parsed.Usage.TotalTokens,
		"response_chars", len(parsed.Choices[0].Message.Content),
	)
	return []byte(parsed.Choices[0].Message.Content), nil
}
