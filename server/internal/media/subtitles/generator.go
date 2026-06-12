package subtitles

import (
	"strings"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
)

type Word struct {
	Text    string `json:"text"`
	StartMs int    `json:"startMs"`
	EndMs   int    `json:"endMs"`
}

type Output struct {
	Style string `json:"style"`
	Words []Word `json:"words"`
}

func FromScript(script *scriptwriter.Output) *Output {
	if script == nil || len(script.Scenes) == 0 {
		return &Output{Style: "tiktok-bold", Words: nil}
	}
	var words []Word
	for _, scene := range script.Scenes {
		sceneWords := tokenize(scene.Narration)
		if len(sceneWords) == 0 {
			continue
		}
		duration := scene.EndMs - scene.StartMs
		if duration <= 0 {
			duration = 3000
		}
		slot := duration / len(sceneWords)
		for i, w := range sceneWords {
			start := scene.StartMs + i*slot
			end := start + slot
			words = append(words, Word{Text: w, StartMs: start, EndMs: end})
		}
	}
	return &Output{Style: "tiktok-bold", Words: words}
}

func tokenize(s string) []string {
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, ".,!?;:")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
