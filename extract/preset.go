package extract

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/coder/preview/types"
	"github.com/hashicorp/hcl/v2"
)

func PresetFromBlock(block *terraform.Block) types.Preset {
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

	return p
}
