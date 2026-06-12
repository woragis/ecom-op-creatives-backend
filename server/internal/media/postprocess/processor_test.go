package postprocess

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

func TestMockProcess(t *testing.T) {
	dir := t.TempDir()
	draft := filepath.Join(dir, "draft.mp4")
	final := filepath.Join(dir, "final.mp4")
	thumb := filepath.Join(dir, "thumb.jpg")
	if err := os.WriteFile(draft, []byte("MOCK_MP4_PHASE1"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := New(config.Config{PostprocessMock: true})
	res, err := p.Process(context.Background(), draft, final, thumb)
	if err != nil {
		t.Fatal(err)
	}
	if res.ThumbnailFrom != "mock" {
		t.Fatal(res)
	}
	if _, err := os.Stat(final); err != nil {
		t.Fatal(err)
	}
}
