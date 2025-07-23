package preview

import (
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func variables(modules terraform.Modules) []types.Variable {
	variableBlocks := modules.GetDatasByType("variable")
	vars := make([]types.Variable, 0, len(variableBlocks))
	for _, block := range variableBlocks {
		vars = append(vars, extract.VariableFromBlock(block))
	}

	// Sort the variables by name for consistency
	slices.SortFunc(vars, func(a, b types.Variable) int {
		return strings.Compare(a.Name, b.Name)
	})
	return vars
}
