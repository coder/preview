package extract

import (
	"encoding/json"
	"errors"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/coder/preview/types"
	"github.com/coder/terraform-provider-coder/v2/provider"
)

func ParametersFromState(state *tfjson.StateModule) ([]types.Parameter, error) {
	parameters := make([]types.Parameter, 0)
	for _, resource := range state.Resources {
		if resource.Mode != "data" {
			continue
		}

		if resource.Type != types.BlockTypeParameter {
			continue
		}

		param, err := ParameterFromState(resource)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, param)
	}

	for _, cm := range state.ChildModules {
		cParams, err := ParametersFromState(cm)
		if err != nil {
			return nil, fmt.Errorf("child module %q: %w", cm.Address, err)
		}
		parameters = append(parameters, cParams...)
	}

	return parameters, nil
}

func ParameterFromState(block *tfjson.StateResource) (types.Parameter, error) {
	st := newStateParse(block.AttributeValues)

	options, err := convertKeyList[*types.ParameterOption](st.values, "option", parameterOption)
	if err != nil {
		return types.Parameter{}, fmt.Errorf("convert param options: %w", err)
	}

	validations, err := convertKeyList(st.values, "validation", parameterValidation)
	if err != nil {
		return types.Parameter{}, fmt.Errorf("convert param validations: %w", err)
	}

	ftmeta := st.optionalString("form_type_metadata")
	var formTypeMeta any
	if ftmeta != "" {
		_ = json.Unmarshal([]byte(ftmeta), &formTypeMeta)
	}

	param := types.Parameter{
		Value: types.StringLiteral(st.string("value")),
		ParameterData: types.ParameterData{
			Name:             st.string("name"),
			Description:      st.optionalString("description"),
			Type:             types.ParameterType(st.optionalString("type")),
			FormType:         provider.ParameterFormType(st.optionalString("form_type")),
			FormTypeMetadata: formTypeMeta,
			Mutable:          st.optionalBool("mutable"),
			DefaultValue:     types.StringLiteral(st.optionalString("default")),
			Icon:             st.optionalString("icon"),
			Options:          options,
			Validations:      validations,
			Required:         st.optionalBool("required"),
			DisplayName:      st.optionalString("display_name"),
			Order:            st.optionalInteger("order"),
			Ephemeral:        st.optionalBool("ephemeral"),
		},
	}

	if len(st.errors) > 0 {
		return types.Parameter{}, errors.Join(st.errors...)
	}

	return param, nil
}

func convertKeyList[T any](vals map[string]any, key string, convert func(map[string]any) (T, error)) ([]T, error) {
	list := make([]T, 0)
	value, ok := vals[key]
	if !ok {
		return list, nil
	}

	if value == nil {
		return list, nil
	}

	elems, ok := value.([]any)
	if !ok {
		return list, fmt.Errorf("option is not a list, found %T", elems)
	}

	for _, elem := range elems {
		elemMap, ok := elem.(map[string]any)
		if !ok {
			return list, fmt.Errorf("option is not a map, found %T", elem)
		}

		converted, err := convert(elemMap)
		if err != nil {
			return list, fmt.Errorf("option: %w", err)
		}
		list = append(list, converted)
	}
	return list, nil
}

func parameterValidation(vals map[string]any) (*types.ParameterValidation, error) {
	st := newStateParse(vals)

	opt := types.ParameterValidation{
		Regex:     st.nullableString("regex"),
		Error:     st.optionalString("error"),
		Min:       st.nullableInteger("min"),
		Max:       st.nullableInteger("max"),
		Monotonic: st.nullableString("monotonic"),
	}

	// The state may have zero values with these booleans.
	if st.optionalBool("min_disabled") {
		opt.Min = nil
	}
	if st.optionalBool("max_disabled") {
		opt.Max = nil
	}

	// This is unfortunate, and unsure how to best resolve this.
	// Zero values are actual <nil>
	if opt.Regex != nil && *opt.Regex == "" {
		opt.Regex = nil
	}

	if opt.Monotonic != nil && *opt.Monotonic == "" {
		opt.Monotonic = nil
	}

	if len(st.errors) > 0 {
		return nil, errors.Join(st.errors...)
	}
	return &opt, nil
}

func parameterOption(vals map[string]any) (*types.ParameterOption, error) {
	st := newStateParse(vals)

	opt := types.ParameterOption{
		Name:        st.string("name"),
		Description: st.optionalString("description"),
		Value:       types.StringLiteral(st.string("value")),
		Icon:        st.optionalString("icon"),
	}

	if len(st.errors) > 0 {
		return nil, errors.Join(st.errors...)
	}
	return &opt, nil
}
