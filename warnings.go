package preview

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/types"
)

func warnings(modules terraform.Modules) hcl.Diagnostics {
	var diags hcl.Diagnostics
	diags = diags.Extend(unexpandedCountBlocks(modules))
	diags = diags.Extend(unresolvedModules(modules))

	return diags
}

// unresolvedModules does a best effort to try and detect if some modules
// failed to resolve. This is usually because `terraform init` is not run.
func unresolvedModules(modules terraform.Modules) hcl.Diagnostics {
	var diags hcl.Diagnostics
	modulesUsed := make(map[string]bool)
	modulesByID := make(map[string]*terraform.Block)

	// There is no easy way to know if a `module` failed to resolve. The failure is
	// only logged in the trivy package. No errors are returned to the caller. So
	// instead this code will infer a failed resolution by checking if any blocks
	// exist that reference each `module` block. This will work as long as the module
	// has some content. If a module is completely empty, then it will be detected as
	// "not loaded".
	blocks := modules.GetBlocks()
	for _, block := range blocks {
		if block.InModule() && block.ModuleBlock() != nil {
			modulesUsed[block.ModuleBlock().ID()] = true
		}

		if block.Type() == "module" {
			modulesByID[block.ID()] = block
			_, ok := modulesUsed[block.ID()]
			if !ok {
				modulesUsed[block.ID()] = false
			}
		}
	}

	for id, v := range modulesUsed {
		if !v {
			block, ok := modulesByID[id]
			if ok {
				label := block.Type()
				for _, l := range block.Labels() {
					label += " " + fmt.Sprintf("%q", l)
				}

				diags = diags.Append(types.DiagnosticCode(&hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "Module not loaded. Did you run `terraform init`?",
					Detail:   fmt.Sprintf("Module '%s' in file %q cannot be resolved. This module will be ignored.", label, block.HCLBlock().DefRange),
					Subject:  &(block.HCLBlock().DefRange),
				}, types.DiagnosticModuleNotLoaded))
			}
		}
	}

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

		// Only warn on coder blocks
		if !strings.HasPrefix(block.NameLabel(), "coder_") {
			continue
		}
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
