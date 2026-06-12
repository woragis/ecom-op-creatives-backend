package pipeline

import "testing"

func TestReprocessOrderForAsset(t *testing.T) {
	if o, ok := ReprocessOrderForAsset("persona"); !ok || o != 7 {
		t.Fatalf("persona = %d %v", o, ok)
	}
	if o, ok := ReprocessOrderForAsset("intro"); !ok || o != 10 {
		t.Fatalf("intro = %d %v", o, ok)
	}
}

func TestStepOrderForType(t *testing.T) {
	if o, ok := StepOrderForType("director"); !ok || o != 4 {
		t.Fatalf("director = %d", o)
	}
}
