package preview

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func RichParameters(modules terraform.Modules) ([]types.Parameter, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	params := make([]types.Parameter, 0)
	exists := make(map[string][]types.Parameter)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypeParameter)
		for _, block := range blocks {
			param, pDiags := extract.ParameterFromBlock(block)
			if len(pDiags) > 0 {
				diags = diags.Extend(pDiags)
			}

			if param != nil {
				params = append(params, *param)

				if _, ok := exists[param.Name]; !ok {
					exists[param.Name] = make([]types.Parameter, 0)
				}
				exists[param.Name] = append(exists[param.Name], *param)
			}
		}
	}

	for k, v := range exists {
		var detail strings.Builder
		for _, p := range v {
			if p.Source != nil {
				detail.WriteString(fmt.Sprintf("block %q at %s\n",
					p.Source.Type()+"."+strings.Join(p.Source.Labels(), "."),
					p.Source.HCLBlock().TypeRange))
			}
		}
		if len(v) > 1 {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Found %d duplicate parameters with name %q, this is not allowed", len(v), k),
				Detail:   detail.String(),
			})
		}
	}

	types.SortParameters(params)
	return params, diags
}
