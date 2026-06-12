package subtitles

import (
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
)

func TestFromScriptWordTiming(t *testing.T) {
	script := &scriptwriter.Output{
		Scenes: []scriptwriter.Scene{
			{StartMs: 0, EndMs: 2000, Narration: "Eu não acreditava"},
		},
	}
	out := FromScript(script)
	if len(out.Words) < 2 {
		t.Fatalf("expected at least 2 words, got %d", len(out.Words))
	}
	if out.Words[0].StartMs != 0 {
		t.Fatalf("first word start = %d", out.Words[0].StartMs)
	}
}
