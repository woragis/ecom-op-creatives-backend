package llm

import "github.com/woragis/ecom-op-creatives-backend/server/internal/config"

func NewFromConfig(cfg config.Config) Client {
	if cfg.LLMMock || (cfg.OpenAIKey == "" && cfg.AnthropicKey == "") {
		return NewMock()
	}
	if cfg.LLMProvider == "openai" && cfg.OpenAIKey != "" {
		return NewOpenAI(cfg.OpenAIKey, "gpt-4o-mini")
	}
	if cfg.OpenAIKey != "" {
		return NewOpenAI(cfg.OpenAIKey, "gpt-4o-mini")
	}
	return NewMock()
}
