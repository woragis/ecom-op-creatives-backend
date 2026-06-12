package video

import (
	"context"
	"fmt"
	"time"

	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
)

type Service struct {
	registry     *Registry
	store        *storage.Local
	maxScenes    int
	pollInterval time.Duration
	maxPoll      time.Duration
	mockBytes    []byte
}

func NewService(cfg config.Config, registry *Registry, store *storage.Local) *Service {
	return &Service{
		registry:     registry,
		store:        store,
		maxScenes:    cfg.VideoMaxScenes,
		pollInterval: time.Duration(cfg.VideoPollIntervalSec) * time.Second,
		maxPoll:      time.Duration(cfg.VideoMaxPollMin) * time.Minute,
		mockBytes:    []byte("MOCK_SCENE_MP4"),
	}
}

func (s *Service) GenerateScenes(
	ctx context.Context,
	providerID, runID string,
	prompter *prompteragent.Output,
	script *scriptwriter.Output,
) (*StepOutput, error) {
	provider, err := s.registry.Get(providerID)
	if err != nil {
		return &StepOutput{Provider: providerID, Skipped: true, Reason: err.Error()}, nil
	}

	prompts := map[string]string{}
	for _, p := range prompter.Scenes {
		prompts[p.SceneID] = p.VideoPrompt
	}

	var clips []Clip
	limit := s.maxScenes
	if limit <= 0 {
		limit = 3
	}

	for i, sc := range script.Scenes {
		if i >= limit {
			break
		}
		prompt := prompts[sc.ID]
		if prompt == "" {
			prompt = sc.Narration
		}
		clip, err := s.generateClip(ctx, provider, providerID, runID, SceneRequest{
			SceneID:     sc.ID,
			Prompt:      prompt,
			DurationSec: clampDuration((sc.EndMs - sc.StartMs) / 1000),
			AspectRatio: "9:16",
		})
		if err != nil {
			return nil, fmt.Errorf("scene %s: %w", sc.ID, err)
		}
		clips = append(clips, *clip)
	}

	return &StepOutput{Provider: providerID, Clips: clips}, nil
}

func (s *Service) generateClip(ctx context.Context, provider Provider, providerID, runID string, req SceneRequest) (*Clip, error) {
	job, err := provider.Submit(ctx, req)
	if err != nil {
		return nil, err
	}

	videoURL, err := s.wait(ctx, provider, job.ID)
	if err != nil {
		return nil, err
	}

	data, err := s.fetchVideo(ctx, videoURL)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("scene_%s.mp4", req.SceneID)
	path, err := s.store.WriteFile(runID, filename, data)
	if err != nil {
		return nil, err
	}

	return &Clip{
		SceneID:   req.SceneID,
		PublicURL: s.store.PublicPath(runID, filename),
		FilePath:  path,
		JobID:     job.ID,
		Provider:  providerID,
	}, nil
}

func (s *Service) wait(ctx context.Context, provider Provider, jobID string) (string, error) {
	deadline := time.Now().Add(s.maxPoll)
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("video job %s timed out", jobID)
		}
		result, err := provider.Poll(ctx, jobID)
		if err != nil {
			return "", err
		}
		switch result.Status {
		case StatusCompleted:
			return result.VideoURL, nil
		case StatusFailed:
			return "", fmt.Errorf("video job failed: %s", result.Error)
		default:
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(s.pollInterval):
			}
		}
	}
}

func (s *Service) fetchVideo(ctx context.Context, url string) ([]byte, error) {
	if url == "https://example.com/mock-video.mp4" {
		return s.mockBytes, nil
	}
	return Download(ctx, url)
}

func clampDuration(sec int) int {
	if sec < 3 {
		return 3
	}
	if sec > 10 {
		return 10
	}
	return sec
}

func ClipsBySceneID(out *StepOutput) map[string]string {
	m := map[string]string{}
	if out == nil {
		return m
	}
	for _, c := range out.Clips {
		m[c.SceneID] = c.PublicURL
	}
	return m
}
