package video

import (
	"context"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
)

func TestResolveVideoModeForceText2Video(t *testing.T) {
	svc := NewService(config.Config{VideoForceText2Video: true}, NewRegistry(config.Config{}), nil)
	mode, imageURL := svc.resolveVideoMode(context.Background(), ModeImage2Video, "/media/x.jpg", "s1")
	if mode != ModeText2Video || imageURL != "" {
		t.Fatalf("got mode=%s imageURL=%q", mode, imageURL)
	}
}

func TestResolveVideoModeLocalhostFallback(t *testing.T) {
	svc := NewService(config.Config{}, NewRegistry(config.Config{}), nil)
	svc.apiPublicURL = "http://localhost:8080"
	mode, imageURL := svc.resolveVideoMode(context.Background(), ModeImage2Video, "/media/runs/x/persona.jpg", "s1")
	if mode != ModeText2Video || imageURL != "" {
		t.Fatalf("got mode=%s imageURL=%q", mode, imageURL)
	}
}

func TestResolveVideoModeKeepsHTTPS(t *testing.T) {
	svc := NewService(config.Config{}, NewRegistry(config.Config{}), nil)
	url := "https://cdn.example.com/scene.jpg"
	mode, imageURL := svc.resolveVideoMode(context.Background(), ModeImage2Video, url, "s1")
	if mode != ModeImage2Video || imageURL != url {
		t.Fatalf("got mode=%s imageURL=%q", mode, imageURL)
	}
}
