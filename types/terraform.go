package types

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Variable struct {
	Name        string
	Default     cty.Value
	Type        cty.Type
	Description string
	Nullable    bool
	Sensitive   bool
	Ephemeral   bool

	// Variables also have 'Validation', which is currently not implemented.

	Diagnostics hcl.Diagnostics
}
