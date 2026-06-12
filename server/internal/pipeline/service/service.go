package service

import (
	"context"
	"encoding/json"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/rabbitmq"
)

type Publisher interface {
	Publish(ctx context.Context, queue string, body []byte) error
}

type Service struct {
	pub Publisher
}

func New(pub Publisher) *Service {
	return &Service{pub: pub}
}

func (s *Service) EnqueueStep(ctx context.Context, step *models.PipelineStep) error {
	if step == nil {
		return apperrors.InternalErr(apperrors.CodeInternal, "step is nil")
	}
	queue := pipeline.QueueForStep(step.StepType)
	body, err := json.Marshal(rabbitmq.JobMessage{
		CreativeRunID: step.CreativeRunID.String(),
		StepID:        step.ID.String(),
		StepType:      step.StepType,
		Attempt:       step.AttemptCount + 1,
	})
	if err != nil {
		return apperrors.Wrapf(err, "marshal job message")
	}
	applog.FromContext(ctx).With("component", "pipeline", "operation", "enqueue").Info("enqueue step",
		"run_id", step.CreativeRunID.String(),
		"step_id", step.ID.String(),
		"step", step.StepType,
		"queue", queue,
		"attempt", step.AttemptCount+1,
	)
	if err := s.pub.Publish(ctx, queue, body); err != nil {
		return apperrors.Wrapf(err, "publish job")
	}
	return nil
}
