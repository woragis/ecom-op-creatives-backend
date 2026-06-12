package postprocess

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
)

// minimal JPEG SOI/EOI
var mockJPEG = []byte{
	0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01, 0x11, 0x00,
	0xff, 0xc4, 0x00, 0x14, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0xff, 0xda, 0x00, 0x08,
	0x01, 0x01, 0x00, 0x00, 0x3f, 0x00, 0x7f, 0xff, 0xd9,
}

type Processor struct {
	ffmpegPath string
	mock       bool
}

type Result struct {
	LoudnessApplied bool
	ThumbnailFrom   string
}

func New(cfg config.Config) *Processor {
	path := cfg.FFmpegPath
	if path == "" {
		path = "ffmpeg"
	}
	return &Processor{ffmpegPath: path, mock: cfg.PostprocessMock}
}

func (p *Processor) Process(ctx context.Context, draftPath, finalPath, thumbPath string) (*Result, error) {
	log := applog.FromContext(ctx).With("service", "ffmpeg", "operation", "postprocess")
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return nil, err
	}
	if p.mock || isMockVideo(draftPath) || !p.ffmpegAvailable() {
		log.Info("ffmpeg postprocess mock", "draft", draftPath, "final", finalPath)
		return p.mockProcess(draftPath, finalPath, thumbPath)
	}
	started := time.Now()
	if err := p.normalize(ctx, draftPath, finalPath); err != nil {
		return nil, err
	}
	if err := p.extractThumbnail(ctx, finalPath, thumbPath); err != nil {
		return nil, err
	}
	log.Info("ffmpeg postprocess completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"draft", draftPath,
		"final", finalPath,
		"thumbnail", thumbPath,
	)
	return &Result{LoudnessApplied: true, ThumbnailFrom: "ffmpeg"}, nil
}

func (p *Processor) mockProcess(draftPath, finalPath, thumbPath string) (*Result, error) {
	data, err := os.ReadFile(draftPath)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(finalPath, data, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(thumbPath, mockJPEG, 0o644); err != nil {
		return nil, err
	}
	return &Result{LoudnessApplied: false, ThumbnailFrom: "mock"}, nil
}

func (p *Processor) ffmpegAvailable() bool {
	_, err := exec.LookPath(p.ffmpegPath)
	return err == nil
}

func (p *Processor) normalize(ctx context.Context, inPath, outPath string) error {
	log := applog.FromContext(ctx).With("service", "ffmpeg", "operation", "loudnorm")
	started := time.Now()
	log.Info("ffmpeg loudnorm started", "input", inPath, "output", outPath)
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-y", "-i", inPath,
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-preset", "fast", "-crf", "23",
		"-c:a", "aac", "-b:a", "192k",
		"-movflags", "+faststart",
		outPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("ffmpeg loudnorm failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"stderr_preview", applog.Truncate(string(out), 400),
		)
		return fmt.Errorf("ffmpeg loudnorm: %w: %s", err, string(out))
	}
	log.Info("ffmpeg loudnorm completed", "duration_ms", time.Since(started).Milliseconds())
	return nil
}

func (p *Processor) extractThumbnail(ctx context.Context, videoPath, thumbPath string) error {
	log := applog.FromContext(ctx).With("service", "ffmpeg", "operation", "thumbnail")
	started := time.Now()
	log.Info("ffmpeg thumbnail started", "input", videoPath, "output", thumbPath)
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-y", "-i", videoPath,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-q:v", "2",
		thumbPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("ffmpeg thumbnail failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"stderr_preview", applog.Truncate(string(out), 400),
		)
		return fmt.Errorf("ffmpeg thumbnail: %w: %s", err, string(out))
	}
	log.Info("ffmpeg thumbnail completed", "duration_ms", time.Since(started).Milliseconds())
	return nil
}

func isMockVideo(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 12 {
		return true
	}
	return string(data) == "MOCK_MP4_PHASE1" || len(data) < 32
}
