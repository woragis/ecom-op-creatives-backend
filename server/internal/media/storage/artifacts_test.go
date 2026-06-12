package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteStepArtifact(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	runID := "run-1"
	out := []byte(`{"ok":true}`)
	if err := store.WriteStepArtifact(runID, "script", out); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "runs", runID, "artifacts", "03-script.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(out) {
		t.Fatalf("got %s", raw)
	}
}
