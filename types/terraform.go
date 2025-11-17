package types

import (
	"github.com/zclconf/go-cty/cty"
)

type Variable struct {
	Name string `json:"name"`
	// JSON marshalling of cty types and values is not safe. Until these json values
	// are needed, serialization of them is not supported.
	// If serialization is supported, a custom marshaller will be needed.
	Default     cty.Value `json:"-"`
	Type        cty.Type  `json:"-"`
	Description string    `json:"description"`
	Nullable    bool      `json:"nullable"`
	Sensitive   bool      `json:"sensitive"`

	// Variables also have 'Validation', which is currently not implemented.

	Diagnostics Diagnostics `json:"diagnostics"`
}
