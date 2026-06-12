package video

import (
	"fmt"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

type Registry struct {
	providers map[string]Provider
	mock      bool
}

func NewRegistry(cfg config.Config) *Registry {
	r := &Registry{
		providers: map[string]Provider{},
		mock:      cfg.VideoMock,
	}
	if cfg.VideoMock {
		for _, id := range []string{"kling", "runway", "luma", "veo", "mock"} {
			r.providers[id] = NewMock()
		}
		return r
	}
	if cfg.KlingAPIKey != "" {
		r.providers["kling"] = NewKling(cfg.KlingAPIKey, cfg.KlingAPIBase)
	}
	if cfg.RunwayAPIKey != "" {
		r.providers["runway"] = NewRunway(cfg.RunwayAPIKey, cfg.RunwayAPIBase, runwayConfig{
			textModel:  cfg.RunwayTextModel,
			imageModel: cfg.RunwayImageModel,
		})
	}
	if cfg.LumaAPIKey != "" {
		r.providers["luma"] = NewLuma(cfg.LumaAPIKey, cfg.LumaAPIBase)
	}
	if cfg.VeoAPIKey != "" {
		r.providers["veo"] = NewVeo(cfg.VeoAPIKey, cfg.VeoAPIBase)
	}
	return r
}

func (r *Registry) Get(id string) (Provider, error) {
	if p, ok := r.providers[id]; ok {
		return p, nil
	}
	if r.mock {
		return NewMock(), nil
	}
	return nil, fmt.Errorf("video provider %q not configured", id)
}

func (r *Registry) Available() []string {
	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}
