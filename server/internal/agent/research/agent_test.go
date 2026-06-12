package research

import (
	"context"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/llm"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/shared/serper"
)

func TestResearchAgentExecute(t *testing.T) {
	agent := New(llm.NewMock(), serper.NewMock())
	out, err := agent.Execute(context.Background(), Input{
		ProductName: "Organizador magnético",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Pains) == 0 {
		t.Fatal("expected pains")
	}
	if len(out.Queries) == 0 {
		t.Fatal("expected queries")
	}
	if len(out.Sources) == 0 {
		t.Fatal("expected sources from serper")
	}
}
