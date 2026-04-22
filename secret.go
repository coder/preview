package preview

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func secrets(modules terraform.Modules) ([]types.SecretRequirement, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	reqs := make([]types.SecretRequirement, 0)
	// Track blocks by label (e.g. "x" in `data "coder_secret" "x"`) so we can
	// emit a single duplicate diagnostic per colliding label. Mirrors the
	// parameter dedup pattern in parameter.go.
	exists := make(map[string][]*terraform.Block)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypeSecret)
		for _, block := range blocks {
			req, rDiags := extract.SecretFromBlock(block)
			if len(rDiags) > 0 {
				diags = diags.Extend(rDiags)
			}
			if req != nil {
				reqs = append(reqs, *req)
				name := block.NameLabel()
				exists[name] = append(exists[name], block)
			}
		}
	}

	for name, blocks := range exists {
		if len(blocks) <= 1 {
			continue
		}
		var detail strings.Builder
		for _, b := range blocks {
			_, _ = detail.WriteString(fmt.Sprintf("block %q at %s\n",
				b.Type()+"."+strings.Join(b.Labels(), "."),
				b.HCLBlock().TypeRange))
		}
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Found %d duplicate coder_secret blocks with name %q, this is not allowed", len(blocks), name),
			Detail:   detail.String(),
		})
	}

	types.SortSecretRequirements(reqs)
	return reqs, diags
}
