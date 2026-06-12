package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleVideoProviders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/video-providers", handleVideoProviders([]string{"kling", "runway"}, "kling"))
	req := httptest.NewRequest(http.MethodGet, "/v1/video-providers", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !contains(rec.Body.String(), "kling") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
