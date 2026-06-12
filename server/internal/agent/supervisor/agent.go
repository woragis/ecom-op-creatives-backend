package supervisor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
)

const systemPrompt = `You QA e-commerce UGC video creatives. Return JSON:
{
  "qualityScore": 85,
  "approved": true,
  "issues": [],
  "suggestions": [],
  "checks": {"hookStrength": 90, "scriptFlow": 85, "ctaClarity": 80}
}
Score 0-100. approved=true if score>=75 and hook is strong.`

type Input struct {
	ProductName string          `json:"productName"`
	Steps       map[string]any  `json:"steps"`
}

type Output struct {
	QualityScore int      `json:"qualityScore"`
	Approved     bool     `json:"approved"`
	Issues       []string `json:"issues"`
	Suggestions  []string `json:"suggestions"`
}

type Agent struct {
	llm     llm.Client
	minScore int
}

func New(c llm.Client, minScore int) *Agent {
	if minScore <= 0 {
		minScore = 75
	}
	return &Agent{llm: c, minScore: minScore}
}

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
	if !out.Approved && out.QualityScore >= a.minScore {
		out.Approved = true
	}
	if out.QualityScore == 0 {
		return nil, fmt.Errorf("supervisor returned empty score")
	}
	return &out, nil
}
