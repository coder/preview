package types

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
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
	Name         string                     `json:"name"`
	DisplayName  string                     `json:"display_name"`
	Description  string                     `json:"description"`
	Type         ParameterType              `json:"type"`
	FormType     provider.ParameterFormType `json:"form_type"`
	Styling      ParameterStyling           `json:"styling"`
	Mutable      bool                       `json:"mutable"`
	DefaultValue HCLString                  `json:"default_value"`
	Icon         string                     `json:"icon"`
	Options      []*ParameterOption         `json:"options"`
	Validations  []*ParameterValidation     `json:"validations"`
	Required     bool                       `json:"required"`
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
}

type ParameterStyling struct {
	Placeholder *string `json:"placeholder,omitempty"`
	Disabled    *bool   `json:"disabled,omitempty"`
	Label       *string `json:"label,omitempty"`
	MaskInput   *bool   `json:"mask_input,omitempty"`
}

type ParameterOption struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Value       HCLString `json:"value"`
	Icon        string    `json:"icon"`
}

func (r *ParameterData) Valid(value HCLString) hcl.Diagnostics {
	var defPtr *string

	if r.DefaultValue.Valid() && r.DefaultValue.IsKnown() {
		def := r.DefaultValue.AsString()
		defPtr = &def
	}

	var valuePtr *string
	// TODO: What to do if it is not valid?
	if value.Valid() {
		val := value.AsString()
		valuePtr = &val
	}

	_, diag := (&provider.Parameter{
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Type:        provider.OptionType(r.Type),
		FormType:    r.FormType,
		Mutable:     r.Mutable,
		Default:     defPtr,
		Icon:        r.Icon,
		Option:      providerOptions(r.Options),
		Validation:  providerValidations(r.Validations),
		Optional:    !r.Required,
		Order:       int(r.Order),
		Ephemeral:   r.Ephemeral,
	}).ValidateInput(valuePtr, nil) // TODO: Pass in previous value

	if diag.HasError() {
		// TODO: We can take the attr path and decorate the error with
		//   source information.
		return hclDiagnostics(diag, r.Source)
	}
	return nil
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
