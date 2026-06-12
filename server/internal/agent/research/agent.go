package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/serper"
)

const (
	querySystemPrompt = `You generate Google search queries for e-commerce product research.
Return JSON: {"queries": ["...", "..."]}
Generate 6-8 queries mixing: reviews, pain points, tiktok, reddit, amazon 1-star, competitor terms.
Language: match product market (Portuguese for BR products, English otherwise).`

	synthSystemPrompt = `You synthesize e-commerce consumer research from search snippets.
Return JSON only:
{
  "pains": ["..."],
  "consumerLanguage": ["..."],
  "objections": ["..."],
  "targetAudience": "...",
  "emotionalTriggers": ["..."],
  "competitors": ["..."]
}
Use real language from snippets. Be specific to the product niche.`
)

type Input struct {
	ProductName string  `json:"productName"`
	ProductURL  *string `json:"productUrl,omitempty"`
	Niche       *string `json:"niche,omitempty"`
}

type Source struct {
	Query   string `json:"query"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

type Output struct {
	Pains             []string `json:"pains"`
	ConsumerLanguage  []string `json:"consumerLanguage"`
	Objections        []string `json:"objections"`
	TargetAudience    string   `json:"targetAudience"`
	EmotionalTriggers []string `json:"emotionalTriggers"`
	Competitors       []string `json:"competitors"`
	Queries           []string `json:"queries"`
	Sources           []Source `json:"sources"`
}

type Agent struct {
	llm    llm.Client
	serper serper.Client
}

func New(llmClient llm.Client, serperClient serper.Client) *Agent {
	return &Agent{llm: llmClient, serper: serperClient}
}

func (a *Agent) Execute(ctx context.Context, in Input) (*Output, error) {
	queries, err := a.generateQueries(ctx, in)
	if err != nil {
		return nil, err
	}

	sources, err := a.searchAll(ctx, queries)
	if err != nil {
		return nil, err
	}

	out, err := a.synthesize(ctx, in, sources)
	if err != nil {
		return nil, err
	}
	out.Queries = queries
	out.Sources = sources
	return out, nil
}

func (a *Agent) generateQueries(ctx context.Context, in Input) ([]string, error) {
	user := fmt.Sprintf("Product: %s\nNiche: %s\nURL: %s",
		in.ProductName, strOr(in.Niche, "general"), strOr(in.ProductURL, ""))
	raw, err := a.llm.CompleteJSON(ctx, querySystemPrompt, user)
	if err != nil {
		return nil, fmt.Errorf("generate queries: %w", err)
	}
	var parsed struct {
		Queries []string `json:"queries"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Queries) == 0 {
		return []string{in.ProductName + " review", in.ProductName + " tiktok"}, nil
	}
	return parsed.Queries, nil
}

func (a *Agent) searchAll(ctx context.Context, queries []string) ([]Source, error) {
	var sources []Source
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		res, err := a.serper.Search(ctx, q)
		if err != nil {
			continue
		}
		for _, hit := range res.Organic {
			sources = append(sources, Source{
				Query:   q,
				Title:   hit.Title,
				Link:    hit.Link,
				Snippet: hit.Snippet,
			})
		}
	}
	return sources, nil
}

func (a *Agent) synthesize(ctx context.Context, in Input, sources []Source) (*Output, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Product: %s\nNiche: %s\n\nSearch results:\n", in.ProductName, strOr(in.Niche, "")))
	for i, s := range sources {
		if i >= 25 {
			break
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", s.Query, s.Title, s.Snippet))
	}

	raw, err := a.llm.CompleteJSON(ctx, synthSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("synthesize: %w", err)
	}
	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func strOr(p *string, def string) string {
	if p != nil && *p != "" {
		return *p
	}
	return def
}
