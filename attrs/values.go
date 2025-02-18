package attrs

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Values interface {
	// Attr returns the value of the attribute with the given key.
	// If the attribute does not exist, it returns cty.NilVal.
	Attr(key string) cty.Value
}

func NewValues[T *terraform.Block | map[string]any](block T) Values {
	switch v := any(block).(type) {
	case *terraform.Block:
		return &blockValues{Values: v}
	case map[string]any:
		return &mapValues{Values: v}
	default:
		panic("unsupported type")
	}
}

type blockValues struct {
	Values *terraform.Block
}

func (b *blockValues) Attr(key string) cty.Value {
	attr := b.Values.GetAttribute(key)
	if attr.IsNil() {
		return cty.NilVal
	}
	return attr.Value()
}

type mapValues struct {
	Values map[string]any
}

func (m *mapValues) Attr(key string) cty.Value {
	val, ok := m.Values[key]
	if !ok {
		return cty.NilVal
	}

	gt, err := gocty.ImpliedType(val)
	if err != nil {
		return cty.NilVal
	}
	v, err := gocty.ToCtyValue(val, gt)
	if err != nil {
		return cty.NilVal
	}
	return v
}
