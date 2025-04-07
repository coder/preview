package types

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/terraform-provider-coder/v2/provider"
)

// @typescript-ignore BlockTypeParameter
// @typescript-ignore BlockTypeWorkspaceTag
const (
	BlockTypeParameter    = "coder_parameter"
	BlockTypeWorkspaceTag = "coder_workspace_tag"

	ValidationMonotonicIncreasing = "increasing"
	ValidationMonotonicDecreasing = "decreasing"
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
	ParameterData
	// Value is not immediately cast into a string.
	// Value is not required at template import, so defer
	// casting to a string until it is absolutely necessary.
	Value HCLString `json:"value"`

	// Diagnostics is used to store any errors that occur during parsing
	// of the parameter.
	Diagnostics Diagnostics `json:"diagnostics"`
}

type ParameterData struct {
	Name             string                     `json:"name"`
	DisplayName      string                     `json:"display_name"`
	Description      string                     `json:"description"`
	Type             ParameterType              `json:"type"`
	FormType         provider.ParameterFormType `json:"form_type"`
	FormTypeMetadata any                        `json:"form_type_metadata"`
	Mutable          bool                       `json:"mutable"`
	DefaultValue     HCLString                  `json:"default_value"`
	Icon             string                     `json:"icon"`
	Options          []*ParameterOption         `json:"options"`
	Validations      []*ParameterValidation     `json:"validations"`
	Required         bool                       `json:"required"`
	// legacy_variable_name was removed (= 14)
	Order     int64 `json:"order"`
	Ephemeral bool  `json:"ephemeral"`

	// Unexported fields, not always available.
	Source *terraform.Block `json:"-"`
}

type ParameterValidation struct {
	Error string `json:"validation_error"`

	// All validation attributes are optional.
	Regex     *string `json:"validation_regex"`
	Min       *int64  `json:"validation_min"`
	Max       *int64  `json:"validation_max"`
	Monotonic *string `json:"validation_monotonic"`
	Invalid   *bool   `json:"validation_invalid"`
}

// Valid takes the type of the value and the value itself and returns an error
// if the value is invalid.
func (v ParameterValidation) Valid(typ string, value string) error {
	// TODO: Validate typ is the enum?
	// Use the provider.Validation struct to validate the value to be
	// consistent with the provider.
	return (&provider.Validation{
		Min:         int(orZero(v.Min)),
		MinDisabled: v.Min == nil,
		Max:         int(orZero(v.Max)),
		MaxDisabled: v.Max == nil,
		Monotonic:   orZero(v.Monotonic),
		Regex:       orZero(v.Regex),
		Error:       v.Error,
	}).Valid(provider.OptionType(typ), value)
}

type ParameterOption struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Value       HCLString `json:"value"`
	Icon        string    `json:"icon"`
}

// CtyType returns the cty.Type for the ParameterData.
// A fixed set of types are supported.
func (r *ParameterData) CtyType() (cty.Type, error) {
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
		return cty.NilType, fmt.Errorf("unsupported type: %q", r.Type)
	}
}

func orZero[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
