package verify

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview"
	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func Compare(t *testing.T, pr *preview.Output, values *tfjson.StateModule) {
	// Assert expected parameters
	stateParams, err := extract.ParametersFromState(values)
	require.NoError(t, err, "extract parameters from state")

	passed := assert.Equal(t, len(stateParams), len(pr.Parameters), "number of parameters")

	types.SortParameters(stateParams)
	types.SortParameters(pr.Parameters)
	for i, param := range stateParams {
		// TODO: A better compare function would be easier to debug
		assert.Equal(t, param, pr.Parameters[i], "parameter %q %d", param.BlockName, i)
	}

	if !passed {
		t.Fatalf("paramaters failed expectations")
	}
}
