package image

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
		mock:      cfg.ImageMock,
	}
	if cfg.ImageMock {
		for _, id := range []string{"flux", "dalle", "ideogram", "stability", "mock"} {
			r.providers[id] = NewMock()
		}
		return r
	}
	if cfg.FluxAPIKey != "" {
		r.providers["flux"] = NewFlux(cfg.FluxAPIKey, cfg.FluxAPIBase)
	}
	if cfg.OpenAIKey != "" {
		r.providers["dalle"] = NewDalle(cfg.OpenAIKey)
	}
	if cfg.IdeogramAPIKey != "" {
		r.providers["ideogram"] = NewIdeogram(cfg.IdeogramAPIKey)
	}
	if cfg.StabilityAPIKey != "" {
		r.providers["stability"] = NewStability(cfg.StabilityAPIKey)
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
	return nil, fmt.Errorf("image provider %q not configured", id)
}

func (r *Registry) Available() []string {
	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}
