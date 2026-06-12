package image

import "context"

const (
	RolePersona = "persona"
	RoleProduct = "product"
	RoleScene   = "scene"
)

type ImageRequest struct {
	SceneID     string
	Prompt      string
	Role        string
	AspectRatio string
}

type GenerateResult struct {
	ImageURL string
	Data     []byte
}

type Provider interface {
	ID() string
	Generate(ctx context.Context, req ImageRequest) (*GenerateResult, error)
}

type GeneratedImage struct {
	SceneID   string `json:"sceneId"`
	Role      string `json:"role"`
	PublicURL string `json:"publicUrl"`
	FilePath  string `json:"filePath"`
	Source    string `json:"source"` // generated | user_asset
	Provider  string `json:"provider,omitempty"`
}

type StepOutput struct {
	Provider string           `json:"provider"`
	Skipped  bool             `json:"skipped"`
	Reason   string           `json:"reason,omitempty"`
	Images   []GeneratedImage `json:"images"`
}
