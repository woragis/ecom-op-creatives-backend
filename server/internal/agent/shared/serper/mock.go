package serper

import "context"

type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Search(ctx context.Context, query string) (*SearchResponse, error) {
	_ = ctx
	return &SearchResponse{
		Query: query,
		Organic: []Result{
			{Title: "Review " + query, Link: "https://example.com/review", Snippet: "Game changer for desk organization"},
			{Title: "Reddit " + query, Link: "https://reddit.com/r/buyitforlife", Snippet: "Best cable organizer I've tried"},
		},
	}, nil
}
