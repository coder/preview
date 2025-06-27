package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser/funcs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"golang.org/x/xerrors"
)

// init intends to override some of the default functions afforded by terraform.
// Specifically, any functions that require the context of the host.
//
// This is really unfortunate, but all the functions are globals, and this
// is the only way to override them.
func init() {
	// PathExpandFunc looks for references to a home directory on the host. The
	// preview rendering should not have access to the host's home directory path,
	// and will return an error if it is used.
	funcs.PathExpandFunc = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// This code is taken directly from https://github.com/mitchellh/go-homedir/blob/af06845cf3004701891bf4fdb884bfe4920b3727/homedir.go#L58
			// The only change is that instead of expanding the path, we return an error
			path := args[0].AsString()
			if len(path) == 0 {
				return cty.StringVal(path), nil
			}

			if path[0] != '~' {
				return cty.StringVal(path), nil
			}

			return cty.NilVal, xerrors.Errorf("not allowed to expand paths starting with '~' in this context")
		},
	})
}
