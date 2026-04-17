package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func secrets(modules terraform.Modules) ([]types.SecretRequirement, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	reqs := make([]types.SecretRequirement, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypeSecret)
		for _, block := range blocks {
			req, rDiags := extract.SecretFromBlock(block)
			if len(rDiags) > 0 {
				diags = diags.Extend(rDiags)
			}
			if req != nil {
				reqs = append(reqs, *req)
			}
		}
	}

	types.SortSecretRequirements(reqs)
	return reqs, diags
}
