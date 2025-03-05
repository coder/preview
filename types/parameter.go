package types

import (
	"encoding/json"
	"errors"
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
	Error string `json:"validation_error"`

	// All validation attributes are optional.
	Regex     *string `json:"validation_regex"`
	Min       *int64  `json:"validation_min"`
	Max       *int64  `json:"validation_max"`
	Monotonic *string `json:"validation_monotonic"`
}

// TODO: Match implementation from https://github.com/coder/terraform-provider-coder/blob/main/provider/parameter.go#L404-L462
// TODO: Does the value have to be an option from the set of options?

// Valid takes the type of the value and the value itself and returns an error
// if the value is invalid.
func (v ParameterValidation) Valid(typ, value string) error {
	if typ != "number" {
		var cannot []string
		if v.Min != nil {
			cannot = append(cannot, "min")
		}

		if v.Max != nil {
			cannot = append(cannot, "max")
		}

		if v.Monotonic != nil {
			cannot = append(cannot, "monotonic")
		}

		if len(cannot) == 1 {
			return fmt.Errorf("field %q is not supported for the type %q", cannot[0], typ)
		}

		if len(cannot) > 1 {
			return fmt.Errorf("fields [%s] are not supported for the type %q", strings.Join(cannot, ", "), typ)
		}
	}

	if typ != "string" && v.Regex != nil {
		return fmt.Errorf("field %q is not supported for the type %q", "regex", typ)
	}

	switch typ {
	case "bool":
		// Terraform boolean literals are "true" and "false" case-sensitive.
		// Do not allow alternate casing.
		if value != "true" && value != "false" {
			return fmt.Errorf(`boolean value can be either "true" or "false"`)
		}
		return nil
	case "string":
		if v.Regex == nil {
			return nil
		}

		exp, err := regexp.Compile(*v.Regex)
		if err != nil {
			return fmt.Errorf("compile regex %q: %s", *v.Regex, err)
		}

		if v.Error == "" {
			return fmt.Errorf("an error must be specified with a regex validation")
		}

		matched := exp.MatchString(value)
		if !matched {
			return fmt.Errorf("%s (value %q does not match %q)", v.errorRendered(value), value, exp)
		}
	case "number":
		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return takeFirstError(v.errorRendered(value), fmt.Errorf("value %q is not a number", value))
		}
		if v.Min != nil && num < *v.Min {
			return takeFirstError(v.errorRendered(value), fmt.Errorf("value %d is less than the minimum %d", num, v.Min))
		}
		if v.Max != nil && num > *v.Max {
			return takeFirstError(v.errorRendered(value), fmt.Errorf("value %d is more than the maximum %d", num, v.Max))
		}
		if v.Monotonic != nil && *v.Monotonic != ValidationMonotonicIncreasing && *v.Monotonic != ValidationMonotonicDecreasing {
			return fmt.Errorf("number monotonicity can be either %q or %q", ValidationMonotonicIncreasing, ValidationMonotonicDecreasing)
		}
	case "list(string)":
		var listOfStrings []string
		err := json.Unmarshal([]byte(value), &listOfStrings)
		if err != nil {
			return fmt.Errorf("value %q is not valid list of strings", value)
		}
	}

	return nil
}

func (v ParameterValidation) errorRendered(value string) error {
	r := strings.NewReplacer(
		"{min}", fmt.Sprintf("%d", safeDeref(v.Min)),
		"{max}", fmt.Sprintf("%d", safeDeref(v.Max)),
		"{value}", value,
	)
	return errors.New(r.Replace(v.Error))
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

func takeFirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return xerrors.Errorf("developer error: error message is not provided")
}
