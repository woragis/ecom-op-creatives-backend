package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	creativerunrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/repository"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
)

type Service struct {
	repo        *creativerunrepo.Repository
	products    *productrepo.Repository
	pipelineSvc *pipelinesvc.Service
}

func New(repo *creativerunrepo.Repository, products *productrepo.Repository, pipelineSvc *pipelinesvc.Service) *Service {
	return &Service{repo: repo, products: products, pipelineSvc: pipelineSvc}
}

type CreateInput struct {
	ProductID     uuid.UUID
	CampaignID    *uuid.UUID
	Hook          *string
	VideoProvider string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*models.CreativeRun, error) {
	if in.ProductID == uuid.Nil {
		return nil, apperrors.Invalid(apperrors.CodeCreativeRunCreateInvalidBody, apperrors.MsgCreativeRunCreateInvalidBody)
	}
	if _, err := s.products.GetByID(ctx, in.ProductID); err != nil {
		if ae, ok := apperrors.As(err); ok && ae.Kind == apperrors.KindNotFound {
			return nil, apperrors.NotFound(apperrors.CodeCreativeRunCreateProductNotFound, apperrors.MsgCreativeRunCreateProductNotFound)
		}
		return nil, err
	}

	provider := strings.TrimSpace(in.VideoProvider)
	if provider == "" {
		provider = pipeline.DefaultVideoProvider()
	}
	if !pipeline.ValidVideoProvider(provider) {
		return nil, apperrors.Invalid(apperrors.CodeCreativeRunCreateInvalidBody, apperrors.MsgCreativeRunCreateInvalidBody)
	}

	run := &models.CreativeRun{
		ProductID:     in.ProductID,
		CampaignID:    in.CampaignID,
		VideoProvider: provider,
		Status:        models.RunStatusDraft,
		Hook:          in.Hook,
	}

	steps := make([]models.PipelineStep, 0, len(pipeline.StepDefinitions))
	for _, def := range pipeline.StepDefinitions {
		steps = append(steps, models.PipelineStep{
			StepType:  def.Type,
			StepOrder: def.Order,
			Status:    models.StepStatusPending,
			InputJSON: []byte(`{}`),
			OutputJSON: []byte(`{}`),
		})
	}

	if err := s.repo.Create(ctx, run, steps); err != nil {
		return nil, apperrors.Wrapf(err, "creative run create")
	}
	return s.repo.GetByID(ctx, run.ID)
}

func (s *Service) List(ctx context.Context) ([]models.CreativeRun, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.Wrapf(err, "creative run list")
	}
	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.CreativeRun, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Start(ctx context.Context, id uuid.UUID) (*models.CreativeRun, error) {
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run.Status != models.RunStatusDraft && run.Status != models.RunStatusFailed {
		return nil, apperrors.ConflictErr(apperrors.CodeCreativeRunStartInvalidState, apperrors.MsgCreativeRunStartInvalidState)
	}
	if err := s.repo.UpdateRunStatus(ctx, id, models.RunStatusRunning); err != nil {
		return nil, apperrors.Wrapf(err, "creative run start")
	}
	first, err := s.repo.FirstStep(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.pipelineSvc.EnqueueStep(ctx, first); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) CompleteStepStub(ctx context.Context, stepID uuid.UUID) error {
	step, err := s.repo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if err := s.repo.MarkStepRunning(ctx, stepID); err != nil {
		return apperrors.Wrapf(err, "mark step running")
	}
	output, _ := json.Marshal(map[string]any{
		"stub":     true,
		"stepType": step.StepType,
	})
	if err := s.repo.CompleteStep(ctx, stepID, output); err != nil {
		return apperrors.Wrapf(err, "complete step")
	}

	next, err := s.repo.NextPendingStep(ctx, step.CreativeRunID, step.StepOrder)
	if err != nil {
		return err
	}
	if next == nil {
		if err := s.repo.UpdateRunStatus(ctx, step.CreativeRunID, models.RunStatusNeedsReview); err != nil {
			return apperrors.Wrapf(err, "update run status")
		}
		return nil
	}
	return s.pipelineSvc.EnqueueStep(ctx, next)
}
