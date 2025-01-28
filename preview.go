package preview

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/aquasecurity/trivy/pkg/iac/scanners/terraform/parser"
	"github.com/coder/preview/types"
	"github.com/hashicorp/hcl/v2"
)

type Input struct {
	ParameterValues []types.ParameterValue
}

type Output struct {
	Parameters []types.RichParameter
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

	var _ = p

	return nil, nil
}

func (i Input) RichParameterValue(key string) (types.ParameterValue, bool) {
	for _, p := range i.ParameterValues {
		if p.Name == key {
			return p, true
		}
	}
	return types.ParameterValue{}, false
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
