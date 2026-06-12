package httpserver

import (
	"net/http"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
)

func handleVideoProviders(configured []string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		all := []string{"kling", "runway", "luma", "veo"}
		configuredSet := map[string]bool{}
		for _, id := range configured {
			configuredSet[id] = true
		}
		items := make([]map[string]any, 0, len(all))
		for _, id := range all {
			if !pipeline.ValidVideoProvider(id) {
				continue
			}
			items = append(items, map[string]any{
				"id":          id,
				"configured":  configuredSet[id],
				"isDefault":   id == pipeline.DefaultVideoProvider(),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}
