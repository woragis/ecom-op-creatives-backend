package llm

import (
	"context"
	"encoding/json"
	"strings"
)

type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) CompleteJSON(ctx context.Context, system, user string) ([]byte, error) {
	_ = ctx
	lower := strings.ToLower(system + user)

	switch {
	case strings.Contains(lower, "search quer"):
		return json.Marshal(map[string]any{
			"queries": []string{
				"organizador cabos review",
				"organizador cabos tiktok",
				"cable organizer problems reddit",
			},
		})
	case strings.Contains(lower, "research") && strings.Contains(lower, "synthes"):
		return json.Marshal(map[string]any{
			"pains":              []string{"cabos emaranhados", "mesa bagunçada"},
			"consumerLanguage":   []string{"game changer", "não sabia que precisava"},
			"objections":         []string{"parece frágil"},
			"targetAudience":     "home office, gamers",
			"emotionalTriggers":  []string{"organização", "produtividade"},
			"competitors":        []string{"genérico aliexpress"},
		})
	case strings.Contains(lower, "hook"):
		return json.Marshal(map[string]any{
			"hooks": []map[string]any{
				{"text": "Eu não acreditava que isso funcionava", "score": 92},
				{"text": "3 motivos para organizar sua mesa hoje", "score": 85},
			},
			"selectedHook": "Eu não acreditava que isso funcionava",
		})
	case strings.Contains(lower, "scriptwriter") || strings.Contains(lower, "roteiro"):
		return json.Marshal(map[string]any{
			"scenes": []map[string]any{
				{"id": "s1", "startMs": 0, "endMs": 4000, "narration": "Eu não acreditava que isso funcionava.", "emotion": "curiosity", "goal": "hook"},
				{"id": "s2", "startMs": 4000, "endMs": 12000, "narration": "Até testar este organizador magnético por uma semana.", "emotion": "surprise", "goal": "proof"},
				{"id": "s3", "startMs": 12000, "endMs": 20000, "narration": "Minha mesa nunca mais foi a mesma. Link na bio.", "emotion": "excitement", "goal": "cta"},
			},
			"totalDurationMs": 20000,
		})
	case strings.Contains(lower, "director"):
		return json.Marshal(map[string]any{
			"scenes": []map[string]any{
				{"sceneId": "s1", "camera": "close-up", "transition": map[string]any{"type": "zoom", "durationMs": 400}, "captionStyle": "tiktok-bold", "background": "#1a1a2e"},
				{"sceneId": "s2", "camera": "medium", "transition": map[string]any{"type": "fade", "durationMs": 300}, "captionStyle": "tiktok-bold", "background": "#16213e"},
				{"sceneId": "s3", "camera": "close-up", "transition": map[string]any{"type": "slide", "durationMs": 300}, "captionStyle": "tiktok-bold", "background": "#0f3460"},
			},
			"format": map[string]any{"width": 1080, "height": 1920, "fps": 30},
			"music":  map[string]any{"track": "upbeat", "volume": 0.2},
		})
	case strings.Contains(lower, "prompt"):
		return json.Marshal(map[string]any{
			"scenes": []map[string]any{
				{"sceneId": "s1", "imagePrompt": "UGC style product unboxing surprise"},
			},
		})
	case strings.Contains(lower, "supervisor"):
		return json.Marshal(map[string]any{
			"qualityScore": 88,
			"approved":     true,
			"issues":       []string{},
			"suggestions":  []string{},
		})
	default:
		return json.Marshal(map[string]any{"ok": true})
	}
}
