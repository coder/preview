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

// ValidatePrebuilds will iterate over each preset, validate the inputs as a set,
// and attach any diagnostics to the preset if there are issues. This must be done
// because prebuilds have to be valid without user input.
//
// This will only validate presets that have prebuilds configured and have no
// existing error diagnostics. This should only be used when doing Template
// Imports as a protection against invalid presets.
//
// A preset doesn't need to specify all required parameters, as users can provide
// the remaining values when creating a workspace. However, since prebuild
// creation has no user input, presets used for prebuilds must provide all
// required parameter values.
func ValidatePrebuilds(ctx context.Context, input Input, preValid []types.Preset, dir fs.FS) {
	for i := range preValid {
		pre := &preValid[i]
		if pre.Prebuilds == nil || pre.Prebuilds.Instances <= 0 {
			// No prebuilds, so presets do not need to be valid without user input
			continue
		}

		if hcl.Diagnostics(pre.Diagnostics).HasErrors() {
			// Piling on diagnostics is not helpful, if an error exists, leave it at that.
			continue
		}

		// Diagnostics are added to the existing preset.
		input.ParameterValues = pre.Parameters
		output, diagnostics := Preview(ctx, input, dir)
		if diagnostics.HasErrors() {
			pre.Diagnostics = append(pre.Diagnostics, diagnostics...)
			// Do not pile on more diagnostics for individual params, it already failed
			continue
		}

		if output == nil {
			continue
		}

		// If any parameter is invalid, then the preset is invalid.
		// A value must be specified for this failing parameter.
		for _, param := range output.Parameters {
			if hcl.Diagnostics(param.Diagnostics).HasErrors() {
				for _, paramDiag := range param.Diagnostics {
					if paramDiag.Severity != hcl.DiagError {
						continue // Only care about errors here
					}
					orig := paramDiag.Summary
					paramDiag.Summary = fmt.Sprintf("Parameter %s: %s", param.Name, orig)
					pre.Diagnostics = append(pre.Diagnostics, paramDiag)
				}
			}
		}
	}
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
	// preValidPresets are extracted as written in terraform. Each individual
	// param value is checked, however, the preset as a whole might not be valid.
	preValidPresets := presets(modules, rp)
	tags, tagDiags := workspaceTags(modules, p.Files())
	vars := variables(modules)

	// Add warnings
	diags = diags.Extend(warnings(modules))

	return &Output{
		ModuleOutput:  outputs,
		Parameters:    rp,
		WorkspaceTags: tags,
		Presets:       preValidPresets,
		Files:         p.Files(),
		Variables:     vars,
	}, diags.Extend(rpDiags).Extend(tagDiags)
}

func (i Input) RichParameterValue(key string) (string, bool) {
	p, ok := i.ParameterValues[key]
	return p, ok
}
