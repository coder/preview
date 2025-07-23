package types

import (
	"github.com/zclconf/go-cty/cty"
)

type Variable struct {
	Name string `json:"name"`
	// Unsure how cty values json marshal
	Default     cty.Value `json:"default"`
	Type        cty.Type  `json:"type"`
	Description string    `json:"description"`
	Nullable    bool      `json:"nullable"`
	Sensitive   bool      `json:"sensitive"`

	// Variables also have 'Validation', which is currently not implemented.

	Diagnostics Diagnostics `json:"diagnostics"`
}
