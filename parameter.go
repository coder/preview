package preview

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func RichParameters(modules terraform.Modules) ([]types.Parameter, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	params := make([]types.Parameter, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypeParameter)
		for _, block := range blocks {
			param, pDiags := extract.ParameterFromBlock(block)
			if len(pDiags) > 0 {
				diags = diags.Extend(pDiags)
			}

			if param != nil {
				params = append(params, *param)
			}
		}
	}

	return params, diags
}
