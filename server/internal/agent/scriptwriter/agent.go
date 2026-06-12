package scriptwriter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/hooks"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/research"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
)

const systemPrompt = `You write UGC TikTok/Reels scripts for e-commerce (15-45 seconds).
Return JSON:
{
  "scenes": [
    {"id": "s1", "startMs": 0, "endMs": 4000, "narration": "...", "emotion": "curiosity", "goal": "hook"}
  ],
  "totalDurationMs": 20000
}
Rules: short punchy sentences, conversational Portuguese (BR), strong hook in s1, CTA in last scene.`

type Input struct {
	ProductName string           `json:"productName"`
	Research    *research.Output `json:"research"`
	Hook        *hooks.Output    `json:"hook"`
}

type Scene struct {
	ID        string `json:"id"`
	StartMs   int    `json:"startMs"`
	EndMs     int    `json:"endMs"`
	Narration string `json:"narration"`
	Emotion   string `json:"emotion"`
	Goal      string `json:"goal"`
}

type Output struct {
	Scenes          []Scene `json:"scenes"`
	TotalDurationMs int     `json:"totalDurationMs"`
}

type Agent struct{ llm llm.Client }

func New(c llm.Client) *Agent { return &Agent{llm: c} }

func (a *Agent) Execute(ctx context.Context, in Input) (*Output, error) {
	payload, _ := json.Marshal(in)
	user := fmt.Sprintf("Write script for:\n%s", string(payload))
	raw, err := a.llm.CompleteJSON(ctx, systemPrompt, user)
	if err != nil {
		return nil, err
	}
	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func FullNarration(out *Output) string {
	var parts []string
	for _, s := range out.Scenes {
		parts = append(parts, s.Narration)
	}
	return joinSentences(parts)
}

func joinSentences(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}
