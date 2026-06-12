package serper

import "context"

type Result struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

type SearchResponse struct {
	Query   string   `json:"query"`
	Organic []Result `json:"organic"`
}

type Client interface {
	Search(ctx context.Context, query string) (*SearchResponse, error)
}
