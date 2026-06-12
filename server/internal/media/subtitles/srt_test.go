package subtitles

import "testing"

func TestToSRT(t *testing.T) {
	out := &Output{
		Words: []Word{{Text: "hello", StartMs: 0, EndMs: 500}},
	}
	srt := string(ToSRT(out))
	if srt == "" || srt[0] != '1' {
		t.Fatalf("srt = %q", srt)
	}
}
