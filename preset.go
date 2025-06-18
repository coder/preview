package preview

import (
	"fmt"
	"slices"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func presets(modules terraform.Modules, parameters []types.Parameter) ([]types.Preset, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	presets := make([]types.Preset, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypePreset)
		for _, block := range blocks {
			preset, pDiags := extract.PresetFromBlock(block)
			if len(pDiags) > 0 {
				diags = diags.Extend(pDiags)
			}

			if preset == nil {
				continue
			}

			for paramName, paramValue := range preset.Parameters {
				templateParamIndex := slices.IndexFunc(parameters, func(p types.Parameter) bool {
					return p.Name == paramName
				})
				if templateParamIndex == -1 {
					preset.Diagnostics = append(preset.Diagnostics, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Undefined Parameter",
						Detail:   fmt.Sprintf("Preset %q requires parameter %q, but it is not defined by the template.", preset.Name, paramName),
					})
					continue
				}
				templateParam := parameters[templateParamIndex]
				for _, diag := range templateParam.Valid(types.StringLiteral(paramValue)) {
					preset.Diagnostics = append(preset.Diagnostics, diag)
				}
			}

			presets = append(presets, *preset)
		}
	}

	return presets, diags
}
