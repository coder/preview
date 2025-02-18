package preview_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"golang.org/x/exp/slices"

	"github.com/coder/preview"
	"github.com/coder/preview/internal/verify"
	"github.com/coder/preview/types"
)

// Test_VerifyPreview will fully evaluate with `terraform apply`
// and verify the output of `preview` against the tfstate. This
// is the e2e test for the preview package.
func Test_VerifyPreview(t *testing.T) {
	t.Parallel()

	installCtx, cancel := context.WithCancel(context.Background())

	versions := verify.TerraformTestVersions(installCtx)
	tfexecs := verify.InstallTerraforms(installCtx, t, versions...)
	cancel()

	dirFs := os.DirFS("testdata")
	entries, err := fs.ReadDir(dirFs, ".")
	require.NoError(t, err)

	for _, entry := range entries {
		entry := entry
		if !entry.IsDir() {
			t.Logf("skipping non directory file %q", entry.Name())
			continue
		}

		entryFiles, err := fs.ReadDir(dirFs, filepath.Join(entry.Name()))
		require.NoError(t, err, "reading test data dir")
		if !slices.ContainsFunc(entryFiles, func(entry fs.DirEntry) bool {
			return filepath.Ext(entry.Name()) == ".tf"
		}) {
			t.Logf("skipping test data dir %q, no .tf files", entry.Name())
			continue
		}

		if slices.ContainsFunc(entryFiles, func(entry fs.DirEntry) bool {
			return entry.Name() == "skip"
		}) {
			t.Logf("skipping test data dir %q, skip file found", entry.Name())
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			entryWrkPath := t.TempDir()

			for _, tfexec := range tfexecs {
				tfexec := tfexec

				t.Run(tfexec.Version, func(t *testing.T) {
					wp := filepath.Join(entryWrkPath, tfexec.Version)
					err := os.MkdirAll(wp, 0755)
					require.NoError(t, err, "creating working dir")

					t.Logf("working dir %q", wp)

					subFS, err := fs.Sub(dirFs, entry.Name())
					require.NoError(t, err, "creating sub fs")

					err = verify.CopyTFFS(wp, subFS)
					require.NoError(t, err, "copying test data to working dir")

					exe, err := tfexec.WorkingDir(wp)
					require.NoError(t, err, "creating working executable")

					ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
					defer cancel()
					err = exe.Init(ctx)
					require.NoError(t, err, "terraform init")

					planOutFile := "tfplan"
					planOutPath := filepath.Join(wp, planOutFile)
					_, err = exe.Plan(ctx, planOutPath)
					require.NoError(t, err, "terraform plan")

					plan, err := exe.ShowPlan(ctx, planOutPath)
					require.NoError(t, err, "terraform show plan")

					pd, err := json.Marshal(plan)
					require.NoError(t, err, "marshalling plan")

					err = os.WriteFile(filepath.Join(wp, "plan.json"), pd, 0644)
					require.NoError(t, err, "writing plan.json")

					_, err = exe.Apply(ctx)
					require.NoError(t, err, "terraform apply")

					state, err := exe.Show(ctx)
					require.NoError(t, err, "terraform show")

					output, diags := preview.Preview(context.Background(),
						preview.Input{
							PlanJSONPath:    "plan.json",
							ParameterValues: map[string]types.ParameterValue{},
						},
						os.DirFS(wp))
					if diags.HasErrors() {
						t.Logf("diags: %s", diags)
					}
					require.False(t, diags.HasErrors(), "preview errors")

					if state.Values == nil {
						t.Fatalf("state values are nil")
					}
					verify.Compare(t, output, state.Values.RootModule)
				})
			}
		})
	}
}

func TestFoo(t *testing.T) {
	ty, err := gocty.ImpliedType([]any{1, 2, 3})
	require.NoError(t, err)
	fmt.Println(ty.FriendlyName())
}

