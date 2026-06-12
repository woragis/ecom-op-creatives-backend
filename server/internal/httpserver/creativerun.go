package httpserver

import (
	"encoding/json"
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

type patchStepRequest struct {
	OutputJSON json.RawMessage `json:"outputJson"`
	Reprocess  *bool           `json:"reprocess"`
}

func (h *creativeRunHandler) patchStep(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	stepID, err := uuid.Parse(r.PathValue("stepId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var req patchStepRequest
	if err := decodeJSON(r, &req); err != nil {
		writeServiceError(w, err)
		return
	}
	reprocess := true
	if req.Reprocess != nil {
		reprocess = *req.Reprocess
	}
	run, err := h.svc.EditStep(r.Context(), creativerunsvc.EditStepInput{
		RunID:      runID,
		StepID:     stepID,
		OutputJSON: req.OutputJSON,
		Reprocess:  reprocess,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

type reprocessRequest struct {
	FromStepType string `json:"fromStepType"`
}

func (h *creativeRunHandler) reprocess(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	var req reprocessRequest
	if err := decodeJSON(r, &req); err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.Reprocess(r.Context(), runID, req.FromStepType)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *creativeRunHandler) continueRun(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.ContinueRun(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *creativeRunHandler) approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.ApproveRun(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *creativeRunHandler) retryStep(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	stepID, err := uuid.Parse(r.PathValue("stepId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	run, err := h.svc.RetryStep(r.Context(), runID, stepID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
