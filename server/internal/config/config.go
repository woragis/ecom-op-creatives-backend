package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	OpenAIKey       string
	AnthropicKey    string
	LLMProvider     string
	LLMMock         bool
	SerperKey       string
	SerperMock      bool
	JinaEnabled     bool
	ElevenLabsKey   string
	ElevenLabsVoice string
	ElevenLabsMock  bool
	StorageDir      string
	RenderDir       string
	SupervisorMin   int
}

func Load() Config {
	return Config{
		OpenAIKey:       strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		AnthropicKey:    strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")),
		LLMProvider:     envOr("LLM_DEFAULT_PROVIDER", "openai"),
		LLMMock:         envBool("LLM_MOCK"),
		SerperKey:       strings.TrimSpace(os.Getenv("SERPER_API_KEY")),
		SerperMock:      envBool("SERPER_MOCK"),
		JinaEnabled:     !envBool("JINA_READER_DISABLED"),
		ElevenLabsKey:   strings.TrimSpace(os.Getenv("ELEVENLABS_API_KEY")),
		ElevenLabsVoice: envOr("ELEVENLABS_VOICE_ID", "21m00Tcm4TlvDq8ikWAM"),
		ElevenLabsMock:  envBool("ELEVENLABS_MOCK"),
		StorageDir:      envOr("STORAGE_DIR", "./storage"),
		RenderDir:       envOr("RENDER_DIR", "../worker-render"),
		SupervisorMin:   envInt("SUPERVISOR_MIN_SCORE", 75),
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	return v == "1" || strings.EqualFold(v, "true")
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
