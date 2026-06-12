package tts

import (
	"context"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

func TestOpenAIFallbackWhenNoElevenLabs(t *testing.T) {
	syn := NewFromConfig(config.Config{
		OpenAIKey:      "sk-test",
		OpenAITTSModel: "tts-1",
		OpenAITTSVoice: "alloy",
	})
	c, ok := syn.(*chain)
	if !ok {
		t.Fatal("expected chain")
	}
	if len(c.providers) != 1 || c.providers[0].name != "openai" {
		t.Fatalf("providers = %+v", c.providers)
	}
}

func TestElevenLabsBeforeOpenAI(t *testing.T) {
	syn := NewFromConfig(config.Config{
		ElevenLabsKey:  "el-key",
		ElevenLabsVoice: "voice",
		OpenAIKey:      "sk-test",
	})
	c := syn.(*chain)
	if len(c.providers) != 2 {
		t.Fatalf("want 2 providers, got %d", len(c.providers))
	}
	if c.providers[0].name != "elevenlabs" || c.providers[1].name != "openai" {
		t.Fatalf("order wrong: %+v", c.providers)
	}
}

func TestMockWhenNoKeys(t *testing.T) {
	syn := NewFromConfig(config.Config{})
	audio, provider, err := syn.Synthesize(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if provider != "mock" || len(audio) == 0 {
		t.Fatalf("provider=%s len=%d", provider, len(audio))
	}
}
