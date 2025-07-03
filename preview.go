package preview

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/coder/preview/hclext"
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

	// TODO: Fix logging. There is no way to pass in an instanced logger to
	//   the parser.
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	// slog.SetDefault(slog.New(log.NewHandler(os.Stderr, nil)))

	varFiles, err := tfVarFiles("", dir)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Files not found",
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

	// moduleSource is "" for a local module
	p := parser.New(dir, "",
		parser.OptionWithLogger(logger),
		parser.OptionStopOnHCLError(false),
		parser.OptionWithDownloads(false),
		parser.OptionWithSkipCachedModules(true),
		parser.OptionWithTFVarsPaths(varFiles...),
		parser.OptionWithEvalHook(planHook),
		parser.OptionWithEvalHook(ownerHook),
		parser.OptionWithWorkingDirectoryPath("/"),
		parser.OptionWithEvalHook(parameterContextsEvalHook(input)),
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
	tags, tagDiags := workspaceTags(modules, p.Files())

	// Add warnings
	diags = diags.Extend(warnings(modules))

	return &Output{
		ModuleOutput:  outputs,
		Parameters:    rp,
		WorkspaceTags: tags,
		Files:         p.Files(),
	}, diags.Extend(rpDiags).Extend(tagDiags)
}

func (i Input) RichParameterValue(key string) (string, bool) {
	p, ok := i.ParameterValues[key]
	return p, ok
}

// tfVarFiles extracts any .tfvars files from the given directory.
// TODO: Test nested directories and how that should behave.
func tfVarFiles(path string, dir fs.FS) ([]string, error) {
	dp := "."
	entries, err := fs.ReadDir(dir, dp)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", dp, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			subD, err := fs.Sub(dir, entry.Name())
			if err != nil {
				return nil, fmt.Errorf("sub dir %q: %w", entry.Name(), err)
			}
			newFiles, err := tfVarFiles(filepath.Join(path, entry.Name()), subD)
			if err != nil {
				return nil, err
			}
			files = append(files, newFiles...)
		}

		if filepath.Ext(entry.Name()) == ".tfvars" {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}
