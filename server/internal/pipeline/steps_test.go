package pipeline

import "testing"

func TestQueueForStep(t *testing.T) {
	tests := map[string]string{
		"research":    "pipeline.research",
		"hooks":       "pipeline.llm",
		"voice":       "pipeline.audio",
		"video":       "pipeline.video",
		"render":      "pipeline.render",
		"supervisor":  "pipeline.supervisor",
	}
	for step, want := range tests {
		if got := QueueForStep(step); got != want {
			t.Fatalf("QueueForStep(%q) = %q, want %q", step, got, want)
		}
	}
}

func TestValidVideoProvider(t *testing.T) {
	if !ValidVideoProvider("kling") {
		t.Fatal("expected kling to be valid")
	}
	if ValidVideoProvider("unknown") {
		t.Fatal("expected unknown to be invalid")
	}
}

func TestStepDefinitionsCount(t *testing.T) {
	if len(StepDefinitions) != 12 {
		t.Fatalf("expected 12 steps, got %d", len(StepDefinitions))
	}
}
