package hooks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/research"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
)

const systemPrompt = `You create viral TikTok/Reels hooks for e-commerce products.
Return JSON:
{
  "hooks": [{"text": "...", "score": 90}],
  "selectedHook": "..."
}
Generate 10 hooks using consumer language from research. selectedHook = highest scoring hook.
Hooks in Portuguese for BR audience unless product is clearly international.`

type Input struct {
	ProductName string           `json:"productName"`
	Research    *research.Output `json:"research"`
	UserHook    *string          `json:"userHook,omitempty"`
}

type Hook struct {
	Text  string `json:"text"`
	Score int    `json:"score"`
}

type Output struct {
	Hooks        []Hook `json:"hooks"`
	SelectedHook string `json:"selectedHook"`
}

type Agent struct{ llm llm.Client }

func New(c llm.Client) *Agent { return &Agent{llm: c} }

func (a *Agent) Execute(ctx context.Context, in Input) (*Output, error) {
	if in.UserHook != nil && *in.UserHook != "" {
		return &Output{
			Hooks:        []Hook{{Text: *in.UserHook, Score: 100}},
			SelectedHook: *in.UserHook,
		}, nil
	}
	researchJSON, _ := json.Marshal(in.Research)
	user := fmt.Sprintf("Product: %s\nResearch:\n%s", in.ProductName, string(researchJSON))
	raw, err := a.llm.CompleteJSON(ctx, systemPrompt, user)
	if err != nil {
		return nil, err
	}
	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out.SelectedHook == "" && len(out.Hooks) > 0 {
		out.SelectedHook = out.Hooks[0].Text
	}
	return &out, nil
}
