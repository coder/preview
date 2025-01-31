package preview

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraformplan/tfjson/parser"
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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
		for _, resource := range plan.PriorState.Values.RootModule.Resources {
			// TODO: Do index references exist here too?
			// TODO: Handle submodule nested resources

			parts := strings.Split(resource.Address, ".")
			if len(parts) < 2 {
				continue
			}

			if parts[0] == "data" && !strings.Contains(resource.Type, "coder") {
				continue
			}

			val, err := attributeCtyVal(resource.AttributeValues)
			if err != nil {
				// TODO: Remove log
				log.Printf("unable to determine value of resource %q: %v", resource.Address, err)
				continue
			}

			ctx.Set(val, parts...)
		}

	}, nil
}

func attributeCtyVal(attr map[string]interface{}) (cty.Value, error) {
	mv := make(map[string]cty.Value)
	for k, v := range attr {
		ty, err := gocty.ImpliedType(v)
		if err != nil {
			return cty.NilVal, fmt.Errorf("implied type for %q: %w", k, err)
		}

		mv[k], err = gocty.ToCtyValue(v, ty)
		if err != nil {
			return cty.NilVal, fmt.Errorf("implied value for %q: %w", k, err)
		}
	}

	return cty.ObjectVal(mv), nil
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