func Test_Extract(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		dir         string
		failPreview bool
		input       preview.Input

		expTags     map[string]string
		expUnknowns []string
		params      map[string]func(t *testing.T, parameter types.Parameter)
	}{
		{
			name:        "bad param values",
			dir:         "badparam",
			failPreview: true,
		},
		{
			name: "simple static values",
			dir:  "static",
			expTags: map[string]string{
				"zone": "developers",
			},
			expUnknowns: []string{},
			params: map[string]func(t *testing.T, parameter types.Parameter){
				"Region": ap[cty.Value]().value(cty.StringVal("us")).f(),
			},
		},
		{
			name:        "conditional",
			dir:         "conditional",
			expTags:     map[string]string{},
			expUnknowns: []string{},
			params: map[string]func(t *testing.T, parameter types.Parameter){
				"Compute": ap[cty.Value]().value(cty.StringVal("huge")).f(),
				"Project": ap[cty.Value]().value(cty.StringVal("massive")).f(),
			},
		},
		{
			name: "conditional",
			dir:  "conditional",
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{
					"project": {
						Value: cty.StringVal("small"),
					},
				},
			},
			expTags:     map[string]string{},
			expUnknowns: []string{},
			params: map[string]func(t *testing.T, parameter types.Parameter){
				"Compute": ap[cty.Value]().value(cty.StringVal("small")).f(),
				"Project": ap[cty.Value]().value(cty.StringVal("small")).f(),
			},
		},
		{
			name: "tags from param values",
			dir:  "paramtags",
			expTags: map[string]string{
				"zone": "eu",
			},
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{
					"region": {
						Value: cty.StringVal("eu"),
					},
				},
			},
			expUnknowns: []string{},
			params: map[string]func(t *testing.T, parameter types.Parameter){
				"Region": ap[cty.Value]().value(cty.StringVal("eu")).f(),
			},
		},
		{
			name: "dynamic block",
			dir:  "dynamicblock",
			expTags: map[string]string{
				"zone": "eu",
			},
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{
					"region": {
						Value: cty.StringVal("eu"),
					},
				},
			},
			expUnknowns: []string{},
			params: map[string]func(t *testing.T, parameter types.Parameter){
				"Region": ap[cty.Value]().
					value(cty.StringVal("eu")).
					options("us", "eu", "au").
					f(),
			},
		},
		{
			name:    "external docker resource",
			dir:     "dockerdata",
			expTags: map[string]string{"qux": "quux"},
			expUnknowns: []string{
				"foo", "bar",
			},

			input:  preview.Input{},
			params: map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name: "external docker resource with plan data",
			dir:  "dockerdata",
			expTags: map[string]string{
				"qux": "quux",
				"foo": "ubuntu@sha256:80dd3c3b9c6cecb9f1667e9290b3bc61b78c2678c02cbdae5f0fea92cc6734ab",
				"bar": "centos@sha256:a27fd8080b517143cbbbab9dfb7c8571c40d67d534bbdee55bd6c473f432b177",
			},
			expUnknowns: []string{},
			input: preview.Input{
				PlanJSONPath: "plan.json",
			},
			params: map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name:        "external module with external data",
			dir:         "module",
			expTags:     map[string]string{},
			expUnknowns: []string{},
			input: preview.Input{
				PlanJSONPath: "before.json",
			},
			params: map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name:    "aws instance list",
			dir:     "instancelist",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "before.json",
				ParameterValues: map[string]types.ParameterValue{},
			},
			expUnknowns: []string{},
			params:      map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name:    "null default",
			dir:     "nulldefault",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{},
			},
			expUnknowns: []string{},
			params:      map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name:    "many modules",
			dir:     "manymodules",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{},
				PlanJSONPath:    "out.json",
			},
			expUnknowns: []string{},
			params:      map[string]func(t *testing.T, parameter types.Parameter){},
		},
		{
			name:    "test",
			dir:     "test",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]types.ParameterValue{},
			},
			expUnknowns: []string{},
			params:      map[string]func(t *testing.T, parameter types.Parameter){},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.expUnknowns == nil {
				tc.expUnknowns = []string{}
			}
			if tc.expTags == nil {
				tc.expTags = map[string]string{}
			}

			dirFs := os.DirFS(filepath.Join("testdata", tc.dir))
			//a, b := fs.ReadDir(dirFs, ".")
			//fmt.Println(a, b)

			output, diags := preview.Preview(context.Background(), tc.input, dirFs)
			if tc.failPreview {
				require.True(t, diags.HasErrors())
				return
			}
			if diags.HasErrors() {
				t.Logf("diags: %s", diags)
			}
			require.False(t, diags.HasErrors())

			// Assert tags
			validTags := output.WorkspaceTags.ValidTags()

			assert.Equal(t, tc.expTags, validTags)
			assert.ElementsMatch(t, tc.expUnknowns, output.WorkspaceTags.InvalidTags().SafeNames())

			// Assert params
			require.Len(t, output.Parameters, len(tc.params), "wrong number of parameters expected")
			for _, param := range output.Parameters {
				check, ok := tc.params[param.Name]
				require.True(t, ok, "unknown parameter %s", param.Name)
				check(t, param)
			}
		})
	}
}

type assertParam[T any] func(t *testing.T, parameter types.Parameter)

func ap[T any]() *assertParam[T] {
	x := assertParam[T](func(t *testing.T, parameter types.Parameter) {})
	return &x
}

func (a *assertParam[T]) f() func(t *testing.T, parameter types.Parameter) {
	return *a
}

func (a *assertParam[T]) options(opts ...string) *assertParam[T] {
	cpy := *a
	x := assertParam[T](func(t *testing.T, parameter types.Parameter) {
		allOpts := make([]string, 0)
		for _, opt := range parameter.Options {
			allOpts = append(allOpts, opt.Value)
		}
		assert.ElementsMatch(t, opts, allOpts)
		cpy(t, parameter)
	})
	return &x
}

func (a *assertParam[T]) value(v T) *assertParam[T] {
	cpy := *a
	x := assertParam[T](func(t *testing.T, parameter types.Parameter) {
		assert.Equal(t, v, parameter.Value.Value, fmt.Sprintf("param %q", parameter.Name))
		cpy(t, parameter)
	})
	return &x
}
