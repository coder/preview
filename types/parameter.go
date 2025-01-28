package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type RichParameter struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"`
	Mutable      bool                   `json:"mutable"`
	DefaultValue string                 `json:"default_value"`
	Icon         string                 `json:"icon"`
	Options      []*RichParameterOption `json:"options"`
	Validation   *ParameterValidation   `json:"validation"`
	Required     bool                   `json:"required"`
	// legacy_variable_name was removed (= 14)
	DisplayName string `json:"display_name"`
	Order       int32  `json:"order"`
	Ephemeral   bool   `json:"ephemeral"`
}

type ParameterValidation struct {
	Regex     string `json:"validation_regex"`
	Error     string `json:"validation_error"`
	Min       *int32 `json:"validation_min"`
	Max       *int32 `json:"validation_max"`
	Monotonic string `json:"validation_monotonic"`
}

type RichParameterOption struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// Hash can be used to compare two RichParameter objects at a glance.
func (r *RichParameter) Hash() ([32]byte, error) {
	// Option order matters, so just json marshal the whole thing.
	data, err := json.Marshal(r)
	if err != nil {
		return [32]byte{}, fmt.Errorf("marshal: %w", err)
	}

	return sha256.Sum256(data), nil
}

// CtyType returns the cty.Type for the RichParameter.
// A fixed set of types are supported.
func (r *RichParameter) CtyType() (cty.Type, error) {
	switch r.Type {
	case "string":
		return cty.String, nil
	case "number":
		return cty.Number, nil
	case "bool":
		return cty.Bool, nil
	case "list(string)":
		return cty.List(cty.String), nil
	default:
		return cty.Type{}, fmt.Errorf("unsupported type: %q", r.Type)
	}
}
