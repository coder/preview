package extract

import (
	"errors"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/types"
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

	param := types.Parameter{
		Value: types.ParameterValue{
			Value: cty.StringVal(st.string("value")),
		},
		RichParameter: types.RichParameter{
			Name:         st.string("name"),
			Description:  st.optionalString("description"),
			Type:         st.string("type"),
			Mutable:      st.optionalBool("mutable"),
			DefaultValue: st.optionalString("default"),
			Icon:         st.optionalString("icon"),
			Options:      options,
			Validations:  validations,
			Required:     st.optionalBool("mutable"),
			DisplayName:  st.optionalString("display_name"),
			Order:        st.optionalInteger("order"),
			Ephemeral:    st.optionalBool("mutable"),
			BlockName:    block.Name,
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
		Regex:     st.optionalString("regex"),
		Error:     st.optionalString("error"),
		Min:       st.nullableInteger("min"),
		Max:       st.nullableInteger("max"),
		Monotonic: st.optionalString("monotonic"),
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
		Value:       st.string("value"),
		Icon:        st.optionalString("icon"),
	}

	if len(st.errors) > 0 {
		return nil, errors.Join(st.errors...)
	}
	return &opt, nil
}
