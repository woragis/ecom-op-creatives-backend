package video

import (
	"context"
	"testing"

	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
)

func TestServiceGenerateScenesMock(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{VideoMock: true, VideoMaxScenes: 2, VideoPollIntervalSec: 1, VideoMaxPollMin: 1}
	reg := NewRegistry(cfg)
	svc := NewService(cfg, reg, store)

	out, err := svc.GenerateScenes(context.Background(), "kling", "run-1",
		&prompteragent.Output{Scenes: []prompteragent.ScenePrompt{
			{SceneID: "s1", VideoPrompt: "UGC product shot"},
			{SceneID: "s2", VideoPrompt: "happy customer"},
		}},
		&scriptwriter.Output{Scenes: []scriptwriter.Scene{
			{ID: "s1", StartMs: 0, EndMs: 5000, Narration: "hook"},
			{ID: "s2", StartMs: 5000, EndMs: 12000, Narration: "proof"},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.Skipped {
		t.Fatal("expected not skipped")
	}
	if len(out.Clips) != 2 {
		t.Fatalf("clips = %d", len(out.Clips))
	}
	if out.Clips[0].PublicURL == "" {
		t.Fatal("expected public url")
	}
}

func TestClipsBySceneID(t *testing.T) {
	m := ClipsBySceneID(&StepOutput{Clips: []Clip{{SceneID: "s1", PublicURL: "/a.mp4"}}})
	if m["s1"] != "/a.mp4" {
		t.Fatal(m)
	}
}
