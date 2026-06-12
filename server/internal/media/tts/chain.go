package tts

import (
	"context"
	"fmt"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/elevenlabs"
)

type provider struct {
	name string
	syn  func(context.Context, string) ([]byte, error)
}

type chain struct {
	providers []provider
	mock      func(context.Context, string) ([]byte, error)
}

func (c *chain) Synthesize(ctx context.Context, text string) ([]byte, string, error) {
	var lastErr error
	for _, p := range c.providers {
		audio, err := p.syn(ctx, text)
		if err == nil {
			return audio, p.name, nil
		}
		lastErr = err
	}
	if c.mock != nil {
		audio, err := c.mock(ctx, text)
		if err != nil {
			return nil, "", err
		}
		return audio, "mock", nil
	}
	if lastErr != nil {
		return nil, "", fmt.Errorf("tts: all providers failed: %w", lastErr)
	}
	return nil, "", fmt.Errorf("tts: no provider configured")
}

func NewFromConfig(cfg config.Config) Synthesizer {
	var providers []provider

	if cfg.ElevenLabsKey != "" && !cfg.ElevenLabsMock {
		el := elevenlabs.New(cfg.ElevenLabsKey, cfg.ElevenLabsVoice, false)
		providers = append(providers, provider{
			name: "elevenlabs",
			syn:  el.Synthesize,
		})
	}
	if cfg.OpenAIKey != "" {
		oai := NewOpenAI(cfg.OpenAIKey, cfg.OpenAITTSModel, cfg.OpenAITTSVoice)
		providers = append(providers, provider{
			name: "openai",
			syn:  oai.Synthesize,
		})
	}
	if len(providers) == 0 {
		mock := elevenlabs.New("", "", true)
		return &chain{
			mock: mock.Synthesize,
		}
	}
	return &chain{providers: providers}
}
