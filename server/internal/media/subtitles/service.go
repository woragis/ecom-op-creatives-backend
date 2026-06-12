package subtitles

import (
	"context"
	"os"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

type Service struct {
	apiKey string
	mock   bool
}

func NewService(cfg config.Config) *Service {
	mock := cfg.SubtitlesMock || cfg.OpenAIKey == ""
	return &Service{apiKey: cfg.OpenAIKey, mock: mock}
}

type GenerateResult struct {
	Output *Output
}

func (s *Service) Generate(ctx context.Context, audioPath string, script *scriptwriter.Output) (*GenerateResult, error) {
	useScript := s.mock
	if !useScript {
		data, err := os.ReadFile(audioPath)
		if err != nil || isMockAudio(data) {
			useScript = true
		}
	}
	if useScript {
		out := FromScript(script)
		out.Source = "script"
		return &GenerateResult{Output: out}, nil
	}
	client := newWhisper(s.apiKey)
	out, err := client.Transcribe(ctx, audioPath)
	if err != nil {
		out := FromScript(script)
		out.Source = "script"
		return &GenerateResult{Output: out}, nil
	}
	if len(out.Words) == 0 {
		fallback := FromScript(script)
		fallback.Source = "script"
		return &GenerateResult{Output: fallback}, nil
	}
	return &GenerateResult{Output: out}, nil
}

// Offset shifts all word timings by offsetMs (e.g. intro clip duration).
func Offset(out *Output, offsetMs int) *Output {
	if out == nil || offsetMs <= 0 {
		return out
	}
	shifted := *out
	shifted.Words = make([]Word, len(out.Words))
	for i, w := range out.Words {
		shifted.Words[i] = Word{
			Text:    w.Text,
			StartMs: w.StartMs + offsetMs,
			EndMs:   w.EndMs + offsetMs,
		}
	}
	return &shifted
}
