package video

import (
	"context"
	"testing"

	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	imagemedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/image"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
)

func TestServiceImage2VideoMock(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{VideoMock: true, VideoMaxScenes: 1, VideoPollIntervalSec: 1, VideoMaxPollMin: 1}
	svc := NewService(cfg, NewRegistry(cfg), store)

	imageOut := &imagemedia.StepOutput{
		Images: []imagemedia.GeneratedImage{
			{SceneID: "s1", Role: imagemedia.RolePersona, PublicURL: "/media/runs/r1/scene_s1_persona.png"},
		},
	}
	out, err := svc.GenerateScenes(context.Background(), "kling", "r1",
		&prompteragent.Output{Scenes: []prompteragent.ScenePrompt{{SceneID: "s1", VideoPrompt: "motion"}}},
		&scriptwriter.Output{Scenes: []scriptwriter.Scene{{ID: "s1", StartMs: 0, EndMs: 5000}}},
		&directoragent.Output{Scenes: []directoragent.SceneDirection{
			{SceneID: "s1", VideoMode: ModeImage2Video, ImageRole: imagemedia.RolePersona},
		}},
		imageOut,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Clips) != 1 {
		t.Fatalf("clips = %d", len(out.Clips))
	}
	if out.Clips[0].Mode != ModeImage2Video {
		t.Fatalf("mode = %s", out.Clips[0].Mode)
	}
}
