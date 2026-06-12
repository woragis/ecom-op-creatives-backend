package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, run *models.CreativeRun, steps []models.PipelineStep) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if run.ID == uuid.Nil {
			run.ID = uuid.New()
		}
		if err := tx.Create(run).Error; err != nil {
			return err
		}
		for i := range steps {
			if steps[i].ID == uuid.Nil {
				steps[i].ID = uuid.New()
			}
			steps[i].CreativeRunID = run.ID
		}
		if len(steps) > 0 {
			if err := tx.Create(&steps).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) List(ctx context.Context) ([]models.CreativeRun, error) {
	var items []models.CreativeRun
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.CreativeRun, error) {
	var run models.CreativeRun
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		First(&run, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound(apperrors.CodeCreativeRunGetNotFound, apperrors.MsgCreativeRunGetNotFound)
		}
		return nil, apperrors.Wrapf(err, "creative run get by id")
	}
	return &run, nil
}

func (r *Repository) UpdateRunStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.CreativeRun{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) UpdateInputAssets(ctx context.Context, id uuid.UUID, assets json.RawMessage) error {
	return r.db.WithContext(ctx).Model(&models.CreativeRun{}).
		Where("id = ?", id).
		Update("input_assets", assets).Error
}

func (r *Repository) GetStepByID(ctx context.Context, id uuid.UUID) (*models.PipelineStep, error) {
	var step models.PipelineStep
	err := r.db.WithContext(ctx).First(&step, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound(apperrors.CodeCreativeRunGetNotFound, apperrors.MsgCreativeRunGetNotFound)
		}
		return nil, apperrors.Wrapf(err, "pipeline step get by id")
	}
	return &step, nil
}

func (r *Repository) MarkStepRunning(ctx context.Context, stepID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("id = ?", stepID).
		Updates(map[string]any{
			"status":        models.StepStatusRunning,
			"started_at":    now,
			"attempt_count": gorm.Expr("attempt_count + 1"),
		}).Error
}

func (r *Repository) CompleteStep(ctx context.Context, stepID uuid.UUID, output []byte) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("id = ?", stepID).
		Updates(map[string]any{
			"status":       models.StepStatusDone,
			"output_json":  output,
			"completed_at": gorm.Expr("now()"),
		}).Error
}

func (r *Repository) FailStep(ctx context.Context, stepID uuid.UUID, message string) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("id = ?", stepID).
		Updates(map[string]any{
			"status":        models.StepStatusFailed,
			"error_message": message,
			"completed_at":  gorm.Expr("now()"),
		}).Error
}

func (r *Repository) NextPendingStep(ctx context.Context, runID uuid.UUID, afterOrder int) (*models.PipelineStep, error) {
	var step models.PipelineStep
	err := r.db.WithContext(ctx).
		Where("creative_run_id = ? AND step_order > ? AND status = ?", runID, afterOrder, models.StepStatusPending).
		Order("step_order ASC").
		First(&step).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperrors.Wrapf(err, "next pending step")
	}
	return &step, nil
}

func (r *Repository) FirstStep(ctx context.Context, runID uuid.UUID) (*models.PipelineStep, error) {
	var step models.PipelineStep
	err := r.db.WithContext(ctx).
		Where("creative_run_id = ?", runID).
		Order("step_order ASC").
		First(&step).Error
	if err != nil {
		return nil, apperrors.Wrapf(err, "first step")
	}
	return &step, nil
}

func (r *Repository) GetStepInRun(ctx context.Context, runID, stepID uuid.UUID) (*models.PipelineStep, error) {
	var step models.PipelineStep
	err := r.db.WithContext(ctx).
		Where("id = ? AND creative_run_id = ?", stepID, runID).
		First(&step).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound(apperrors.CodeCreativeRunGetNotFound, apperrors.MsgCreativeRunGetNotFound)
		}
		return nil, apperrors.Wrapf(err, "pipeline step in run")
	}
	return &step, nil
}

func (r *Repository) StepByRunAndOrder(ctx context.Context, runID uuid.UUID, order int) (*models.PipelineStep, error) {
	var step models.PipelineStep
	err := r.db.WithContext(ctx).
		Where("creative_run_id = ? AND step_order = ?", runID, order).
		First(&step).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NotFound(apperrors.CodeCreativeRunGetNotFound, apperrors.MsgCreativeRunGetNotFound)
		}
		return nil, apperrors.Wrapf(err, "pipeline step by order")
	}
	return &step, nil
}

func (r *Repository) UpdateStepOutput(ctx context.Context, stepID uuid.UUID, output json.RawMessage) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("id = ?", stepID).
		Updates(map[string]any{
			"status":       models.StepStatusDone,
			"output_json":  output,
			"error_message": nil,
			"completed_at": gorm.Expr("now()"),
		}).Error
}

func resetStepFields() map[string]any {
	return map[string]any{
		"status":        models.StepStatusPending,
		"output_json":   json.RawMessage(`{}`),
		"error_message": nil,
		"started_at":    nil,
		"completed_at":  nil,
	}
}

// ResetStepsAfterOrder marks steps with step_order > afterOrder as pending.
func (r *Repository) ResetStepsAfterOrder(ctx context.Context, runID uuid.UUID, afterOrder int) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("creative_run_id = ? AND step_order > ?", runID, afterOrder).
		Updates(resetStepFields()).Error
}

// ResetStepsFromOrder marks steps with step_order >= fromOrder as pending.
func (r *Repository) ResetStepsFromOrder(ctx context.Context, runID uuid.UUID, fromOrder int) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("creative_run_id = ? AND step_order >= ?", runID, fromOrder).
		Updates(resetStepFields()).Error
}

func (r *Repository) ResetStepForRetry(ctx context.Context, stepID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.PipelineStep{}).
		Where("id = ?", stepID).
		Updates(map[string]any{
			"status":        models.StepStatusPending,
			"error_message": nil,
			"started_at":    nil,
			"completed_at":  nil,
		}).Error
}
