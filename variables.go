package preview

import (
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"

	"github.com/coder/preview/extract"
	"github.com/coder/preview/types"
)

func variables(modules terraform.Modules) []types.Variable {
	vars := make([]types.Variable, 0)

	for _, mod := range modules {
		// Only extract variables from root modules. Child modules have their
		// vars set in the parent module.
		if mod.Parent() == nil {
			variableBlocks := mod.GetBlocks().OfType("variable")
			for _, block := range variableBlocks {
				vars = append(vars, extract.VariableFromBlock(block))
			}
		}
	}

	// Sort the variables by name for consistency
	slices.SortFunc(vars, func(a, b types.Variable) int {
		return strings.Compare(a.Name, b.Name)
	})
	return vars
}
