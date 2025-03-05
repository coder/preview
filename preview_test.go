package preview_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/coder/preview"
	"github.com/coder/preview/types"
)

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
		unknownTags []string
		params      map[string]assertParam
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
			unknownTags: []string{},
			params: map[string]assertParam{
				"Region": ap().value("us").
					def("us").
					optVals("us", "eu"),
			},
		},
		{
			name:        "conditional-no-inputs",
			dir:         "conditional",
			expTags:     map[string]string{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"Project": ap().
					optVals("small", "massive").
					value("massive"),
				"Compute": ap().
					optVals("micro", "small", "medium", "huge").
					value("huge"),
			},
		},
		{
			name:        "conditional-inputs",
			dir:         "conditional",
			expTags:     map[string]string{},
			unknownTags: []string{},
			input: preview.Input{
				ParameterValues: map[string]string{
					"Project": "small",
					"Compute": "micro",
				},
			},
			params: map[string]assertParam{
				"Project": ap().
					optVals("small", "massive").
					def("massive").
					value("small"),
				"Compute": ap().
					optVals("micro", "small").
					def("small").
					value("micro"),
			},
		},
		{
			name: "tags from param values",
			dir:  "paramtags",
			expTags: map[string]string{
				"zone": "eu",
			},
			input: preview.Input{
				ParameterValues: map[string]string{
					"Region": "eu",
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"Region": ap().value("eu"),
			},
		},
		{
			name: "dynamic block",
			dir:  "dynamicblock",
			expTags: map[string]string{
				"zone": "eu",
			},
			input: preview.Input{
				ParameterValues: map[string]string{
					"Region": "eu",
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"Region": ap().
					value("eu").
					optVals("us", "eu", "au"),
			},
		},
		{
			name:    "external docker resource",
			dir:     "dockerdata",
			expTags: map[string]string{"qux": "quux"},
			unknownTags: []string{
				"foo", "bar",
			},

			input: preview.Input{},
			params: map[string]assertParam{
				"Example": ap().
					unknown().
					// Value is unknown, but this is the safe string
					value("data.coder_parameter.example.value"),
			},
		},
		{
			name: "external docker resource with plan data",
			dir:  "dockerdata",
			expTags: map[string]string{
				"qux": "quux",
				"foo": "sha256:18305429afa14ea462f810146ba44d4363ae76e4c8dfc38288cf73aa07485005",
				"bar": "sha256:a27fd8080b517143cbbbab9dfb7c8571c40d67d534bbdee55bd6c473f432b177",
			},
			unknownTags: []string{},
			input: preview.Input{
				PlanJSONPath: "plan.json",
			},
			params: map[string]assertParam{
				"Example": ap().
					value("18305429afa14ea462f810146ba44d4363ae76e4c8dfc38288cf73aa07485005"),
			},
		},
		{
			name:    "external module",
			dir:     "module",
			expTags: map[string]string{},
			unknownTags: []string{
				"foo",
			},
			input: preview.Input{},
			params: map[string]assertParam{
				"jetbrains_ide": ap().
					optVals("CL", "GO", "IU", "PY", "WS").
					value("GO"),
			},
		},
		{
			name:    "aws instance list",
			dir:     "instancelist",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "before.json",
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:    "null default",
			dir:     "nulldefault",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:    "many modules",
			dir:     "manymodules",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
				PlanJSONPath:    "out.json",
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:    "dupemodparams",
			dir:     "dupemodparams",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name: "not-exists",
			dir:  "not-existing-directory",
		},
		{
			name:    "groups",
			dir:     "groups",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner: types.WorkspaceOwner{
					Groups: []string{"developer", "manager", "admin"},
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"Groups": ap().
					optVals("developer", "manager", "admin"),
			},
		},
		{
			name:    "ambigious",
			dir:     "ambigious",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner: types.WorkspaceOwner{
					Groups: []string{"developer", "manager", "admin"},
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"IsAdmin": ap().
					value("true"),
				"IsAdmin_Root": ap().
					value("true"),
			},
		},
		{
			name:    "ambigious",
			dir:     "ambigious",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner: types.WorkspaceOwner{
					Groups: []string{},
				},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name: "demo",
			dir:  "demo",
			expTags: map[string]string{
				"cluster": "confidential",
			},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner: types.WorkspaceOwner{
					Groups: []string{"admin"},
				},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:    "defexpression",
			dir:     "defexpression",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "plan.json",
				ParameterValues: map[string]string{},
				Owner:           types.WorkspaceOwner{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{
				//"hash": ap[cty.Value]().value(cty.StringVal("hash")).
				//	f(),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.unknownTags == nil {
				tc.unknownTags = []string{}
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
			validTags := output.WorkspaceTags.Tags()

			assert.Equal(t, tc.expTags, validTags)
			assert.ElementsMatch(t, tc.unknownTags, output.WorkspaceTags.UnusableTags().SafeNames())

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

type assertParam func(t *testing.T, parameter types.Parameter)

func ap() assertParam {
	return func(t *testing.T, parameter types.Parameter) {}
}

func (a assertParam) unknown() assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.False(t, parameter.Value.IsKnown(), "parameter unknown check")
	})
}

func (a assertParam) value(str string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Equal(t, str, parameter.Value.AsString(), "parameter value equality check")
	})
}

func (a assertParam) def(str string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Equal(t, str, parameter.DefaultValue, "parameter default equality check")
	})
}

func (a assertParam) optVals(opts ...string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		var values []string
		for _, opt := range parameter.Options {
			values = append(values, opt.Value)
		}
		assert.ElementsMatch(t, opts, values, "parameter option values equality check")
	})
}

func (a assertParam) opts(opts ...types.ParameterOption) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.ElementsMatch(t, opts, parameter.Options, "parameter options equality check")
	})
}

func (a assertParam) extend(f assertParam) assertParam {
	if a == nil {
		a = func(t *testing.T, parameter types.Parameter) {}
	}

	return func(t *testing.T, parameter types.Parameter) {
		(a)(t, parameter)
		f(t, parameter)
	}
}
