package pipeline

var StepDefinitions = []struct {
	Type  string
	Order int
}{
	{Type: "research", Order: 1},
	{Type: "hooks", Order: 2},
	{Type: "script", Order: 3},
	{Type: "director", Order: 4},
	{Type: "prompter", Order: 5},
	{Type: "voice", Order: 6},
	{Type: "image", Order: 7},
	{Type: "video", Order: 8},
	{Type: "subtitles", Order: 9},
	{Type: "render", Order: 10},
	{Type: "postprocess", Order: 11},
	{Type: "supervisor", Order: 12},
}

func QueueForStep(stepType string) string {
	switch stepType {
	case "research":
		return "pipeline.research"
	case "voice":
		return "pipeline.audio"
	case "image":
		return "pipeline.image"
	case "video":
		return "pipeline.video"
	case "render":
		return "pipeline.render"
	case "postprocess":
		return "pipeline.media"
	case "supervisor":
		return "pipeline.supervisor"
	default:
		return "pipeline.llm"
	}
}

func AllQueues() []string {
	return []string{
		"pipeline.research",
		"pipeline.llm",
		"pipeline.audio",
		"pipeline.image",
		"pipeline.video",
		"pipeline.render",
		"pipeline.media",
		"pipeline.supervisor",
	}
}

func ValidVideoProvider(provider string) bool {
	switch provider {
	case "kling", "runway", "luma", "veo":
		return true
	default:
		return false
	}
}

func DefaultVideoProvider() string { return "kling" }
