package service

import "testing"

func TestCanEditRun(t *testing.T) {
	if !canEditRun("needs_review") {
		t.Fatal("expected needs_review editable")
	}
	if canEditRun("running") {
		t.Fatal("running should not be editable")
	}
}
