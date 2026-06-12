package httpserver

import (
	"net/http"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/middleware"
)

func NewHandler(app *App, cfg middleware.Config) http.Handler {
	mux := http.NewServeMux()
	Mount(mux, app)
	return middleware.Chain(cfg, mux)
}
