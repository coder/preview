package preview

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/coder/preview/hclext"
	"github.com/coder/preview/tfvars"
	"github.com/coder/preview/types"
)

type Input struct {
	// PlanJSONPath is an optional path to a plan file. If PlanJSON isn't
	// specified, and PlanJSONPath is, then the file will be read and treated
	// as if the contents were passed in directly.
	PlanJSONPath    string
	PlanJSON        json.RawMessage
	ParameterValues map[string]string
	Owner           types.WorkspaceOwner
	Logger          *slog.Logger
	// TFVars will override any variables set in '.tfvars' files.
	// The value set must be a cty.Value, as the type can be anything.
	TFVars map[string]cty.Value
}

type Output struct {
	// ModuleOutput is any 'output' values from the terraform files. This has 0
	// effect on the parameters, tags, etc. It can be helpful for debugging, as it
	// allows exporting some terraform values to the caller to review.
	//
	// JSON marshalling is handled in the custom methods.
	ModuleOutput cty.Value `json:"-"`

	Parameters    []types.Parameter `json:"parameters"`
	WorkspaceTags types.TagBlocks   `json:"workspace_tags"`
	Presets       []types.Preset    `json:"presets"`
	Variables     []types.Variable  `json:"variables"`
	// Files is included for printing diagnostics.
	// They can be marshalled, but not unmarshalled. This is a limitation
	// of the HCL library.
	Files map[string]*hcl.File `json:"-"`
}

// MarshalJSON includes the ModuleOutput and files in the JSON output. Output
// should never be unmarshalled. Marshalling to JSON is strictly useful for
// debugging information.
func (o Output) MarshalJSON() ([]byte, error) {
	// Do not make this a fatal error, as it is supplementary information.
	modOutput, _ := ctyjson.Marshal(o.ModuleOutput, o.ModuleOutput.Type())

	type Alias Output
	return json.Marshal(&struct {
		ModuleOutput json.RawMessage      `json:"module_output"`
		Files        map[string]*hcl.File `json:"files"`
		Alias
	}{
		ModuleOutput: modOutput,
		Files:        o.Files,
		Alias:        (Alias)(o),
	})
}

func Preview(ctx context.Context, input Input, dir fs.FS) (output *Output, diagnostics hcl.Diagnostics) {
	// The trivy package works with `github.com/zclconf/go-cty`. This package is
	// similar to `reflect` in its usage. This package can panic if types are
	// misused. To protect the caller, a general `recover` is used to catch any
	// mistakes. If this happens, there is a developer bug that needs to be resolved.
	defer func() {
		if r := recover(); r != nil {
			diagnostics = hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Panic occurred in preview. This should not happen, please report this to Coder.",
					Detail:   fmt.Sprintf("panic in preview: %+v", r),
				},
			}
		}
	}()

	varFiles, err := tfvars.TFVarFiles("", dir)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Files not found",
				Detail:   err.Error(),
			},
		}
	}

	variableValues, err := tfvars.LoadTFVars(dir, varFiles)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to load tfvars from files",
				Detail:   err.Error(),
			},
		}
	}

	planHook, err := planJSONHook(dir, input)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Parsing plan JSON",
				Detail:   err.Error(),
			},
		}
	}

	ownerHook, err := workspaceOwnerHook(dir, input)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Workspace owner hook",
				Detail:   err.Error(),
			},
		}
	}

	logger := input.Logger
	if logger == nil { // Default to discarding logs
		logger = slog.New(slog.DiscardHandler)
	}

	// Override with user-supplied variables
	for k, v := range input.TFVars {
		variableValues[k] = v
	}

	// moduleSource is "" for a local module
	p := parser.New(dir, "",
		parser.OptionWithLogger(logger),
		parser.OptionStopOnHCLError(false),
		parser.OptionWithDownloads(false),
		parser.OptionWithSkipCachedModules(true),
		parser.OptionWithEvalHook(planHook),
		parser.OptionWithEvalHook(ownerHook),
		parser.OptionWithWorkingDirectoryPath("/"),
		parser.OptionWithEvalHook(parameterContextsEvalHook(input)),
		// 'OptionsWithTfVars' cannot be set with 'OptionWithTFVarsPaths'. So load the
		// tfvars from the files ourselves and merge with the user-supplied tf vars.
		parser.OptionsWithTfVars(variableValues),
	)

	err = p.ParseFS(ctx, ".")
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Parse terraform files",
				Detail:   err.Error(),
			},
		}
	}

	modules, err := p.EvaluateAll(ctx)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Evaluate terraform files",
				Detail:   err.Error(),
			},
		}
	}

	outputs := hclext.ExportOutputs(modules)

	diags := make(hcl.Diagnostics, 0)
	rp, rpDiags := parameters(modules)
	presets := presets(modules, rp)
	tags, tagDiags := workspaceTags(modules, p.Files())
	vars := variables(modules)

	// Add warnings
	diags = diags.Extend(warnings(modules))

	return &Output{
		ModuleOutput:  outputs,
		Parameters:    rp,
		WorkspaceTags: tags,
		Presets:       presets,
		Files:         p.Files(),
		Variables:     vars,
	}, diags.Extend(rpDiags).Extend(tagDiags)
}

func (i Input) RichParameterValue(key string) (string, bool) {
	p, ok := i.ParameterValues[key]
	return p, ok
}
