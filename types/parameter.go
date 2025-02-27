package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

const (
	BlockTypeParameter    = "coder_parameter"
	BlockTypeWorkspaceTag = "coder_workspace_tag"
)

func SortParameters(lists []Parameter) {
	slices.SortFunc(lists, func(a, b Parameter) int {
		order := int(a.Order - b.Order)
		if order != 0 {
			return order
		}

		if a.BlockName != b.BlockName {
			return strings.Compare(a.BlockName, b.BlockName)
		}

		return strings.Compare(a.Name, b.Name)
	})
}

type Parameter struct {
	// TODO: Might be value in a "Lazy" parameter
	Value ParameterValue
	RichParameter
}

type ParameterValue struct {
	// Value must be string due to the terraform type constraints
	Value string `json:"value"`
}

type RichParameter struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"`
	Mutable      bool                   `json:"mutable"`
	DefaultValue string                 `json:"default_value"`
	Icon         string                 `json:"icon"`
	Options      []*ParameterOption     `json:"options"`
	Validations  []*ParameterValidation `json:"validations"`
	Required     bool                   `json:"required"`
	// legacy_variable_name was removed (= 14)
	DisplayName string `json:"display_name"`
	Order       int64  `json:"order"`
	Ephemeral   bool   `json:"ephemeral"`

	// HCL props
	BlockName string `json:"block_name"`
}

type ParameterValidation struct {
	Regex     string `json:"validation_regex"`
	Error     string `json:"validation_error"`
	Min       *int64 `json:"validation_min"`
	Max       *int64 `json:"validation_max"`
	Monotonic string `json:"validation_monotonic"`
}

type ParameterOption struct {
	Name        string `json:"name" hcl:"name,attr"`
	Description string `json:"description" hcl:"description,attr"`
	Value       string `json:"value" hcl:"value,attr"`
	Icon        string `json:"icon" hcl:"icon,attr"`
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
