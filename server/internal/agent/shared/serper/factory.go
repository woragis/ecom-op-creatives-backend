package serper

import "github.com/woragis/ecom-op-creatives-backend/server/internal/config"

func NewFromConfig(cfg config.Config) Client {
	if cfg.SerperMock || cfg.SerperKey == "" {
		return NewMock()
	}
	return NewHTTP(cfg.SerperKey)
}
