package types

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"golang.org/x/xerrors"
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

// TODO: Match implementation from https://github.com/coder/terraform-provider-coder/blob/main/provider/parameter.go#L404-L462
// TODO: Does the value have to be an option from the set of options?
func (v ParameterValidation) Valid(p string) error {
	validErr := xerrors.New(v.errorRendered(p))
	if v.Regex != nil {
		exp, err := regexp.Compile(*v.Regex)
		if err != nil {
			return fmt.Errorf("invalid regex %q: %w", *v.Regex, err)
		}

		if !exp.MatchString(p) {
			return validErr
		}
	}

	if v.Min != nil || v.Max != nil {
		vd, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid number value %q: %w", p, err)
		}

		if v.Min != nil && vd < *v.Min {
			return validErr
		}

		if v.Max != nil && vd > *v.Max {
			return validErr
		}
	}

	// Monotonic?

	return nil
}

func (v ParameterValidation) errorRendered(value string) string {
	r := strings.NewReplacer(
		"{min}", fmt.Sprintf("%d", safeDeref(v.Min)),
		"{max}", fmt.Sprintf("%d", safeDeref(v.Max)),
		"{value}", value,
	)
	return r.Replace(v.Error)
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
		return cty.NilType, fmt.Errorf("unsupported type: %q", r.Type)
	}
}

func safeDeref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
