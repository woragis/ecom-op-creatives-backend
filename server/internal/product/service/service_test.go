package service

import (
	"context"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
)

func TestCreateRequiresName(t *testing.T) {
	s := New(nil)
	_, err := s.Create(context.Background(), CreateInput{Name: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := apperrors.As(err)
	if !ok || ae.Kind != apperrors.KindInvalid {
		t.Fatalf("expected invalid error, got %v", err)
	}
}
