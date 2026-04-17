package extract

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/types"
)

// SecretFromBlock decodes a `data "coder_secret" {}` Terraform block into a
// SecretRequirement. Exactly one of `env` or `file` must be set, and
// `help_message` is required. Returns (nil, diags) on validation failure.
func SecretFromBlock(block *terraform.Block) (*types.SecretRequirement, hcl.Diagnostics) {
	// help_message is required AND must be a string; requiredString
	// handles both checks and emits a proper type diagnostic.
	var diags hcl.Diagnostics
	help, helpDiag := requiredString(block, "help_message")
	if helpDiag != nil {
		diags = diags.Append(helpDiag)
	}

	env := optionalString(block, "env")
	file := optionalString(block, "file")

	// Mutual exclusivity: exactly one of env/file must be set.
	switch {
	case env == "" && file == "":
		r := block.HCLBlock().Body.MissingItemRange()
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "coder_secret" block`,
			Detail:   `Exactly one of "env" or "file" must be set, neither were set`,
			Subject:  &r,
		})
	case env != "" && file != "":
		r := block.HCLBlock().Body.MissingItemRange()
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
