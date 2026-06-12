package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleImageProviders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/image-providers", handleImageProviders([]string{"flux"}, "flux"))
	req := httptest.NewRequest(http.MethodGet, "/v1/image-providers", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !contains(rec.Body.String(), "flux") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
