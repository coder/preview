package preview_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/preview"
	"github.com/coder/preview/types"
	"github.com/coder/terraform-provider-coder/v2/provider"
)

func Test_Extract(t *testing.T) {
	t.Parallel()

	// nice helper to match tf jsonencode
	jsonencode := func(v interface{}) string {
		b, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	var _ = jsonencode

	for _, tc := range []struct {
		skip        string
		name        string
		dir         string
		failPreview bool
		input       preview.Input

		expTags     map[string]string
		unknownTags []string
		params      map[string]assertParam
		warnings    []*regexp.Regexp
	}{
		{
			name:        "bad param values",
			dir:         "badparam",
			failPreview: true,
		},
		{
			name: "sometags",
			dir:  "sometags",
			expTags: map[string]string{
				"string": "foo",
				"number": "42",
				"bool":   "true",
				// null tags are omitted
			},
			unknownTags: []string{
				"complex", "map", "list",
			},
		},
		{
			name: "simple static values",
			dir:  "static",
			expTags: map[string]string{
				"zone": "developers",
			},
			params: map[string]assertParam{
				"region": ap().value("us").
					def("us").
					optVals("us", "eu").
					formType(provider.ParameterFormTypeRadio),
				"numerical": ap().value("5"),
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
				"indexed_0": ap(),
				"indexed_1": ap(),
			},
		},
		{
			name: "external docker resource without plan data",
			dir:  "dockerdata",
			expTags: map[string]string{
				"qux":    "quux",
				"ubuntu": "0000000000000000000000000000000000000000000000000000000000000000",
				"centos": "0000000000000000000000000000000000000000000000000000000000000000",
			},
			unknownTags: []string{},
			input:       preview.Input{},
			params: map[string]assertParam{
				"os": apWithDiags().
					errorDiagnostics("unique").
					value("0000000000000000000000000000000000000000000000000000000000000000"),
			},
		},
		{
			name: "external docker resource with plan data",
			dir:  "dockerdata",
			expTags: map[string]string{
				"qux":    "quux",
				"ubuntu": "18305429afa14ea462f810146ba44d4363ae76e4c8dfc38288cf73aa07485005",
				"centos": "a27fd8080b517143cbbbab9dfb7c8571c40d67d534bbdee55bd6c473f432b177",
			},
			unknownTags: []string{},
			input: preview.Input{
				PlanJSONPath: "plan.json",
			},
			params: map[string]assertParam{
				"os": ap().
					value("18305429afa14ea462f810146ba44d4363ae76e4c8dfc38288cf73aa07485005"),
			},
		},
		{
			name:        "external module",
			dir:         "module",
			expTags:     map[string]string{},
			unknownTags: []string{},
			input: preview.Input{
				PlanJSONPath: "plan.json",
				ParameterValues: map[string]string{
					"extra": "foobar",
				},
			},
			params: map[string]assertParam{
				"jetbrains_ide": ap().
					optVals("CL", "GO", "IU", "PY", "WS").
					value("GO"),
				"extra": ap().
					value("foobar"),
			},
		},
		{
			// TODO: Add aws instance list test with args
			name:    "aws instance list",
			dir:     "instancelist",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "plan.json",
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"Home": ap().
					optVals("us", "eu").def("us").value("us"),
				"Region": ap().def("us-east-1"),
				"instance_type": ap().numOpts(20).
					optExists("m7g.12xlarge").
					optExists("t3a.large").
					def("m7gd.8xlarge").
					value("m7gd.8xlarge"),
			},
		},
		{
			name:    "empty file",
			dir:     "empty",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:        "empty default",
			dir:         "emptydefault",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"word": apWithDiags().
					errorDiagnostics("Required"),
			},
		},
		{
			name:        "required",
			dir:         "required",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"region": apWithDiags().errorDiagnostics("Required"),
			},
		},
		{
			name:    "many modules",
			dir:     "manymodules",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
				PlanJSONPath:    "plan.json",
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"main_question": ap().def("main").
					optVals("main", "one", "two", "1.11.1", "1.15.15", "one-a"),
				"one_question":   ap().def("one").optVals("one", "one-a", "1.11.1"),
				"two_question":   ap().def("two").optVals("two", "1.15.15"),
				"one_a_question": ap().def("one-a").optVals("one-a", "1.11.2", "bar"),
			},
		},
		{
			name:        "dupemodparams",
			dir:         "dupemodparams",
			expTags:     map[string]string{},
			failPreview: true, // duplicate parameters
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:        "dupeparams",
			dir:         "dupeparams",
			expTags:     map[string]string{},
			failPreview: true, // duplicate parameters
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
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
				"groups": ap().
					optVals("developer", "manager", "admin"),
			},
		},
		{
			name:    "submodule cannot affect dynamic parent elements",
			dir:     "submoduledynamic",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner:           types.WorkspaceOwner{},
			},
			unknownTags: []string{},
			// should be 0 params
			params: map[string]assertParam{},
		},
		{
			name: "demo",
			dir:  "demo",
			expTags: map[string]string{
				"cluster": "confidential",
				"hash":    "52bb4d943694f2f5867a251780f85e5a68906787b4ffa3157e29b9ef510b1a97",
			},
			input: preview.Input{
				PlanJSONPath: "plan.json",
				ParameterValues: map[string]string{
					"hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
				},
				Owner: types.WorkspaceOwner{
					Groups: []string{"admin"},
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"hash": ap().
					value("b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"),
				"security_level": ap(),
				"region":         ap(),
				"cpu":            ap(),
				"browser":        ap(),
				"team":           ap().optVals("frontend", "backend", "fullstack"),
				"jetbrains_ide":  ap(),
			},
		},
		{
			name: "demo_flat",
			dir:  "demo_flat",
			expTags: map[string]string{
				"cluster": "confidential",
				"hash":    "52bb4d943694f2f5867a251780f85e5a68906787b4ffa3157e29b9ef510b1a97",
			},
			input: preview.Input{
				PlanJSONPath: "plan.json",
				ParameterValues: map[string]string{
					"hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
				},
				Owner: types.WorkspaceOwner{
					Groups: []string{"admin"},
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"hash": ap().
					value("b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"),
				"security_level": ap(),
				"region":         ap(),
				"cpu":            ap(),
				"browser":        ap(),
				"team":           ap().optVals("frontend", "backend", "fullstack"),
				"jetbrains_ide":  ap(),
			},
		},
		{
			name:    "count",
			dir:     "count",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath:    "",
				ParameterValues: map[string]string{},
				Owner:           types.WorkspaceOwner{},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"ref": ap().value("Index 2").
					optVals("Index 0", "Index 1", "Index 2"),
				"ref_count": ap().value("Index 2"),
			},
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
			params: map[string]assertParam{
				"hash": ap().value("b15ac50ce93fbdae93e39791c64fe77be508abdbf50e72d7c10d18e04983b3f7"),
			},
		},
		{
			name:    "defexpression no plan.json",
			dir:     "defexpression",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
				Owner:           types.WorkspaceOwner{},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"hash": ap().unknown(),
			},
		},
		{
			name:    "cyclical",
			dir:     "cyclical",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"alpha": ap().unknown(),
				"beta":  ap().unknown(),
			},
		},
		{
			name:    "missing_module",
			dir:     "missingmodule",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
			warnings: []*regexp.Regexp{
				regexp.MustCompile("Module not loaded"),
			},
		},
		{
			skip:    "skip until https://github.com/aquasecurity/trivy/pull/8479 is resolved",
			name:    "submodcount",
			dir:     "submodcount",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params:      map[string]assertParam{},
		},
		{
			name:    "plan_stringindex",
			dir:     "plan_stringindex",
			expTags: map[string]string{},
			input: preview.Input{
				PlanJSONPath: "plan.json",
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"jetbrains_ide": ap().
					optVals("GO", "IU", "PY").
					optNames("GoLand 2024.3", "IntelliJ IDEA Ultimate 2024.3", "PyCharm Professional 2024.3"),
			},
		},
		{
			name:        "unknownoption",
			dir:         "unknownoption",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"unknown": apWithDiags().
					errorDiagnostics("The set of options cannot be resolved"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.skip != "" {
				t.Skip(tc.skip)
				return
			}

			if tc.unknownTags == nil {
				tc.unknownTags = []string{}
			}
			if tc.expTags == nil {
				tc.expTags = map[string]string{}
			}

			dirFs := os.DirFS(filepath.Join("testdata", tc.dir))

			output, diags := preview.Preview(context.Background(), tc.input, dirFs)
			if tc.failPreview {
				require.True(t, diags.HasErrors())
				return
			}
			if diags.HasErrors() {
				t.Logf("diags: %s", diags)
			}
			require.False(t, diags.HasErrors())

			if len(tc.warnings) > 0 {
				for _, w := range tc.warnings {
					idx := slices.IndexFunc(diags, func(diagnostic *hcl.Diagnostic) bool {
						return w.MatchString(diagnostic.Error())

					})
					require.Greater(t, idx, -1, "expected warning %q to be present in diags", w.String())
				}
			}

			// Assert tags
			validTags := output.WorkspaceTags.Tags()

			for k, expected := range tc.expTags {
				tag, ok := validTags[k]
				if !ok {
					t.Errorf("expected tag %q to be present in output, but it was not", k)
					continue
				}
				if tag != expected {
					assert.JSONEqf(t, expected, tag, "tag %q does not match expected, nor is it a json equivalent", k)
				}
			}
			assert.Equal(t, len(tc.expTags), len(output.WorkspaceTags.Tags()), "unexpected number of tags in output")

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
	return func(t *testing.T, parameter types.Parameter) {
		t.Helper()
		assert.Empty(t, parameter.Diagnostics, "parameter should have no diagnostics")
	}
}

func apWithDiags() assertParam {
	return func(t *testing.T, parameter types.Parameter) {}
}

func (a assertParam) errorDiagnostics(patterns ...string) assertParam {
	return a.diagnostics(hcl.DiagError, patterns...)
}

func (a assertParam) warnDiagnostics(patterns ...string) assertParam {
	return a.diagnostics(hcl.DiagWarning, patterns...)
}

func (a assertParam) diagnostics(sev hcl.DiagnosticSeverity, patterns ...string) assertParam {
	shadow := patterns
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		checks := make([]string, len(shadow))
		copy(checks, shadow)

	DiagLoop:
		for _, diag := range parameter.Diagnostics {
			if diag.Severity != sev {
				continue
			}
			for i, pat := range checks {
				if strings.Contains(diag.Summary, pat) || strings.Contains(diag.Detail, pat) {
					checks = append(checks[:i], checks[i+1:]...)
					break DiagLoop
				}
			}
		}

		assert.Equal(t, []string{}, checks, "missing expected diagnostic errors")
	})
}

