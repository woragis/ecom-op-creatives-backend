package httpserver

import (
	"net/http"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
)

func handleImageProviders(configured []string, defaultProvider string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		all := []string{"flux", "dalle", "ideogram", "stability"}
		configuredSet := map[string]bool{}
		for _, id := range configured {
			configuredSet[id] = true
		}
		if defaultProvider == "" {
			defaultProvider = pipeline.DefaultImageProvider()
		}
		items := make([]map[string]any, 0, len(all))
		for _, id := range all {
			if !pipeline.ValidImageProvider(id) {
				continue
			}
			items = append(items, map[string]any{
				"id":          id,
				"configured":  configuredSet[id],
				"isDefault":   id == defaultProvider,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}
