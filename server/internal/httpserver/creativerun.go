package httpserver

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	creativerunsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/service"
)

type creativeRunHandler struct {
	svc *creativerunsvc.Service
}

func newCreativeRunHandler(svc *creativerunsvc.Service) *creativeRunHandler {
	return &creativeRunHandler{svc: svc}
}

type createCreativeRunRequest struct {
	ProductID     string  `json:"productId"`
	CampaignID    *string `json:"campaignId"`
	Hook          *string `json:"hook"`
	VideoProvider string  `json:"videoProvider"`
	ImageProvider string  `json:"imageProvider"`
}

func (h *creativeRunHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *creativeRunHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createCreativeRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeServiceError(w, err)
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var campaignID *uuid.UUID
	if req.CampaignID != nil && *req.CampaignID != "" {
		id, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		campaignID = &id
	}
	run, err := h.svc.Create(r.Context(), creativerunsvc.CreateInput{
		ProductID:     productID,
		CampaignID:    campaignID,
		Hook:          req.Hook,
		VideoProvider: req.VideoProvider,
		ImageProvider: req.ImageProvider,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (h *creativeRunHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *creativeRunHandler) start(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.Start(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *creativeRunHandler) uploadAsset(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	assetType := strings.TrimSpace(r.PathValue("type"))
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeServiceError(w, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeServiceError(w, err)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	ext := filepath.Ext(header.Filename)
	assets, err := h.svc.UploadAsset(r.Context(), id, assetType, data, ext)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assets)
}
