package image

import "testing"

func TestOpenAIImageSize(t *testing.T) {
	if got := openAIImageSize("gpt-image-1-mini", "9:16"); got != "1024x1536" {
		t.Fatalf("gpt portrait = %q", got)
	}
	if got := openAIImageSize("dall-e-3", "9:16"); got != "1024x1792" {
		t.Fatalf("dalle portrait = %q", got)
	}
}
