package types

import (
	"slices"
	"strings"
)

// @typescript-ignore BlockTypeSecret
const BlockTypeSecret = "coder_secret"

// SecretRequirement describes a `data "coder_secret"` block declared in a
// template. Exactly one of Env or File will be non-empty; validation of that
// invariant happens during extraction.
type SecretRequirement struct {
	Env         string `json:"env,omitempty"`
	File        string `json:"file,omitempty"`
	HelpMessage string `json:"help_message,omitempty"`
}

// SortSecretRequirements orders requirements first by Env then by File so
// diagnostic output is stable across runs.
func SortSecretRequirements(reqs []SecretRequirement) {
	slices.SortFunc(reqs, func(a, b SecretRequirement) int {
		if c := strings.Compare(a.Env, b.Env); c != 0 {
			return c
		}
		return strings.Compare(a.File, b.File)
	})
}
