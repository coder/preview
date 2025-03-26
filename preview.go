package preview

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser"
	"github.com/aquasecurity/trivy/pkg/log"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

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
}

type Output struct {
	ModuleOutput  cty.Value
	Parameters    []types.Parameter
	WorkspaceTags types.TagBlocks
	Files         map[string]*hcl.File
}

func Preview(ctx context.Context, input Input, dir fs.FS) (*Output, hcl.Diagnostics) {
	// TODO: FIX LOGGING
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.SetDefault(slog.New(log.NewHandler(os.Stderr, nil)))

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

	planHook, err := PlanJSONHook(dir, input)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Parsing plan JSON",
				Detail:   err.Error(),
			},
		}
	}

	ownerHook, err := WorkspaceOwnerHook(dir, input)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Workspace owner hook",
				Detail:   err.Error(),
			},
		}
	}
	var _ = ownerHook

	// moduleSource is "" for a local module
	p := parser.New(dir, "",
		parser.OptionStopOnHCLError(false),
		parser.OptionWithDownloads(false),
		parser.OptionWithSkipCachedModules(true),
		parser.OptionWithTFVarsPaths(varFiles...),
		parser.OptionWithEvalHook(planHook),
		parser.OptionWithEvalHook(ownerHook),
		parser.OptionWithEvalHook(ParameterContextsEvalHook(input)),
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

	modules, outputs, err := p.EvaluateAll(ctx)
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Evaluate terraform files",
				Detail:   err.Error(),
			},
		}
	}

	diags := make(hcl.Diagnostics, 0)
	rp, rpDiags := RichParameters(modules)
	tags, tagDiags := WorkspaceTags(modules, p.Files())

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
