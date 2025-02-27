package extract

import (
	"fmt"
)

type stateParse struct {
	errors []error
	values map[string]any
}

func newStateParse(v map[string]any) *stateParse {
	return &stateParse{
		errors: make([]error, 0),
		values: v,
	}
}

func (p *stateParse) optionalString(key string) string {
	return optional[string](p.values, key)
}

func (p *stateParse) optionalInteger(key string) int64 {
	return optional[int64](p.values, key)
}

func (p *stateParse) optionalBool(key string) bool {
	return optional[bool](p.values, key)
}

func (p *stateParse) nullableInteger(key string) *int64 {
	if p.values[key] == nil {
		return nil
	}
	v := optional[int64](p.values, key)
	return &v
}

func (p *stateParse) nullableString(key string) *string {
	if p.values[key] == nil {
		return nil
	}
	v := optional[string](p.values, key)
	return &v
}

func (p *stateParse) string(key string) string {
	v, err := expected[string](p.values, key)
	if err != nil {
		p.errors = append(p.errors, err)
		return ""
	}
	return v
}

func optional[T any](vals map[string]any, key string) T {
	v, _ := expected[T](vals, key)
	return v
}

func expected[T any](vals map[string]any, key string) (T, error) {
	v, ok := vals[key]
	if !ok {
		return *new(T), fmt.Errorf("missing required key %q", key)
	}

	val, ok := v.(T)
	if !ok {
		return *new(T), fmt.Errorf("key %q is not of type %T", key, v)
	}
	return val, nil
}
