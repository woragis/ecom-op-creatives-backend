package subtitles

import (
	"fmt"
	"strings"
)

func ToSRT(out *Output) []byte {
	if out == nil || len(out.Words) == 0 {
		return []byte("")
	}
	var b strings.Builder
	for i, w := range out.Words {
		b.WriteString(fmt.Sprintf("%d\n", i+1))
		b.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTime(w.StartMs), formatSRTTime(w.EndMs)))
		b.WriteString(w.Text)
		b.WriteString("\n\n")
	}
	return []byte(b.String())
}

func formatSRTTime(ms int) string {
	if ms < 0 {
		ms = 0
	}
	h := ms / 3600000
	ms %= 3600000
	m := ms / 60000
	ms %= 60000
	s := ms / 1000
	ms %= 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
