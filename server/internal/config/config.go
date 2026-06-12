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
	VideoMock       bool
	VideoMaxScenes  int
	VideoPollIntervalSec int
	VideoMaxPollMin int
	KlingAPIKey     string
	KlingAPIBase    string
	RunwayAPIKey    string
	RunwayAPIBase   string
	LumaAPIKey      string
	LumaAPIBase     string
	VeoAPIKey       string
	VeoAPIBase      string
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
		SupervisorMin:        envInt("SUPERVISOR_MIN_SCORE", 75),
		VideoMock:            envBool("VIDEO_MOCK"),
		VideoMaxScenes:       envInt("VIDEO_MAX_SCENES", 3),
		VideoPollIntervalSec: envInt("VIDEO_POLL_INTERVAL_SEC", 5),
		VideoMaxPollMin:      envInt("VIDEO_MAX_POLL_MIN", 15),
		KlingAPIKey:          strings.TrimSpace(os.Getenv("KLING_API_KEY")),
		KlingAPIBase:         envOr("KLING_API_BASE", ""),
		RunwayAPIKey:         strings.TrimSpace(os.Getenv("RUNWAY_API_KEY")),
		RunwayAPIBase:        envOr("RUNWAY_API_BASE", ""),
		LumaAPIKey:           strings.TrimSpace(os.Getenv("LUMA_API_KEY")),
		LumaAPIBase:          envOr("LUMA_API_BASE", ""),
		VeoAPIKey:            strings.TrimSpace(os.Getenv("GOOGLE_VEO_API_KEY")),
		VeoAPIBase:           envOr("GOOGLE_VEO_API_BASE", ""),
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
