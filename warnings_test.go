package preview_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview"
	"github.com/coder/preview/types"
)

// Test_UnresolvedModules asserts that the "Module not loaded" warning tracks
// whether a module actually loaded.
func Test_UnresolvedModules(t *testing.T) {
	t.Parallel()

	t.Run("LateExpansionIsNotAFailure", func(t *testing.T) {
		t.Parallel()

		// Modules gated on another module's output are expanded after their own
		// blocks are parsed, which clones the module block. The clone must not be
		// mistaken for a module that failed to load.
		output, diags := previewTemplate(t, "lateexpansion")
		require.Empty(t, moduleNotLoaded(diags),
			"modules that loaded must not be reported as unresolved")

		// The parameter is declared by the `count` expanded module, so rendering
		// it proves the fixture still exercises the expansion path rather than
		// passing because nothing expanded.
		require.Len(t, output.Parameters, 1)
		require.Equal(t, "ai_model", output.Parameters[0].Name)
	})

	t.Run("MissingModuleStillWarns", func(t *testing.T) {
		t.Parallel()

		// The fixture declares one module that cannot be resolved, one that
		// resolves locally, and one expanded to zero instances. Only the first
		// may be reported, so this fails if the identity change either silenced
		// a real failure or started flagging a module that loaded.
		_, diags := previewTemplate(t, "missingmodule")
		details := moduleNotLoaded(diags)
		require.Len(t, details, 1, "exactly one module cannot be resolved")
		require.Contains(t, details[0], `module "does-not-exist"`)
	})
}

func previewTemplate(t *testing.T, dir string) (*preview.Output, hcl.Diagnostics) {
	t.Helper()

	output, diags := preview.Preview(context.Background(), preview.Input{},
		os.DirFS(filepath.Join("testdata", dir)))
	require.False(t, diags.HasErrors(), "unexpected errors: %s", diags)
	require.NotNil(t, output)

	return output, diags
}

func moduleNotLoaded(diags hcl.Diagnostics) []string {
	var details []string
	for _, diag := range diags {
		if types.ExtractDiagnosticExtra(diag).Code == types.DiagnosticModuleNotLoaded {
			details = append(details, diag.Detail)
		}
	}
	return details
}
