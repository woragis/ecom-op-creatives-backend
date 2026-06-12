package models

import "encoding/json"

type RunAssets struct {
	PersonaImage string `json:"personaImage,omitempty"`
	ProductImage string `json:"productImage,omitempty"`
	IntroClip    string `json:"introClip,omitempty"`
}

func ParseRunAssets(raw json.RawMessage) *RunAssets {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return &RunAssets{}
	}
	var assets RunAssets
	if err := json.Unmarshal(raw, &assets); err != nil {
		return &RunAssets{}
	}
	return &assets
}

func (a *RunAssets) JSON() json.RawMessage {
	if a == nil {
		return json.RawMessage(`{}`)
	}
	b, err := json.Marshal(a)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}
