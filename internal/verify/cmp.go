package verify

import (
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview"
	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func Compare(t *testing.T, pr *preview.Output, values *tfjson.StateModule) {
	passed := CompareParameters(t, pr, values)

	// TODO: Compare workspace tags

	if !passed {
		t.Fatalf("parameters failed expectations")
	}
}

func CompareParameters(t *testing.T, pr *preview.Output, values *tfjson.StateModule) bool {
	t.Helper()

	// Assert expected parameters
	stateParams, err := extract.ParametersFromState(values)
	require.NoError(t, err, "extract parameters from state")

	passed := assert.Equal(t, len(stateParams), len(pr.Parameters), "number of parameters")

	types.SortParameters(stateParams)
	types.SortParameters(pr.Parameters)
	for i, param := range stateParams {
		adata, err := json.Marshal(param)
		require.NoError(t, err, "marshal parameter %q", param.Name)

		bdata, err := json.Marshal(pr.Parameters[i])
		require.NoError(t, err, "marshal parameter %q", pr.Parameters[i].Name)

		require.JSONEq(t, string(adata), string(bdata), "parameter %q", param.Name)
	}

	return passed
}
