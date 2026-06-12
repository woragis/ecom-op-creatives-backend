package httpserver

import "net/http"

func Mount(mux *http.ServeMux, app *App) {
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /ready", handleReady(app.DB, app.RabbitMQ))

	if app.StorageDir != "" {
		mux.Handle("GET /media/", http.StripPrefix("/media/", http.FileServer(http.Dir(app.StorageDir))))
	}

	mux.HandleFunc("GET /v1/video-providers", handleVideoProviders(app.VideoProviders, app.DefaultVideoProvider))
	mux.HandleFunc("GET /v1/image-providers", handleImageProviders(app.ImageProviders, app.DefaultImageProvider))

	if app.Products != nil {
		ph := newProductHandler(app.Products)
		mux.HandleFunc("GET /v1/products", ph.list)
		mux.HandleFunc("POST /v1/products", ph.create)
		mux.HandleFunc("GET /v1/products/{id}", ph.getByID)
	}

	if app.Runs != nil {
		rh := newCreativeRunHandler(app.Runs)
		mux.HandleFunc("GET /v1/creative-runs", rh.list)
		mux.HandleFunc("POST /v1/creative-runs", rh.create)
		mux.HandleFunc("GET /v1/creative-runs/{id}", rh.getByID)
		mux.HandleFunc("POST /v1/creative-runs/{id}/start", rh.start)
		mux.HandleFunc("POST /v1/creative-runs/{id}/assets/{type}", rh.uploadAsset)
		mux.HandleFunc("PATCH /v1/creative-runs/{id}/steps/{stepId}", rh.patchStep)
		mux.HandleFunc("POST /v1/creative-runs/{id}/reprocess", rh.reprocess)
		mux.HandleFunc("POST /v1/creative-runs/{id}/continue", rh.continueRun)
		mux.HandleFunc("POST /v1/creative-runs/{id}/approve", rh.approve)
		mux.HandleFunc("POST /v1/creative-runs/{id}/steps/{stepId}/retry", rh.retryStep)
	}
}
