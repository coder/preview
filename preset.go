package preview

import (
	"fmt"
	"slices"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

// presets extracts all presets from the given modules. It then validates the name,
// parameters and default preset.
func presets(modules terraform.Modules, parameters []types.Parameter) []types.Preset {
	foundPresets := make([]types.Preset, 0)
	var defaultPreset *types.Preset

	for _, mod := range modules {
		blocks := mod.GetDatasByType(types.BlockTypePreset)
		for _, block := range blocks {
			preset := extract.PresetFromBlock(block)
			switch true {
			case defaultPreset != nil && preset.Default:
				preset.Diagnostics = append(preset.Diagnostics, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Multiple default presets",
					Detail:   fmt.Sprintf("Only one preset can be marked as default. %q is already marked as default", defaultPreset.Name),
				})
			case defaultPreset == nil && preset.Default:
				defaultPreset = &preset
			}

			for paramName, paramValue := range preset.Parameters {
				templateParamIndex := slices.IndexFunc(parameters, func(p types.Parameter) bool {
					return p.Name == paramName
				})
				if templateParamIndex == -1 {
					preset.Diagnostics = append(preset.Diagnostics, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Undefined Parameter",
						Detail:   fmt.Sprintf("Preset parameter %q is not defined by the template.", paramName),
					})
					continue
				}
				templateParam := parameters[templateParamIndex]
				for _, diag := range templateParam.Valid(types.StringLiteral(paramValue)) {
					preset.Diagnostics = append(preset.Diagnostics, diag)
				}
			}

			foundPresets = append(foundPresets, preset)
		}
	}

	return foundPresets
}
