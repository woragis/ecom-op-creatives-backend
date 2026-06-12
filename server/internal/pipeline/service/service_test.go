package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
)

type fakePublisher struct {
	lastQueue string
	lastBody  []byte
}

func (f *fakePublisher) Publish(ctx context.Context, queue string, body []byte) error {
	f.lastQueue = queue
	f.lastBody = append([]byte(nil), body...)
	return nil
}

func TestEnqueueStepUsesPipelineQueue(t *testing.T) {
	pub := &fakePublisher{}
	svc := New(pub)
	step := &models.PipelineStep{
		ID:            uuid.New(),
		CreativeRunID: uuid.New(),
		StepType:      "research",
	}
	if err := svc.EnqueueStep(context.Background(), step); err != nil {
		t.Fatal(err)
	}
	if pub.lastQueue != "pipeline.research" {
		t.Fatalf("queue = %q", pub.lastQueue)
	}
	var msg struct {
		StepType string `json:"stepType"`
	}
	if err := json.Unmarshal(pub.lastBody, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.StepType != "research" {
		t.Fatalf("stepType = %q", msg.StepType)
	}
}
