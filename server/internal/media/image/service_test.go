package image

import (
	"context"
	"testing"

	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
)

func TestServiceGenerateScenesMock(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{ImageMock: true, ImageMaxScenes: 2}
	reg := NewRegistry(cfg)
	svc := NewService(cfg, reg, store)

	out, err := svc.GenerateScenes(context.Background(), "flux", "run-1",
		&prompteragent.Output{Scenes: []prompteragent.ScenePrompt{
			{SceneID: "s1", ImagePrompt: "UGC persona"},
			{SceneID: "s2", ImagePrompt: "product shot"},
		}},
		&scriptwriter.Output{Scenes: []scriptwriter.Scene{
			{ID: "s1", StartMs: 0, EndMs: 4000, Narration: "hook"},
			{ID: "s2", StartMs: 4000, EndMs: 12000, Narration: "proof"},
		}},
		&directoragent.Output{Scenes: []directoragent.SceneDirection{
			{SceneID: "s1", ImageRole: RolePersona, VideoMode: directoragent.VideoModeImage2Video},
			{SceneID: "s2", ImageRole: RoleProduct, VideoMode: directoragent.VideoModeImage2Video},
		}},
		&models.RunAssets{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.Skipped {
		t.Fatal("expected not skipped")
	}
	if len(out.Images) != 2 {
		t.Fatalf("images = %d", len(out.Images))
	}
}

func TestServiceUsesUserAsset(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{ImageMock: true, ImageMaxScenes: 1}
	svc := NewService(cfg, NewRegistry(cfg), store)

	out, err := svc.GenerateScenes(context.Background(), "flux", "run-2",
		&prompteragent.Output{Scenes: []prompteragent.ScenePrompt{{SceneID: "s1", ImagePrompt: "x"}}},
		&scriptwriter.Output{Scenes: []scriptwriter.Scene{{ID: "s1", StartMs: 0, EndMs: 4000}}},
		&directoragent.Output{Scenes: []directoragent.SceneDirection{{SceneID: "s1", ImageRole: RolePersona}}},
		&models.RunAssets{PersonaImage: "/media/runs/run-2/input/persona.jpg"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Images) != 1 || out.Images[0].Source != "user_asset" {
		t.Fatalf("got %+v", out.Images)
	}
}
