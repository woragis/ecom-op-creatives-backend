package director

import (
	"context"
	"encoding/json"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
)

const systemPrompt = `You direct UGC TikTok/Reels videos (9:16). Return JSON:
{
  "scenes": [
    {"sceneId": "s1", "camera": "close-up", "transition": {"type": "zoom", "durationMs": 400}, "captionStyle": "tiktok-bold", "background": "#1a1a2e"}
  ],
  "format": {"width": 1080, "height": 1920, "fps": 30},
  "music": {"track": "upbeat", "volume": 0.2}
}
Transitions: zoom, fade, slide. Backgrounds: dark vibrant hex colors for UGC text-on-screen style.`

type Transition struct {
	Type       string `json:"type"`
	DurationMs int    `json:"durationMs"`
}

type SceneDirection struct {
	SceneID      string     `json:"sceneId"`
	Camera       string     `json:"camera"`
	Transition   Transition `json:"transition"`
	CaptionStyle string     `json:"captionStyle"`
	Background   string     `json:"background"`
}

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	FPS    int `json:"fps"`
}

type Music struct {
	Track  string  `json:"track"`
	Volume float64 `json:"volume"`
}

type Output struct {
	Scenes []SceneDirection `json:"scenes"`
	Format Format           `json:"format"`
	Music  Music            `json:"music"`
}

type Agent struct{ llm llm.Client }

func New(c llm.Client) *Agent { return &Agent{llm: c} }

func (a *Agent) Execute(ctx context.Context, script *scriptwriter.Output) (*Output, error) {
	payload, _ := json.Marshal(script)
	raw, err := a.llm.CompleteJSON(ctx, systemPrompt, string(payload))
	if err != nil {
		return nil, err
	}
	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out.Format.Width == 0 {
		out.Format = Format{Width: 1080, Height: 1920, FPS: 30}
	}
	return &out, nil
}
