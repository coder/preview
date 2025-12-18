package extract

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/types"
)

func PresetFromBlock(block *terraform.Block) (tfPreset types.Preset) {
	defer func() {
		// Extra safety mechanism to ensure that if a panic occurs, we do not break
		// everything else.
		if r := recover(); r != nil {
			tfPreset = types.Preset{
				PresetData: types.PresetData{
					Name: block.Label(),
				},
				Diagnostics: types.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Panic occurred in extracting preset. This should not happen, please report this to Coder.",
						Detail:   fmt.Sprintf("panic in preset extract: %+v", r),
					},
				},
			}
		}
	}()

	p := types.Preset{
		PresetData: types.PresetData{
			Parameters: make(map[string]string),
		},
		Diagnostics: types.Diagnostics{},
	}

	if !block.IsResourceType(types.BlockTypePreset) {
		p.Diagnostics = append(p.Diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid Preset",
			Detail:   "Block is not a preset",
		})
		return p
	}

	pName, nameDiag := requiredString(block, "name")
	if nameDiag != nil {
		p.Diagnostics = append(p.Diagnostics, nameDiag)
	}
	p.Name = pName

	// GetAttribute and AsMapValue both gracefully handle `nil`, `null` and `unknown` values.
	// All of these return an empty map, which then makes the loop below a no-op.
	params := block.GetAttribute("parameters").AsMapValue()
	for presetParamName, presetParamValue := range params.Value() {
		p.Parameters[presetParamName] = presetParamValue
	}

	defaultAttr := block.GetAttribute("default")
	if defaultAttr != nil {
		p.Default = defaultAttr.Value().True()
	}

	prebuildBlock := block.GetBlock("prebuilds")
	if prebuildBlock != nil {
		p.Prebuilds = &types.PrebuildData{
			// Invalid values will be set to 0
			Instances: int(optionalInteger(prebuildBlock, "instances")),
		}
	}

	return p
}
