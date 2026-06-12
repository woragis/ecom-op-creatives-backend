package subtitles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

func TestServiceMockAudioUsesScript(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "narration.mp3")
	if err := os.WriteFile(audio, []byte("MOCK_MP3_PHASE1"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := NewService(config.Config{SubtitlesMock: true})
	res, err := svc.Generate(context.Background(), audio, &scriptwriter.Output{
		Scenes: []scriptwriter.Scene{{StartMs: 0, EndMs: 2000, Narration: "test word"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output.Source != "script" {
		t.Fatalf("source = %s", res.Output.Source)
	}
}

func TestOffset(t *testing.T) {
	out := Offset(&Output{Words: []Word{{Text: "a", StartMs: 0, EndMs: 100}}}, 2500)
	if out.Words[0].StartMs != 2500 {
		t.Fatal(out.Words[0])
	}
}
