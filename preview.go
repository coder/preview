package preview

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform"

	"github.com/coder/preview/types"
)

type Input struct {
	ParameterValues []types.ParameterValue
}

type Output struct {
	Parameters    []types.RichParameter
	WorkspaceTags types.TagBlocks
}

func Preview(ctx context.Context, input Input, dir fs.FS) (*Output, hcl.Diagnostics) {
	adfs := afero.NewReadOnlyFs(afero.FromIOFS{FS: dir})

	// terraform parsing
	tp := terraform.NewParser(adfs)
	mod, diags := tp.LoadConfigDir(".", ".")
	if diags.HasErrors() {
		return nil, diags
	}

	config, diags := terraform.BuildConfig(mod, terraform.ModuleWalkerFunc(
		func(req *terraform.ModuleRequest) (*terraform.Module, *version.Version, hcl.Diagnostics) {
			// TODO: Load in coder registry modules?
			return nil, nil, nil
		}),
	)

	variableValues, diags := terraform.VariableValues(config)
	if diags.HasErrors() {
		return nil, diags
	}

	evaluator := &terraform.Evaluator{
		Meta:           &terraform.ContextMeta{},
		ModulePath:     config.Path.UnkeyedInstanceShim(),
		Config:         config,
		VariableValues: variableValues,
	}

	output, diags := extract(evaluator, mod, input)
	return &output, diags
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
