package llm

import "context"

type Message struct {
	Role    string
	Content string
}

type Client interface {
	CompleteJSON(ctx context.Context, system string, user string) ([]byte, error)
}
