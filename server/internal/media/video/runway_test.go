package video

import "testing"

func TestRunwayRatio(t *testing.T) {
	if got := runwayRatio("9:16"); got != "720:1280" {
		t.Fatalf("9:16 ratio = %q", got)
	}
	if got := runwayRatio("16:9"); got != "1280:720" {
		t.Fatalf("16:9 ratio = %q", got)
	}
}

func TestParseRunwayOutputURLs(t *testing.T) {
	url, err := parseRunwayOutputURLs([]byte(`["https://cdn.example.com/out.mp4"]`))
	if err != nil || url != "https://cdn.example.com/out.mp4" {
		t.Fatalf("string array: url=%q err=%v", url, err)
	}
	url, err = parseRunwayOutputURLs([]byte(`[{"url":"https://cdn.example.com/legacy.mp4"}]`))
	if err != nil || url != "https://cdn.example.com/legacy.mp4" {
		t.Fatalf("object array: url=%q err=%v", url, err)
	}
}

func TestClampDuration(t *testing.T) {
	if clampDuration(1) != 3 {
		t.Fatal("min duration")
	}
	if clampDuration(15) != 10 {
		t.Fatal("max duration")
	}
}
