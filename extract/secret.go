package extract

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/types"
)

// SecretFromBlock decodes a `data "coder_secret" {}` Terraform block into a
// SecretRequirement. Exactly one of `env` or `file` must be set, and
// `help_message` is required. Returns (nil, diags) on validation failure.
func SecretFromBlock(block *terraform.Block) (req *types.SecretRequirement, diags hcl.Diagnostics) {
	defer func() {
		// Extra safety mechanism to ensure that if a panic occurs, we do not break
		// everything else.
		if r := recover(); r != nil {
			req = nil
			diags = hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Panic occurred in extracting secret requirement. This should not happen, please report this to Coder.",
					Detail:   fmt.Sprintf("panic in secret extract: %+v", r),
				},
			}
		}
	}()

	// help_message is required AND must be a string; requiredString
	// handles both checks and emits a proper type diagnostic.
	help, helpDiag := requiredString(block, "help_message")
	if helpDiag != nil {
		diags = diags.Append(helpDiag)
	}

	// Check presence separately from value so we can distinguish "attribute
	// absent" from "attribute present but wrong type"; the latter must produce
	// a type diagnostic instead of being silently treated as unset.
	envAttr := block.GetAttribute("env")
	fileAttr := block.GetAttribute("file")
	envSet := envAttr != nil && !envAttr.IsNil()
	fileSet := fileAttr != nil && !fileAttr.IsNil()

	var env, file string
	if envSet {
		v, d := requiredString(block, "env")
		if d != nil {
			diags = diags.Append(d)
		}
		env = v
	}
	if fileSet {
		v, d := requiredString(block, "file")
		if d != nil {
			diags = diags.Append(d)
		}
		file = v
	}

	// Mutual exclusivity is based on presence, not parsed value, so a
	// wrong-type attribute still counts as "set" here.
	switch {
	case !envSet && !fileSet:
		r := block.HCLBlock().DefRange
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "coder_secret" block`,
			Detail:   `Exactly one of "env" or "file" must be set, neither were set`,
			Subject:  &r,
		})
	case envSet && fileSet:
		r := block.HCLBlock().DefRange
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "coder_secret" block`,
			Detail:   `Exactly one of "env" or "file" must be set, both were set`,
			Subject:  &r,
		})
	}

	if diags.HasErrors() {
		return nil, diags
	}

	return &types.SecretRequirement{
		Env:         env,
		File:        file,
		HelpMessage: help,
	}, diags
}
