package types

const (
	BlockTypePreset = "coder_workspace_preset"
)

type Preset struct {
	PresetData
	// Diagnostics is used to store any errors that occur during parsing
	// of the parameter.
	Diagnostics Diagnostics `json:"diagnostics"`
}

type PresetData struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"`
}
