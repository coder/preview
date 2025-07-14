// Code taken from https://github.com/aquasecurity/trivy/blob/0449787eb52854cbdd7f4c5794adbf58965e60f8/pkg/iac/scanners/terraform/parser/load_vars.go
package tfvars

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

// TFVarFiles extracts any .tfvars files from the given directory.
// TODO: Test nested directories and how that should behave.
func TFVarFiles(path string, dir fs.FS) ([]string, error) {
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
			newFiles, err := TFVarFiles(filepath.Join(path, entry.Name()), subD)
			if err != nil {
				return nil, err
			}
			files = append(files, newFiles...)
		}

		if filepath.Ext(entry.Name()) == ".tfvars" || strings.HasSuffix(entry.Name(), ".tfvars.json") {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}

func LoadTFVars(srcFS fs.FS, filenames []string) (map[string]cty.Value, error) {
	combinedVars := make(map[string]cty.Value)

	// Intentionally avoid loading terraform variables from the host environment.
	// Trivy (and terraform) use os.Environ() to search for "TF_VAR_" prefixed
	// environment variables.
	//
	// Preview should be sandboxed, so this code should not be included.

	for _, filename := range filenames {
		vars, err := LoadTFVarsFile(srcFS, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load tfvars from %s: %w", filename, err)
		}
		for k, v := range vars {
			combinedVars[k] = v
		}
	}

	return combinedVars, nil
}

func LoadTFVarsFile(srcFS fs.FS, filename string) (map[string]cty.Value, error) {
	inputVars := make(map[string]cty.Value)
	if filename == "" {
		return inputVars, nil
	}

	src, err := fs.ReadFile(srcFS, filepath.ToSlash(filename))
	if err != nil {
		return nil, err
	}

	var attrs hcl.Attributes
	if strings.HasSuffix(filename, ".json") {
		variableFile, err := hcljson.Parse(src, filename)
		if err != nil {
			return nil, err
		}
		attrs, err = variableFile.Body.JustAttributes()
		if err != nil {
			return nil, err
		}
	} else {
		variableFile, err := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
		if err != nil {
			return nil, err
		}
		attrs, err = variableFile.Body.JustAttributes()
		if err != nil {
			return nil, err
		}
	}

	for _, attr := range attrs {
		inputVars[attr.Name], _ = attr.Expr.Value(&hcl.EvalContext{})
	}

	return inputVars, nil
}
