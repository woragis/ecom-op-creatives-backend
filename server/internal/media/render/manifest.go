package render

import (
	"encoding/json"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/subtitles"
)

const DefaultIntroDurationMs = 2500

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
	RunID           string            `json:"runId"`
	Format          director.Format   `json:"format"`
	Scenes          []Scene           `json:"scenes"`
	Captions        *subtitles.Output `json:"captions"`
	Audio           Audio             `json:"audio"`
	Product         string            `json:"productName"`
	IntroClip       string            `json:"introClip,omitempty"`
	IntroDurationMs int               `json:"introDurationMs,omitempty"`
	MediaBaseURL    string            `json:"mediaBaseUrl,omitempty"`
}

type Input struct {
	RunID           string
	ProductName     string
	NarrationURL    string
	IntroClip       string
	IntroDurationMs int
	MediaBaseURL    string
	Script          *scriptwriter.Output
	Director        *director.Output
	Captions        *subtitles.Output
	SceneVideos     map[string]string
}

func BuildManifest(in Input) *Manifest {
	sceneVideos := in.SceneVideos
	if sceneVideos == nil {
		sceneVideos = map[string]string{}
	}
	dirMap := map[string]director.SceneDirection{}
	if in.Director != nil {
		for _, s := range in.Director.Scenes {
			dirMap[s.SceneID] = s
		}
	}

	introMs := 0
	if in.IntroClip != "" {
		introMs = in.IntroDurationMs
		if introMs <= 0 {
			introMs = DefaultIntroDurationMs
		}
	}

	var scenes []Scene
	if in.Script != nil {
		for _, sc := range in.Script.Scenes {
			d, ok := dirMap[sc.ID]
			bg := "#1a1a2e"
			tr := director.Transition{Type: "fade", DurationMs: 300}
			if ok {
				bg = d.Background
				tr = d.Transition
			}
			scenes = append(scenes, Scene{
				ID:         sc.ID,
				StartMs:    sc.StartMs + introMs,
				DurationMs: sc.EndMs - sc.StartMs,
				Background: bg,
				Narration:  sc.Narration,
				VideoURL:   sceneVideos[sc.ID],
				Transition: tr,
			})
		}
	}

	caps := in.Captions
	if introMs > 0 && caps != nil {
		caps = subtitles.Offset(caps, introMs)
	}

	musicVol := 0.2
	if in.Director != nil && in.Director.Music.Volume > 0 {
		musicVol = in.Director.Music.Volume
	}

	format := director.Format{Width: 1080, Height: 1920, FPS: 30}
	if in.Director != nil && in.Director.Format.Width > 0 {
		format = in.Director.Format
	}

	return &Manifest{
		RunID:           in.RunID,
		Format:          format,
		Scenes:          scenes,
		Captions:        caps,
		Audio:           Audio{NarrationURL: in.NarrationURL, MusicVolume: musicVol},
		Product:         in.ProductName,
		IntroClip:       in.IntroClip,
		IntroDurationMs: introMs,
		MediaBaseURL:    in.MediaBaseURL,
	}
}

func (m *Manifest) TotalDurationMs() int {
	total := m.IntroDurationMs
	for _, s := range m.Scenes {
		end := s.StartMs + s.DurationMs
		if end > total {
			total = end
		}
	}
	if total <= 0 {
		return 20000
	}
	return total
}

func (m *Manifest) JSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}
