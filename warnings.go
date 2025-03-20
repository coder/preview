package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
)

func warnings(modules terraform.Modules) hcl.Diagnostics {
	var diags hcl.Diagnostics
	diags = diags.Extend(unexpandedCountBlocks(modules))

	return diags
}

// unexpandedCountBlocks is to compensate for a bug in the trivy parser.
// It is related to https://github.com/aquasecurity/trivy/pull/8479.
// Essentially, submodules are processed once. So if there is interdependent
// submodule references, then
func unexpandedCountBlocks(modules terraform.Modules) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, block := range modules.GetBlocks() {
		block := block

		if countAttr, ok := block.Attributes()["count"]; ok {
			if block.IsExpanded() {
				continue
			}

			diags = append(diags, &hcl.Diagnostic{
				Severity:    hcl.DiagWarning,
				Summary:     fmt.Sprintf("Unexpanded count argument on block %q", block.FullName()),
				Detail:      "The count argument is not expanded. This may lead to unexpected behavior. The default behavior is to assume count is 1.",
				Subject:     &countAttr.HCLAttribute().Range,
				Context:     &block.HCLBlock().DefRange,
				Expression:  countAttr.HCLAttribute().Expr,
				EvalContext: block.Context().Inner(),
			})
		}
	}
	return diags
}