func (a assertParam) noDiagnostics() assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Empty(t, parameter.Diagnostics, "parameter should have no diagnostics")
	})
}

func (a assertParam) formType(exp provider.ParameterFormType) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Equal(t, exp, parameter.FormType, "parameter form type equality check")
	})
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

func (a assertParam) optExists(v string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		for _, opt := range parameter.Options {
			if opt.Value.AsString() == v {
				return
			}
		}
		assert.Failf(t, "parameter option existence check", "option %q not found", v)
	})
}

func (a assertParam) numOpts(n int) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Len(t, parameter.Options, n, "parameter options length check")
	})
}

func (a assertParam) def(str string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.Equal(t, str, parameter.DefaultValue.AsString(), "parameter default equality check")
	})
}

func (a assertParam) optNames(opts ...string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		var values []string
		for _, opt := range parameter.Options {
			values = append(values, opt.Name)
		}
		assert.ElementsMatch(t, opts, values, "parameter option names equality check")
	})
}

func (a assertParam) optVals(opts ...string) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		var values []string
		for _, opt := range parameter.Options {
			values = append(values, opt.Value.AsString())
		}
		assert.ElementsMatch(t, opts, values, "parameter option values equality check")
	})
}

//nolint:unused
func (a assertParam) opts(opts ...types.ParameterOption) assertParam {
	return a.extend(func(t *testing.T, parameter types.Parameter) {
		assert.ElementsMatch(t, opts, parameter.Options, "parameter options equality check")
	})
}

//nolint:revive
func (a assertParam) extend(f assertParam) assertParam {
	if a == nil {
		a = func(t *testing.T, parameter types.Parameter) {}
	}

	return func(t *testing.T, parameter types.Parameter) {
		t.Helper()
		(a)(t, parameter)
		f(t, parameter)
	}
}
