package preview

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraformplan/tfjson/parser"
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

func PlanJSONHook(dfs fs.FS, input Input) (func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value), error) {
	if input.PlanJSONPath == "" {
		return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {}, nil
	}
	file, err := dfs.Open(input.PlanJSONPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open plan JSON file: %w", err)
	}

	plan, err := ParsePlanJSON(file)
	if err != nil {
		return nil, fmt.Errorf("unable to parse plan JSON: %w", err)
	}

	//plan.PriorState

	var _ = plan
	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		// 'data' blocks are loaded into prior state
		//plan.PriorState.Values.RootModule.Resources

	}, nil
}

func extract(parents []string, mod *tfjson.StateModule) {

}

// ParsePlanJSON can parse the JSON output of a Terraform plan.
// terraform plan out.plan
// terraform show -json out.plan
func ParsePlanJSON(reader io.Reader) (*tfjson.Plan, error) {
	plan := new(tfjson.Plan)
	return plan, json.NewDecoder(reader).Decode(plan)
}

// ParsePlanJSON can parse the JSON output of a Terraform plan.
// terraform plan out.plan
// terraform show -json out.plan
func TrivyParsePlanJSON(reader io.Reader) (*tfjson.Plan, error) {
	p := parser.New()
	plan, err := p.Parse(reader)
	var _ = plan

	plan.ToFS()

	return nil, err
}
