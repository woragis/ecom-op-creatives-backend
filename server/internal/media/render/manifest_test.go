package render

import (
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/subtitles"
)

func TestBuildManifestIntroOffset(t *testing.T) {
	m := BuildManifest(Input{
		RunID:           "r1",
		IntroClip:       "/media/runs/r1/input/intro.mp4",
		IntroDurationMs: 2500,
		Script: &scriptwriter.Output{
			Scenes: []scriptwriter.Scene{
				{ID: "s1", StartMs: 0, EndMs: 4000, Narration: "hook"},
			},
		},
		Captions: &subtitles.Output{
			Words: []subtitles.Word{{Text: "hook", StartMs: 0, EndMs: 400}},
		},
	})
	if m.Scenes[0].StartMs != 2500 {
		t.Fatalf("scene start = %d", m.Scenes[0].StartMs)
	}
	if m.Captions.Words[0].StartMs != 2500 {
		t.Fatalf("caption start = %d", m.Captions.Words[0].StartMs)
	}
	if m.TotalDurationMs() < 6500 {
		t.Fatalf("total = %d", m.TotalDurationMs())
	}
}
