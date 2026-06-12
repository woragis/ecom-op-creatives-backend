package video

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	imagemedia "github.com/woragis/ecom-op-creatives-backend/server/internal/media/image"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
)

type Service struct {
	registry     *Registry
	store        *storage.Local
	maxScenes    int
	pollInterval time.Duration
	maxPoll      time.Duration
	mockBytes          []byte
	apiPublicURL       string
	forceText2Video    bool
}

func NewService(cfg config.Config, registry *Registry, store *storage.Local) *Service {
	publicURL := strings.TrimSpace(cfg.APIPublicURL)
	if publicURL == "" {
		publicURL = strings.TrimSpace(os.Getenv("API_PUBLIC_URL"))
	}
	if publicURL == "" {
		publicURL = "http://localhost:8080"
	}
	return &Service{
		registry:     registry,
		store:        store,
		maxScenes:    cfg.VideoMaxScenes,
		pollInterval: time.Duration(cfg.VideoPollIntervalSec) * time.Second,
		maxPoll:      time.Duration(cfg.VideoMaxPollMin) * time.Minute,
		mockBytes:       []byte("MOCK_SCENE_MP4"),
		apiPublicURL:    strings.TrimRight(publicURL, "/"),
		forceText2Video: cfg.VideoForceText2Video,
	}
}

func (s *Service) GenerateScenes(
	ctx context.Context,
	providerID, runID string,
	prompter *prompteragent.Output,
	script *scriptwriter.Output,
	director *directoragent.Output,
	imageOut *imagemedia.StepOutput,
	assets *models.RunAssets,
) (*StepOutput, error) {
	provider, err := s.registry.Get(providerID)
	if err != nil {
		return &StepOutput{Provider: providerID, Skipped: true, Reason: err.Error()}, nil
	}

	prompts := map[string]string{}
	for _, p := range prompter.Scenes {
		prompts[p.SceneID] = p.VideoPrompt
	}
	dirMap := directoragent.SceneMap(director)

	var clips []Clip
	limit := s.maxScenes
	if limit <= 0 {
		limit = 3
	}

	for i, sc := range script.Scenes {
		if i >= limit {
			break
		}
		dir, _ := dirMap[sc.ID]
		mode := dir.VideoMode
		if mode == "" {
			mode = ModeText2Video
		}

		prompt := prompts[sc.ID]
		if prompt == "" {
			prompt = sc.Narration
		}

		imageURL := ""
		if mode == ModeImage2Video {
			role := dir.ImageRole
			if role == "" {
				role = imagemedia.RoleScene
			}
			imageURL = imagemedia.ImageForScene(imageOut, sc.ID, role)
			if imageURL == "" {
				imageURL = assetImageForRole(assets, role)
			}
			if imageURL == "" {
				mode = ModeText2Video
			}
		}
		mode, imageURL = s.resolveVideoMode(ctx, mode, imageURL, sc.ID)

		clip, err := s.generateClip(ctx, provider, providerID, runID, SceneRequest{
			SceneID:     sc.ID,
			Prompt:      prompt,
			DurationSec: clampDuration((sc.EndMs - sc.StartMs) / 1000),
			AspectRatio: "9:16",
			Mode:        mode,
			ImageURL:    s.resolveMediaURL(imageURL),
		})
		if err != nil {
			return nil, fmt.Errorf("scene %s: %w", sc.ID, err)
		}
		clip.Mode = mode
		clips = append(clips, *clip)
	}

	return &StepOutput{Provider: providerID, Clips: clips}, nil
}

func assetImageForRole(assets *models.RunAssets, role string) string {
	if assets == nil {
		return ""
	}
	switch role {
	case imagemedia.RolePersona:
		return assets.PersonaImage
	case imagemedia.RoleProduct:
		return assets.ProductImage
	default:
		return ""
	}
}

func (s *Service) resolveMediaURL(publicPath string) string {
	if publicPath == "" {
		return ""
	}
	if strings.HasPrefix(publicPath, "http://") || strings.HasPrefix(publicPath, "https://") {
		return publicPath
	}
	return s.apiPublicURL + publicPath
}

func (s *Service) resolveVideoMode(ctx context.Context, mode, imageURL, sceneID string) (string, string) {
	log := applog.FromContext(ctx).With("service", "video", "operation", "resolve_mode", "scene_id", sceneID)
	if s.forceText2Video && mode == ModeImage2Video {
		log.Info("forcing text2video", "reason", "VIDEO_FORCE_TEXT2VIDEO")
		return ModeText2Video, ""
	}
	if mode != ModeImage2Video || imageURL == "" {
		return mode, imageURL
	}
	resolved := s.resolveMediaURL(imageURL)
	if isExternalImageURL(resolved) {
		return mode, imageURL
	}
	log.Warn("falling back to text2video", "reason", "image_url_not_public_https", "image_url", resolved)
	return ModeText2Video, ""
}

func isExternalImageURL(url string) bool {
	if !strings.HasPrefix(url, "https://") {
		return false
	}
	host := strings.ToLower(url)
	return !strings.Contains(host, "://localhost") &&
		!strings.Contains(host, "://127.0.0.1") &&
		!strings.Contains(host, "://api:") &&
		!strings.Contains(host, "://host.docker.internal")
}

func (s *Service) generateClip(ctx context.Context, provider Provider, providerID, runID string, req SceneRequest) (*Clip, error) {
	log := applog.FromContext(ctx).With("service", providerID, "operation", "video.generate", "scene_id", req.SceneID, "mode", req.Mode)
	log.Info("video submit",
		"prompt_preview", applog.Truncate(req.Prompt, 160),
		"duration_sec", req.DurationSec,
		"image_url", req.ImageURL != "",
	)

	job, err := provider.Submit(ctx, req)
	if err != nil {
		log.Error("video submit failed", "error", err.Error())
		return nil, err
	}
	log.Info("video job submitted", "job_id", job.ID)

	videoURL, err := s.wait(ctx, provider, providerID, job.ID)
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

	log.Info("video job completed", "job_id", job.ID, "source_url", videoURL != "")

	return &Clip{
		SceneID:   req.SceneID,
		PublicURL: s.store.PublicPath(runID, filename),
		FilePath:  path,
		JobID:     job.ID,
		Provider:  providerID,
	}, nil
}

func (s *Service) wait(ctx context.Context, provider Provider, providerID, jobID string) (string, error) {
	log := applog.FromContext(ctx).With("service", providerID, "operation", "video.poll", "job_id", jobID)
	deadline := time.Now().Add(s.maxPoll)
	polls := 0
	for {
		if time.Now().After(deadline) {
			log.Error("video job timed out", "polls", polls)
			return "", fmt.Errorf("video job %s timed out", jobID)
		}
		result, err := provider.Poll(ctx, jobID)
		if err != nil {
			return "", err
		}
		polls++
		switch result.Status {
		case StatusCompleted:
			log.Info("video poll completed", "polls", polls)
			return result.VideoURL, nil
		case StatusFailed:
			log.Error("video poll failed", "polls", polls, "error", result.Error)
			return "", fmt.Errorf("video job failed: %s", result.Error)
		default:
			log.Debug("video poll running", "polls", polls, "status", result.Status)
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
