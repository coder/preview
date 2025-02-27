package types

import (
	"fmt"
	"slices"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// @typescript-ignore BlockTypeParameter
// @typescript-ignore BlockTypeWorkspaceTag
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

		return strings.Compare(a.Name, b.Name)
	})
}

type Parameter struct {
	RichParameter
	// Value is not immediately cast into a string.
	// Value is not required at template import, so defer
	// casting to a string until it is absolutely necessary.
	Value HCLString `json:"value"`

	// Diagnostics is used to store any errors that occur during parsing
	// of the parameter.
	Diagnostics Diagnostics `json:"diagnostics"`
}

type RichParameter struct {
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Description  string                 `json:"description"`
	FormControl  string                 `json:"form_control"`
	Type         ParameterType          `json:"type"`
	Mutable      bool                   `json:"mutable"`
	DefaultValue string                 `json:"default_value"`
	Icon         string                 `json:"icon"`
	Options      []*ParameterOption     `json:"options"`
	Validations  []*ParameterValidation `json:"validations"`
	Required     bool                   `json:"required"`
	// legacy_variable_name was removed (= 14)
	Order     int64 `json:"order"`
	Ephemeral bool  `json:"ephemeral"`
}

type ParameterValidation struct {
	Regex     *string `json:"validation_regex"`
	Error     string  `json:"validation_error"`
	Min       *int64  `json:"validation_min"`
	Max       *int64  `json:"validation_max"`
	Monotonic *string `json:"validation_monotonic"`
}

type ParameterOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Icon        string `json:"icon"`
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
