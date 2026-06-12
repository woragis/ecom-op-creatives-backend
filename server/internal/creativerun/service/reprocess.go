package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
)

func canEditRun(status string) bool {
	switch status {
	case models.RunStatusNeedsReview, models.RunStatusApproved, models.RunStatusFailed:
		return true
	default:
		return false
	}
}

func canUploadAssets(status string) bool {
	switch status {
	case models.RunStatusDraft, models.RunStatusFailed, models.RunStatusNeedsReview, models.RunStatusApproved:
		return true
	default:
		return false
	}
}

func validateOutputJSON(raw json.RawMessage) error {
	if len(raw) == 0 {
		return apperrors.Invalid(apperrors.CodeCreativeRunStepEditInvalid, "outputJson is required")
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return apperrors.Invalid(apperrors.CodeCreativeRunStepEditInvalid, "outputJson must be valid JSON")
	}
	return nil
}

type EditStepInput struct {
	RunID      uuid.UUID
	StepID     uuid.UUID
	OutputJSON json.RawMessage
	Reprocess  bool
}

func (s *Service) EditStep(ctx context.Context, in EditStepInput) (*models.CreativeRun, error) {
	if err := validateOutputJSON(in.OutputJSON); err != nil {
		return nil, err
	}
	run, err := s.repo.GetByID(ctx, in.RunID)
	if err != nil {
		return nil, err
	}
	if run.Status == models.RunStatusRunning {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunStepEditInvalid, apperrors.MsgCreativeRunStepEditInvalid)
	}
	step, err := s.repo.GetStepInRun(ctx, in.RunID, in.StepID)
	if err != nil {
		return nil, err
	}
	if step.Status != models.StepStatusDone {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunStepEditInvalid, apperrors.MsgCreativeRunStepEditInvalid)
	}
	if err := s.repo.UpdateStepOutput(ctx, step.ID, in.OutputJSON); err != nil {
		return nil, apperrors.Wrapf(err, "update step output")
	}
	reprocess := in.Reprocess
	if reprocess {
		if err := s.reprocessAfterOrder(ctx, in.RunID, step.StepOrder); err != nil {
			return nil, err
		}
	}
	return s.repo.GetByID(ctx, in.RunID)
}

func (s *Service) Reprocess(ctx context.Context, runID uuid.UUID, fromStepType string) (*models.CreativeRun, error) {
	order, ok := pipeline.StepOrderForType(fromStepType)
	if !ok {
		return nil, apperrors.Invalid(apperrors.CodeCreativeRunReprocessInvalid, apperrors.MsgCreativeRunReprocessInvalid)
	}
	run, err := s.repo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if !canEditRun(run.Status) && run.Status != models.RunStatusDraft {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunReprocessInvalid, apperrors.MsgCreativeRunReprocessInvalid)
	}
	if run.Status == models.RunStatusRunning {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunReprocessInvalid, apperrors.MsgCreativeRunReprocessInvalid)
	}
	if err := s.reprocessFromOrder(ctx, runID, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, runID)
}

func (s *Service) reprocessAfterOrder(ctx context.Context, runID uuid.UUID, editedOrder int) error {
	if err := s.repo.ResetStepsAfterOrder(ctx, runID, editedOrder); err != nil {
		return apperrors.Wrapf(err, "reset downstream steps")
	}
	nextOrder := editedOrder + 1
	step, err := s.repo.StepByRunAndOrder(ctx, runID, nextOrder)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateRunStatus(ctx, runID, models.RunStatusRunning); err != nil {
		return apperrors.Wrapf(err, "update run status")
	}
	return s.pipelineSvc.EnqueueStep(ctx, step)
}

func (s *Service) reprocessFromOrder(ctx context.Context, runID uuid.UUID, fromOrder int) error {
	if err := s.repo.ResetStepsFromOrder(ctx, runID, fromOrder); err != nil {
		return apperrors.Wrapf(err, "reset steps from order")
	}
	step, err := s.repo.StepByRunAndOrder(ctx, runID, fromOrder)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateRunStatus(ctx, runID, models.RunStatusRunning); err != nil {
		return apperrors.Wrapf(err, "update run status")
	}
	return s.pipelineSvc.EnqueueStep(ctx, step)
}

func (s *Service) reprocessAfterAssetUpload(ctx context.Context, runID uuid.UUID, assetType string) error {
	order, ok := pipeline.ReprocessOrderForAsset(assetType)
	if !ok {
		return nil
	}
	return s.reprocessFromOrder(ctx, runID, order)
}
