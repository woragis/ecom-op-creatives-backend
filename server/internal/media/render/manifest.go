package render

import (
	"encoding/json"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/subtitles"
)

type Scene struct {
	ID         string              `json:"id"`
	StartMs    int                 `json:"startMs"`
	DurationMs int                 `json:"durationMs"`
	Background string              `json:"background"`
	Narration  string              `json:"narration"`
	VideoURL   string              `json:"videoUrl,omitempty"`
	Transition director.Transition `json:"transition"`
}

type Audio struct {
	NarrationURL string  `json:"narrationUrl"`
	MusicVolume  float64 `json:"musicVolume"`
}

type Manifest struct {
	RunID    string             `json:"runId"`
	Format   director.Format    `json:"format"`
	Scenes   []Scene            `json:"scenes"`
	Captions *subtitles.Output  `json:"captions"`
	Audio    Audio              `json:"audio"`
	Product  string             `json:"productName"`
}

func BuildManifest(
	runID, productName, narrationURL string,
	script *scriptwriter.Output,
	dir *director.Output,
	caps *subtitles.Output,
	sceneVideos map[string]string,
) *Manifest {
	if sceneVideos == nil {
		sceneVideos = map[string]string{}
	}
	dirMap := map[string]director.SceneDirection{}
	for _, s := range dir.Scenes {
		dirMap[s.SceneID] = s
	}

	var scenes []Scene
	for _, sc := range script.Scenes {
		d, ok := dirMap[sc.ID]
		bg := "#1a1a2e"
		tr := director.Transition{Type: "fade", DurationMs: 300}
		if ok {
			bg = d.Background
			tr = d.Transition
		}
		scenes = append(scenes, Scene{
			ID:         sc.ID,
			StartMs:    sc.StartMs,
			DurationMs: sc.EndMs - sc.StartMs,
			Background: bg,
			Narration:  sc.Narration,
			VideoURL:   sceneVideos[sc.ID],
			Transition: tr,
		})
	}

	musicVol := 0.2
	if dir.Music.Volume > 0 {
		musicVol = dir.Music.Volume
	}

	return &Manifest{
		RunID:    runID,
		Format:   dir.Format,
		Scenes:   scenes,
		Captions: caps,
		Audio:    Audio{NarrationURL: narrationURL, MusicVolume: musicVol},
		Product:  productName,
	}
}

func (m *Manifest) JSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}
