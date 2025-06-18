package extract

import (
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/coder/preview/types"
	"github.com/hashicorp/hcl/v2"
)

func PresetFromBlock(block *terraform.Block) (*types.Preset, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	pName, nameDiag := requiredString(block, "name")
	if nameDiag != nil {
		diags = append(diags, nameDiag)
	}

	p := types.Preset{
		PresetData: types.PresetData{
			Name:       pName,
			Parameters: make(map[string]string),
		},
		Diagnostics: types.Diagnostics{},
	}

	params := block.GetAttribute("parameters").AsMapValue()
	for presetParamName, presetParamValue := range params.Value() {
		p.Parameters[presetParamName] = presetParamValue
	}

	return &p, diags
}
