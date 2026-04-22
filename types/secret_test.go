package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/preview/types"
)

// Test_SecretRequirement_JSON_Omitempty verifies that empty Env, File, and
// HelpMessage fields are omitted from the marshaled JSON. Since extraction
// enforces exactly one of Env/File is non-empty, omitempty makes the on-the-
// wire shape explicit about which field applies.
func Test_SecretRequirement_JSON_Omitempty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  types.SecretRequirement
		want string
	}{
		{
			name: "env only",
			req:  types.SecretRequirement{Env: "FOO", HelpMessage: "set FOO"},
			want: `{"env":"FOO","help_message":"set FOO"}`,
		},
		{
			name: "file only",
			req:  types.SecretRequirement{File: "~/bar", HelpMessage: "set bar"},
			want: `{"file":"~/bar","help_message":"set bar"}`,
		},
		{
			name: "empty help_message is omitted",
			req:  types.SecretRequirement{Env: "FOO"},
			want: `{"env":"FOO"}`,
		},
		{
			name: "all empty produces empty object",
			req:  types.SecretRequirement{},
			want: `{}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := json.Marshal(tc.req)
			require.NoError(t, err)
			require.JSONEq(t, tc.want, string(got))
		})
	}
}
