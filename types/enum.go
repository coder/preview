package types

import (
	"fmt"
	"strings"
)

type ParameterType string

const (
	ParameterTypeString     ParameterType = "string"
	ParameterTypeNumber     ParameterType = "number"
	ParameterTypeBoolean    ParameterType = "boolean"
	ParameterTypeListString ParameterType = "list(string)"
)

func (t ParameterType) Valid() error {
	switch t {
	case ParameterTypeString, ParameterTypeNumber, ParameterTypeBoolean, ParameterTypeListString:
		return nil
	default:
		return fmt.Errorf("invalid parameter type %q, expected one of [%s]", t,
			strings.Join([]string{
				string(ParameterTypeString),
				string(ParameterTypeNumber),
				string(ParameterTypeBoolean),
				string(ParameterTypeListString),
			}, ", "))
	}
}
