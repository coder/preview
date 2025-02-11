package preview

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"reflect"
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
		for _, block := range blocks {
			if block.InModule() {

				x := block.ModuleKey()
				y := block.ModuleBlock().FullName()
				var _, _ = x, y
				fmt.Println(block.ModuleKey())
				continue
			}

			err = loadResourcesToContext(block.Context().Parent(), plan.PriorState.Values.RootModule.Resources)
			if err != nil {
				panic(fmt.Sprintf("unable to load resources to context: %v", err))
			}
		}

		// 'data' blocks are loaded into prior state
		//plan.PriorState.Values.RootModule.Resources
		for _, resource := range plan.PriorState.Values.RootModule.Resources {
			// TODO: Do index references exist here too?
			// TODO: Handle submodule nested resources

			parts := strings.Split(resource.Address, ".")
			if len(parts) < 2 {
				continue
			}

			if parts[0] != "data" || strings.Contains(parts[1], "coder") {
				continue
			}

			val, err := toCtyValue(resource.AttributeValues)
			if err != nil {
				// TODO: Remove log
				log.Printf("unable to determine value of resource %q: %v", resource.Address, err)
				continue
			}

			ctx.Set(val, parts...)
		}

	}, nil
}

func planResources(plan *tfjson.Plan, block *terraform.Block) error {
	if !block.InModule() {
		return loadResourcesToContext(block.Context().Parent(), plan.PriorState.Values.RootModule.Resources)
	}

	var path []string
	mod := block.ModuleBlock()

	for {
		path = append([]string{mod.FullName()}, path...)
		break
	}
	return nil
}

func loadResourcesToContext(ctx *tfcontext.Context, resources []*tfjson.StateResource) error {
	for _, resource := range resources {
		if resource.Mode != "data" {
			continue
		}

		if strings.HasPrefix(resource.Type, "coder_") {
			// Ignore coder blocks
			continue
		}

		val, err := toCtyValue(resource.AttributeValues)
		if err != nil {
			return fmt.Errorf("unable to determine value of resource %q: %w", resource.Address, err)
		}

		ctx.Set(val, string(resource.Mode), resource.Type, resource.Name)
	}
	return nil
}

func toCtyValue(a any) (cty.Value, error) {
	if a == nil {
		return cty.NilVal, nil
	}
	av := reflect.ValueOf(a)
	switch av.Type().Kind() {
	case reflect.Slice, reflect.Array:
		sv := make([]cty.Value, 0, av.Len())
		for i := 0; i < av.Len(); i++ {
			v, err := toCtyValue(av.Index(i).Interface())
			if err != nil {
				return cty.NilVal, fmt.Errorf("slice value %d: %w", i, err)
			}
			sv = append(sv, v)
		}
		return cty.ListVal(sv), nil
	case reflect.Map:
		if av.Type().Key().Kind() != reflect.String {
			return cty.NilVal, fmt.Errorf("map keys must be string, found %q", av.Type().Key().Kind())
		}

		mv := make(map[string]cty.Value)
		var err error
		for _, k := range av.MapKeys() {
			v := av.MapIndex(k)
			mv[k.String()], err = toCtyValue(v.Interface())
			if err != nil {
				return cty.NilVal, fmt.Errorf("map value %q: %w", k.String(), err)
			}
		}
		return cty.ObjectVal(mv), nil
	default:
		ty, err := gocty.ImpliedType(a)
		if err != nil {
			return cty.NilVal, fmt.Errorf("implied type: %w", err)
		}

		cv, err := gocty.ToCtyValue(a, ty)
		if err != nil {
			return cty.NilVal, fmt.Errorf("implied value: %w", err)
		}
		return cv, nil
	}
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
