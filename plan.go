package preview

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"slices"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraformplan/tfjson/parser"
	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/coder/preview/hclext"
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

	return func(ctx *tfcontext.Context, blocks terraform.Blocks, inputVars map[string]cty.Value) {
		loaded := make(map[*tfjson.StateModule]bool)

		// Do not recurse to child blocks.
		// TODO: Only load into the single parent context for the module.
		// And do not load context for a module more than once
		for _, block := range blocks {
			// TODO: Maybe switch to the 'configuration' block
			planMod := priorPlanModule(plan, block)
			if planMod == nil {
				continue
			}

			if loaded[planMod] {
				// No need to load this module into state again
				continue
			}

			rootCtx := block.Context()
			for {
				if rootCtx.Parent() != nil {
					rootCtx = rootCtx.Parent()
					continue
				}
				break
			}

			// Load state into the context
			err := loadResourcesToContext(rootCtx, planMod.Resources)
			if err != nil {
				// TODO: Somehow handle this error
				panic(fmt.Sprintf("unable to load resources to context: %v", err))
			}
			loaded[planMod] = true
		}
	}, nil
}

// priorPlanModule returns the state data of the module a given block is in.
func priorPlanModule(plan *tfjson.Plan, block *terraform.Block) *tfjson.StateModule {
	if !block.InModule() {
		return plan.PriorState.Values.RootModule
	}

	var modPath []string
	mod := block.ModuleBlock()
	for {
		modPath = append([]string{mod.LocalName()}, modPath...)
		mod = mod.ModuleBlock()
		if mod == nil {
			break
		}
	}

	current := plan.PriorState.Values.RootModule
	for i := range modPath {
		idx := slices.IndexFunc(current.ChildModules, func(m *tfjson.StateModule) bool {
			return m.Address == strings.Join(modPath[:i+1], ".")
		})
		if idx == -1 {
			// Maybe throw a diag here?
			return nil
		}

		current = current.ChildModules[idx]
	}

	return current
}

func matchingBlock(block *terraform.Block, planMod *tfjson.StateModule) *tfjson.StateResource {
	ref := block.Reference()
	matchKey := keyMatcher(ref.RawKey())

	for _, resource := range planMod.Resources {
		if ref.BlockType().ShortName() == string(resource.Mode) &&
			ref.TypeLabel() == resource.Type &&
			ref.NameLabel() == resource.Name &&
			matchKey(resource.Index) {

			return resource
		}
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

		path := []string{string(resource.Mode), resource.Type, resource.Name}

		// Always merge with any existing values
		existing := ctx.Get(path...)

		val, err := toCtyValue(resource.AttributeValues)
		if err != nil {
			return fmt.Errorf("unable to determine value of resource %q: %w", resource.Address, err)
		}

		var merged cty.Value
		switch resource.Index.(type) {
		case int, int32, int64, float32, float64:
			asInt, ok := toInt(resource.Index)
			if !ok {
				return fmt.Errorf("unable to convert index '%v' to int", resource.Index)
			}

			if !existing.Type().IsTupleType() {
				continue
			}
			merged = hclext.MergeWithTupleElement(existing, int(asInt), val)
		case nil:
			merged = hclext.MergeObjects(existing, val)
		default:
			return fmt.Errorf("unsupported index type %T", resource.Index)
		}

		ctx.Set(merged, string(resource.Mode), resource.Type, resource.Name)
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
	plan.FormatVersion = tfjson.PlanFormatVersionConstraints
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

func keyMatcher(key cty.Value) func(to any) bool {
	switch {
	case key.Type().Equals(cty.Number):
		idx, _ := key.AsBigFloat().Int64()
		return func(to any) bool {
			asInt, ok := toInt(to)
			return ok && asInt == idx
		}

	case key.Type().Equals(cty.String):
		// TODO: handle key strings
	}

	return func(to any) bool {
		return true
	}
}

func toInt(to any) (int64, bool) {
	switch typed := to.(type) {
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float32:
		return int64(typed), true
	case float64:
		return int64(typed), true
	}
	return 0, false
}
