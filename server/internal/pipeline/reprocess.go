package pipeline

// StepOrderForType returns the pipeline order for a step type.
func StepOrderForType(stepType string) (int, bool) {
	for _, def := range StepDefinitions {
		if def.Type == stepType {
			return def.Order, true
		}
	}
	return 0, false
}

// ReprocessOrderForAsset returns the first step to re-run after an input asset changes.
func ReprocessOrderForAsset(assetType string) (int, bool) {
	switch assetType {
	case "persona", "product":
		return 7, true // image
	case "intro":
		return 10, true // render
	default:
		return 0, false
	}
}
