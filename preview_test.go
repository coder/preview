package preview_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
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
				PlanJSONPath:    "before.json",
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
