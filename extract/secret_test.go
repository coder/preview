package extract_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview/extract"
)

// Test_SecretFromBlock_PanicRecover verifies that a panic inside
// SecretFromBlock is converted into an error diagnostic rather than crashing
// the whole extraction pass. A nil block triggers a nil pointer dereference
// inside requiredString, which the deferred recover should catch.
func Test_SecretFromBlock_PanicRecover(t *testing.T) {
	t.Parallel()

	req, diags := extract.SecretFromBlock(nil)
	require.Nil(t, req)
	require.True(t, diags.HasErrors(), "expected diagnostics; got %v", diags)

	var found bool
	for _, d := range diags {
		if d.Severity == hcl.DiagError && strings.Contains(d.Summary, "Panic occurred in extracting secret requirement") {
			found = true
			break
		}
	}
	require.True(t, found, "expected panic diagnostic; got: %v", diags)
}
