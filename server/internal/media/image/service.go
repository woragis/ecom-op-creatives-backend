package image

import (
	"context"
	"fmt"

	directoragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/director"
	prompteragent "github.com/woragis/ecom-op-creatives-backend/server/internal/agent/prompter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/agent/scriptwriter"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/media/storage"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/applog"
)

type Service struct {
	registry         *Registry
	store            *storage.Local
	maxScenes        int
	preferGenerated  bool
}

func NewService(cfg config.Config, registry *Registry, store *storage.Local) *Service {
	return &Service{
		registry:        registry,
		store:           store,
		maxScenes:       cfg.ImageMaxScenes,
		preferGenerated: cfg.ImagePreferGenerated,
	}
}

func (s *Service) GenerateScenes(
	ctx context.Context,
	providerID, runID string,
	prompter *prompteragent.Output,
	script *scriptwriter.Output,
	director *directoragent.Output,
	assets *models.RunAssets,
) (*StepOutput, error) {
	provider, err := s.registry.Get(providerID)
	if err != nil {
		return &StepOutput{Provider: providerID, Skipped: true, Reason: err.Error()}, nil
	}

	imagePrompts := map[string]prompteragent.ScenePrompt{}
	for _, p := range prompter.Scenes {
		imagePrompts[p.SceneID] = p
	}
	dirMap := directoragent.SceneMap(director)

	limit := s.maxScenes
	if limit <= 0 {
		limit = 3
	}

	var images []GeneratedImage
	for i, sc := range script.Scenes {
		if i >= limit {
			break
		}
		dir, _ := dirMap[sc.ID]
		role := dir.ImageRole
		if role == "" {
			role = RoleScene
		}

		if !s.preferGenerated {
			if url := assetURLForRole(assets, role); url != "" {
				applog.FromContext(ctx).With("step", "image", "scene_id", sc.ID, "role", role).Info(
					"reusing uploaded asset",
					"source", "user_asset",
					"public_url", url,
				)
				images = append(images, GeneratedImage{
					SceneID:   sc.ID,
					Role:      role,
					PublicURL: url,
					FilePath:  s.store.FilePath(runID, storage.PublicToRelPath(url)),
					Source:    "user_asset",
				})
				continue
			}
		}

		prompt := imagePrompts[sc.ID].ImagePrompt
		if prompt == "" {
			prompt = sc.Narration
		}
		result, err := provider.Generate(ctx, ImageRequest{
			SceneID:     sc.ID,
			Prompt:      prompt,
			Role:        role,
			AspectRatio: "9:16",
		})
		if err != nil {
			return nil, fmt.Errorf("scene %s: %w", sc.ID, err)
		}
		data := result.Data
		if len(data) == 0 && result.ImageURL != "" {
			data, err = Download(ctx, result.ImageURL)
			if err != nil {
				return nil, err
			}
		}
		filename := fmt.Sprintf("scene_%s_%s.png", sc.ID, role)
		path, err := s.store.WriteFile(runID, filename, data)
		if err != nil {
			return nil, err
		}
		images = append(images, GeneratedImage{
			SceneID:   sc.ID,
			Role:      role,
			PublicURL: s.store.PublicPath(runID, filename),
			FilePath:  path,
			Source:    "generated",
			Provider:  providerID,
		})
	}

	return &StepOutput{Provider: providerID, Images: images}, nil
}

func assetURLForRole(assets *models.RunAssets, role string) string {
	if assets == nil {
		return ""
	}
	switch role {
	case RolePersona:
		return assets.PersonaImage
	case RoleProduct:
		return assets.ProductImage
	default:
		return ""
	}
}

func ImagesBySceneRole(out *StepOutput) map[string]map[string]string {
	m := map[string]map[string]string{}
	if out == nil {
		return m
	}
	for _, img := range out.Images {
		if m[img.SceneID] == nil {
			m[img.SceneID] = map[string]string{}
		}
		m[img.SceneID][img.Role] = img.PublicURL
	}
	return m
}

func ImageForScene(out *StepOutput, sceneID, role string) string {
	if out == nil {
		return ""
	}
	for _, img := range out.Images {
		if img.SceneID == sceneID && (role == "" || img.Role == role) {
			return img.PublicURL
		}
	}
	return ""
}
