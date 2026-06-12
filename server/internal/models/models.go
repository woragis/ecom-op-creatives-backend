package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `json:"name"`
	URL       *string   `json:"url,omitempty"`
	Niche     *string   `json:"niche,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Campaign struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID   uuid.UUID  `gorm:"type:uuid;not null" json:"productId"`
	Name        string     `json:"name"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CreativeRun struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID     uuid.UUID  `gorm:"type:uuid;not null" json:"productId"`
	CampaignID    *uuid.UUID `gorm:"type:uuid" json:"campaignId,omitempty"`
	VideoProvider string     `json:"videoProvider"`
	Status        string     `json:"status"`
	Hook          *string    `json:"hook,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	Steps []PipelineStep `gorm:"foreignKey:CreativeRunID" json:"steps,omitempty"`
}

type PipelineStep struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreativeRunID uuid.UUID `gorm:"type:uuid;not null" json:"creativeRunId"`
	StepType      string    `json:"stepType"`
	StepOrder     int       `json:"stepOrder"`
	Status        string    `json:"status"`
	InputJSON     []byte    `gorm:"type:jsonb" json:"inputJson"`
	OutputJSON    []byte    `gorm:"type:jsonb" json:"outputJson"`
	AttemptCount  int            `json:"attemptCount"`
	ErrorMessage  *string        `json:"errorMessage,omitempty"`
	ProviderUsed  *string        `json:"providerUsed,omitempty"`
	StartedAt     *time.Time     `json:"startedAt,omitempty"`
	CompletedAt   *time.Time     `json:"completedAt,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

const (
	RunStatusDraft        = "draft"
	RunStatusRunning      = "running"
	RunStatusNeedsReview  = "needs_review"
	RunStatusApproved     = "approved"
	RunStatusFailed       = "failed"

	StepStatusPending      = "pending"
	StepStatusRunning      = "running"
	StepStatusDone         = "done"
	StepStatusFailed       = "failed"
	StepStatusInvalidated  = "invalidated"
)
