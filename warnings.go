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
	modulesUsed := make(map[moduleIdentity]bool)
	modulesByIdentity := make(map[moduleIdentity]*terraform.Block)

	// There is no easy way to know if a `module` failed to resolve. The failure is
	// only logged in the trivy package. No errors are returned to the caller. So
	// instead this code will infer a failed resolution by checking if any blocks
	// exist that reference each `module` block. This will work as long as the module
	// has some content. If a module is completely empty, then it will be detected as
	// "not loaded".
	blocks := modules.GetBlocks()
	for _, block := range blocks {
		if block.InModule() && block.ModuleBlock() != nil {
			modulesUsed[identifyModule(block.ModuleBlock())] = true
		}

		if block.Type() == "module" {
			identity := identifyModule(block)
			modulesByIdentity[identity] = block
			_, ok := modulesUsed[identity]
			if !ok {
				modulesUsed[identity] = false
			}
		}
	}

	for identity, v := range modulesUsed {
		if !v {
			block, ok := modulesByIdentity[identity]
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

// moduleIdentity identifies a `module` block by where it is declared rather
// than by the block instance.
//
// terraform.Block.ID() is a per-instance UUID that Clone() regenerates, and a
// module block is cloned whenever `count` or `for_each` is expanded. Expansion
// can happen after the module's own blocks captured their ModuleBlock()
// pointer, in which case the enumerated block and the captured block are
// different instances with different IDs, and a module that loaded correctly
// is reported as not loaded.
//
// Both fields survive Clone(): the block reference is built from the original
// labels before Clone() rewrites them, and copyBlock() shallow copies the
// hcl.Block, preserving DefRange. Keeping them as separate fields rather than a
// formatted string means two declarations can never be conflated by the way
// their parts are joined.
//
// Instances of the same expanded module block share an identity, so a module is
// considered loaded if any of its instances contributed blocks.
type moduleIdentity struct {
	// key is the module address terraform uses in modules.json, without an
	// instance index, for example "workspace.claude_code".
	key string
	// declaration is the range of the `module` block in its source file, which
	// distinguishes declarations that share a key.
	declaration hcl.Range
}

func identifyModule(block *terraform.Block) moduleIdentity {
	return moduleIdentity{
		key:         block.ModuleKey(),
		declaration: block.HCLBlock().DefRange,
	}
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
