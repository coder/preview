package hclext

import "github.com/zclconf/go-cty/cty"

func MergeObjects(a, b cty.Value) cty.Value {
	output := make(map[string]cty.Value)

	for key, val := range a.AsValueMap() {
		output[key] = val
	}
	b.ForEachElement(func(key, val cty.Value) (stop bool) {
		k := key.AsString()
		old := output[k]
		if old.IsKnown() && isNotEmptyObject(old) && isNotEmptyObject(val) {
			output[k] = MergeObjects(old, val)
		} else {
			output[k] = val
		}
		return false
	})
	return cty.ObjectVal(output)
}

func isNotEmptyObject(val cty.Value) bool {
	return !val.IsNull() && val.IsKnown() && val.Type().IsObjectType() && val.LengthInt() > 0
}
