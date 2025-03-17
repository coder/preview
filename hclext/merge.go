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

func MergeWithTupleElement(list cty.Value, idx int, val cty.Value) cty.Value {
	if list.IsNull() ||
		!list.Type().IsTupleType() ||
		list.LengthInt() <= idx {
		return InsertTupleElement(list, idx, val)
	}

	existingElement := list.Index(cty.NumberIntVal(int64(idx)))
	merged := MergeObjects(existingElement, val)
	return InsertTupleElement(list, idx, merged)
}

// InsertTupleElement inserts a value into a tuple at the specified index.
// If the idx is outside the bounds of the list, it grows the tuple to
// the new size, and fills in `cty.NilVal` for the missing elements.
//
// This function will not panic. If the list value is not a list, it will
// be replaced with an empty list.
func InsertTupleElement(list cty.Value, idx int, val cty.Value) cty.Value {
	if list.IsNull() || !list.Type().IsTupleType() {
		// better than a panic
		list = cty.EmptyTupleVal
	}

	if idx < 0 {
		// Nothing to do?
		return list
	}

	newList := make([]cty.Value, max(idx+1, list.LengthInt()))
	for i := 0; i < len(newList); i++ {
		newList[i] = cty.NilVal // Always insert a nil by default

		if i < list.LengthInt() { // keep the original
			newList[i] = list.Index(cty.NumberIntVal(int64(i)))
		}

		if i == idx { // add the new value
			newList[i] = val
		}
	}

	return cty.TupleVal(newList)
}
