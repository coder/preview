package preview

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser"
	"github.com/hashicorp/hcl/v2"

	"github.com/coder/preview/types"
)

type Input struct {
	ParameterValues map[string]types.ParameterValue
}

type Output struct {
	Parameters    []types.Parameter
	WorkspaceTags types.TagBlocks
	Files         map[string]*hcl.File
}

func Preview(ctx context.Context, input Input, dir fs.FS) (*Output, hcl.Diagnostics) {
	varFiles, err := tfVarFiles("", dir)
	if err != nil {
		return nil, nil
	}

	diags := make(hcl.Diagnostics, 0)
	hook := ParameterContextsEvalHook(input, diags)
	// moduleSource is "" for a local module
	p := parser.New(dir, "",
		parser.OptionWithDownloads(false),
		parser.OptionWithTFVarsPaths(varFiles...),
		parser.OptionWithEvalHook(hook),
	)
	if diags.HasErrors() {
		return nil, diags
	}

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
	var _ = outputs

	rp, diags := RichParameters(modules)
	return &Output{
		Parameters: rp,
		Files:      p.Files(),
	}, diags
}

func (i Input) RichParameterValue(key string) (types.ParameterValue, bool) {
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
