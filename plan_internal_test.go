package preview

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func Test_toCtyValue(t *testing.T) {
	t.Parallel()

	t.Run("EmptyList", func(t *testing.T) {
		t.Parallel()
		val, err := toCtyValue([]any{})
		require.NoError(t, err)
		require.True(t, val.Type().IsTupleType())
	})

	t.Run("HeterogeneousList", func(t *testing.T) {
		t.Parallel()
		val, err := toCtyValue([]any{5, "hello", true})
		require.NoError(t, err)
		require.True(t, val.Type().IsTupleType())
		require.Equal(t, 3, val.LengthInt())
		require.True(t, val.Equals(cty.TupleVal([]cty.Value{
			cty.NumberIntVal(5),
			cty.StringVal("hello"),
			cty.BoolVal(true),
		})).True())
	})
}
