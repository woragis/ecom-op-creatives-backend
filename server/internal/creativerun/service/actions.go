package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
)

func (s *Service) ApproveRun(ctx context.Context, id uuid.UUID) (*models.CreativeRun, error) {
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run.Status != models.RunStatusNeedsReview && run.Status != models.RunStatusApproved {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunApproveInvalid, apperrors.MsgCreativeRunApproveInvalid)
	}
	pending, err := s.repo.HasPendingSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunApproveInvalid, apperrors.MsgCreativeRunApproveInvalid)
	}
	if err := s.repo.UpdateRunStatus(ctx, id, models.RunStatusApproved); err != nil {
		return nil, apperrors.Wrapf(err, "approve run")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ContinueRun(ctx context.Context, id uuid.UUID) (*models.CreativeRun, error) {
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run.Status != models.RunStatusNeedsReview {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunContinueInvalid, apperrors.MsgCreativeRunContinueInvalid)
	}
	next, err := s.repo.FirstPendingStep(ctx, id)
	if err != nil {
		return nil, err
	}
	if next == nil {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunContinueInvalid, apperrors.MsgCreativeRunContinueInvalid)
	}
	if err := s.repo.UpdateRunStatus(ctx, id, models.RunStatusRunning); err != nil {
		return nil, apperrors.Wrapf(err, "continue run")
	}
	if err := s.pipelineSvc.EnqueueStep(ctx, next); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) RetryStep(ctx context.Context, runID, stepID uuid.UUID) (*models.CreativeRun, error) {
	run, err := s.repo.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.Status == models.RunStatusRunning {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunRetryInvalid, apperrors.MsgCreativeRunRetryInvalid)
	}
	step, err := s.repo.GetStepInRun(ctx, runID, stepID)
	if err != nil {
		return nil, err
	}
	if step.Status != models.StepStatusFailed {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunRetryInvalid, apperrors.MsgCreativeRunRetryInvalid)
	}
	if err := s.repo.ResetStepForRetry(ctx, stepID); err != nil {
		return nil, apperrors.Wrapf(err, "reset step for retry")
	}
	if err := s.repo.ResetStepsAfterOrder(ctx, runID, step.StepOrder); err != nil {
		return nil, apperrors.Wrapf(err, "reset downstream for retry")
	}
	if err := s.repo.UpdateRunStatus(ctx, runID, models.RunStatusRunning); err != nil {
		return nil, apperrors.Wrapf(err, "update run status")
	}
	if err := s.pipelineSvc.EnqueueStep(ctx, step); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, runID)
}
