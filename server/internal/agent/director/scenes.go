package director

const (
	VideoModeText2Video  = "text2video"
	VideoModeImage2Video = "image2video"
	VideoModeUserClip    = "user_clip"
)

func SceneMap(out *Output) map[string]SceneDirection {
	m := map[string]SceneDirection{}
	if out == nil {
		return m
	}
	for _, s := range out.Scenes {
		m[s.SceneID] = s
	}
	return m
}
