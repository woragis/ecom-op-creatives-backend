package tts

import "context"

type Synthesizer interface {
	Synthesize(ctx context.Context, text string) (audio []byte, provider string, err error)
}
