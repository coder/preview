package types_test

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview/types"
)

func TestDiagnosticExtra(t *testing.T) {
	diag := &hcl.Diagnostic{
		Severity: hcl.DiagWarning,
		Summary:  "Some summary",
		Detail:   "Some detail",
	}

	extra := types.ExtractDiagnosticExtra(diag)
	require.Empty(t, extra.Code)

	// Set it
	extra.Code = "foobar"
	types.SetDiagnosticExtra(diag, extra)

	extra = types.ExtractDiagnosticExtra(diag)
	require.Equal(t, "foobar", extra.Code)

	// Set it again
	extra.Code = "bazz"
	types.SetDiagnosticExtra(diag, extra)

	extra = types.ExtractDiagnosticExtra(diag)
	require.Equal(t, "bazz", extra.Code)
}

// TestDiagnosticExtraExisting is a test case where the DiagnosticExtra
// is already set in the wrapped chain of diagnostics.
// The `parent` wrapped is lost here, so calling `SetDiagnosticExtra` is
// lossy. In practice, we only call this once, so it's ok.
// TODO: Fix SetDiagnosticExtra to maintain the parents
// if the DiagnosticExtra already exists in the chain.
func TestDiagnosticExtraExisting(t *testing.T) {
	diag := &hcl.Diagnostic{
		Severity: hcl.DiagWarning,
		Summary:  "Some summary",
		Detail:   "Some detail",
		// parent -> existing -> child
		Extra: wrappedDiagnostic{
			name: "parent",
			wrapped: types.DiagnosticExtra{
				Code: "foobar",
				Wrapped: wrappedDiagnostic{
					name:    "child",
					wrapped: nil,
				},
			},
		},
	}

	extra := types.DiagnosticExtra{
		Code: "foo",
	}
	types.SetDiagnosticExtra(diag, extra)

	// The parent wrapped is lost
	isExtra, ok := diag.Extra.(types.DiagnosticExtra)
	require.True(t, ok)
	require.Equal(t, "foo", isExtra.Code)
	wrapped, ok := isExtra.UnwrapDiagnosticExtra().(wrappedDiagnostic)
	require.True(t, ok)
	require.Equal(t, wrapped.name, "child")
}

func TestDiagnosticsJSON(t *testing.T) {
	diags := types.Diagnostics{
		{
			Severity: hcl.DiagWarning,
			Summary:  "Some summary",
			Detail:   "Some detail",
			Extra: types.DiagnosticExtra{
				Code: "foobar",
			},
		},
		{
			Severity: hcl.DiagError,
			Summary:  "Some summary",
			Detail:   "Some detail",
			Extra: types.DiagnosticExtra{
				Code: "",
			},
		},
	}

	data, err := json.Marshal(diags)
	require.NoError(t, err, "marshal")

	var newDiags types.Diagnostics
	err = json.Unmarshal(data, &newDiags)
	require.NoError(t, err, "unmarshal")

	require.Equal(t, diags, newDiags)
}

type wrappedDiagnostic struct {
	name    string
	wrapped any
}

var _ hcl.DiagnosticExtraUnwrapper = wrappedDiagnostic{}

func (e wrappedDiagnostic) UnwrapDiagnosticExtra() interface{} {
	return e.wrapped
}
