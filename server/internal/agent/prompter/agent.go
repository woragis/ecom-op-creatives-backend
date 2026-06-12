package prompter

import (
	"context"
	"encoding/json"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
)

const systemPrompt = `Generate image/video prompts for UGC e-commerce scenes.
Return JSON: {"scenes": [{"sceneId": "s1", "imagePrompt": "...", "videoPrompt": "..."}]}
Style: ultra realistic UGC, natural lighting, TikTok aesthetic.`

type ScenePrompt struct {
	SceneID      string `json:"sceneId"`
	ImagePrompt  string `json:"imagePrompt"`
	VideoPrompt  string `json:"videoPrompt"`
}

type Output struct {
	Scenes []ScenePrompt `json:"scenes"`
}

type Input struct {
	Script      *scriptwriter.Output `json:"script"`
	Director    *director.Output     `json:"director"`
	Product     string               `json:"productName"`
	Description *string              `json:"description,omitempty"`
}

type Agent struct{ llm llm.Client }

func New(c llm.Client) *Agent { return &Agent{llm: c} }

func (a *Agent) Execute(ctx context.Context, in Input) (*Output, error) {
	payload, _ := json.Marshal(in)
	raw, err := a.llm.CompleteJSON(ctx, systemPrompt, string(payload))
	if err != nil {
		return nil, err
	}
	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
