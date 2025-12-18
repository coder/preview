package types

const (
	BlockTypePreset = "coder_workspace_preset"
)

type Preset struct {
	PresetData
	// Diagnostics is used to store any errors that occur during parsing
	// of the preset.
	Diagnostics Diagnostics `json:"diagnostics"`
}

type PrebuildData struct {
	Instances int `json:"instances"`
}

type PresetData struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"`
	Default    bool              `json:"default"`
	Prebuilds  *PrebuildData     `json:"prebuilds,omitempty"`
}
