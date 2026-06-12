package httpserver

import "net/http"

func Mount(mux *http.ServeMux, app *App) {
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /ready", handleReady(app.DB, app.RabbitMQ))

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
	}
}
