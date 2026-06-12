package httpserver

import (
	"net/http"

	"github.com/google/uuid"
	productsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/product/service"
)

type productHandler struct {
	svc *productsvc.Service
}

func newProductHandler(svc *productsvc.Service) *productHandler {
	return &productHandler{svc: svc}
}

type createProductRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	URL         *string `json:"url"`
	Niche       *string `json:"niche"`
}

func (h *productHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *productHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeServiceError(w, err)
		return
	}
	item, err := h.svc.Create(r.Context(), productsvc.CreateInput{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Niche:       req.Niche,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *productHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
