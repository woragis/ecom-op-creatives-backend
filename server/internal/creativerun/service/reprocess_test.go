package service

import (
	"encoding/json"
	"testing"
)

func TestValidateOutputJSON(t *testing.T) {
	if err := validateOutputJSON(json.RawMessage(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := validateOutputJSON(json.RawMessage(`not json`)); err == nil {
		t.Fatal("expected error")
	}
	if err := validateOutputJSON(nil); err == nil {
		t.Fatal("expected error for empty")
	}
}
