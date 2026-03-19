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
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview"
	"github.com/coder/preview/hclext"
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

		expTags      map[string]string
		unknownTags  []string
		params       map[string]assertParam
		variables    map[string]assertVariable
		presetsFuncs func(t *testing.T, presets []types.Preset)
		presets      map[string]assertPreset
		warnings     []*regexp.Regexp
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
			variables: map[string]assertVariable{
				"string": av().def(cty.StringVal("Hello, world!")).typeEq(cty.String).
					description("test").nullable(true).sensitive(true),
				"number":        av().def(cty.NumberIntVal(7)).typeEq(cty.Number),
				"boolean":       av().def(cty.BoolVal(true)).typeEq(cty.Bool),
				"coerce_string": av().def(cty.StringVal("5")).typeEq(cty.String),
				"complex": av().typeEq(cty.Object(map[string]cty.Type{
					"list": cty.List(cty.String),
					"name": cty.String,
					"age":  cty.Number,
				})),
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
			variables: map[string]assertVariable{
				"regions": av().def(cty.SetVal([]cty.Value{
					cty.StringVal("us"), cty.StringVal("eu"), cty.StringVal("au"),
				})).typeEq(cty.Set(cty.String)),
			},
		},
		{
			name: "dynamic block with nested locals",
			skip: "requires trivy fork fix: expandDynamic must use IsWhollyKnown() instead of IsKnown()",
			dir:  "dynamicblock-nested-locals",
			params: map[string]assertParam{
				"ide_picker": ap().
					optNames("VSCode").
					optVals("vscode").
					formType(provider.ParameterFormTypeMultiSelect),
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
			variables: map[string]assertVariable{
				"regions": av().typeEq(cty.List(cty.String)),
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
			name:        "valid prebuild",
			dir:         "preset",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"number":      ap(),
				"has_default": ap(),
			},
			presets: map[string]assertPreset{
				"valid_preset": aPre().
					value("number", "9").
					value("has_default", "changed").
					prebuildCount(3),
				"prebuild_instance_zero": aPre().
					prebuildCount(0),
				"not_prebuild": aPre().
					prebuildCount(0),
			},
		},
		{
			name:        "invalid presets",
			dir:         "invalidpresets",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"valid_parameter_name": ap().
					optVals("valid_option_value"),
			},
			presets: map[string]assertPreset{
				"empty_parameters":        aPre(),
				"no_parameters":           aPre(),
				"invalid_parameter_name":  aPreWithDiags().errorDiagnostics("Preset parameter \"invalid_parameter_name\" is not defined by the template."),
				"invalid_parameter_value": aPreWithDiags().errorDiagnostics("the value \"invalid_value\" must be defined as one of options"),
				"valid_preset":            aPre().value("valid_parameter_name", "valid_option_value"),
				"another_default_preset":  aPre().def(true),
				"default_preset":          aPreWithDiags().errorDiagnostics("Only one preset can be marked as default. \"another_default_preset\" is already marked as default"),
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
			variables: map[string]assertVariable{
				"security": av().def(cty.StringVal("high")).typeEq(cty.String),
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
			variables: map[string]assertVariable{
				"jetbrains_ides":     av().typeEq(cty.List(cty.String)).description("The list of IDE product codes."),
				"releases_base_link": av(),
				"channel":            av(),
				"download_base_link": av(),
				"arch":               av(),
				"jetbrains_ide_versions": av().typeEq(cty.Map(cty.Object(map[string]cty.Type{
					"build_number": cty.String,
					"version":      cty.String,
				}))),
			},
		},
		{
			name:    "tfvars_from_file",
			dir:     "tfvars",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"variable_values": ap().
					def("alex").optVals("alex", "bob", "claire", "jason"),
			},
			variables: map[string]assertVariable{
				"one":   av(),
				"two":   av(),
				"three": av(),
				"four":  av(),
			},
		},
		{
			name:    "tfvars_from_input",
			dir:     "tfvars",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{},
				TFVars: map[string]cty.Value{
					"one":   cty.StringVal("andrew"),
					"two":   cty.StringVal("bill"),
					"three": cty.StringVal("carter"),
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"variable_values": ap().
					def("andrew").optVals("andrew", "bill", "carter", "jason"),
			},
			variables: map[string]assertVariable{
				"one":   av(),
				"two":   av(),
				"three": av(),
				"four":  av(),
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
			variables: map[string]assertVariable{
				"unknown": av().def(cty.NilVal),
			},
		},
		{
			name:        "presetok",
			dir:         "presetok",
			expTags:     map[string]string{},
			input:       preview.Input{},
			unknownTags: []string{},
			params: map[string]assertParam{
				"use_custom_image": ap().value("false"),
			},
			presets: map[string]assertPreset{
				"valid_preset": aPre().
					value("use_custom_image", "true").
					value("custom_image_url", "docker.io/codercom/test:latest").
					prebuildCount(1),
			},
		},
		{
			name:    "presetok-true-input",
			dir:     "presetok",
			expTags: map[string]string{},
			input: preview.Input{
				ParameterValues: map[string]string{
					"use_custom_image": "true",
					"custom_image_url": "hello world",
				},
			},
			unknownTags: []string{},
			params: map[string]assertParam{
				"use_custom_image": ap().value("true"),
				"custom_image_url": ap().value("hello world"),
			},
			presets: map[string]assertPreset{
				"valid_preset": aPre().
					value("use_custom_image", "true").
					value("custom_image_url", "docker.io/codercom/test:latest").
					prebuildCount(1),
			},
		},
		{
			name: "override",
			dir:  "override",
			params: map[string]assertParam{
				"region":            ap().value("ap").def("ap").optVals("ap"),
				"size":              ap().value("50").def("50").optVals("10", "50", "100"),
				"static_to_dynamic": ap().value("a").def("a").optVals("a", "b", "c"),
				"dynamic_to_static": ap().value("x").def("x").optVals("x", "y"),
			},
			presets: map[string]assertPreset{
				"dev-override": aPre().value("region", "ap"),
			},
			expTags: map[string]string{
				"env":  "production",
				"team": "mango",
			},
			variables: map[string]assertVariable{
				"string_to_number": av().def(cty.NumberIntVal(40)).typeEq(cty.Number),
				"zones":            av().def(cty.SetVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("c")})).typeEq(cty.Set(cty.String)),
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

			// Validate prebuilds too
			preview.ValidatePrebuilds(context.Background(), tc.input, output.Presets, dirFs)

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

			for _, preset := range output.Presets {
				check, ok := tc.presets[preset.Name]
				require.True(t, ok, "unknown preset %s", preset.Name)
				check(t, preset)
			}

			// Assert variables
			require.Len(t, output.Variables, len(tc.variables), "wrong number of variables expected")
			for _, variable := range output.Variables {
				check, ok := tc.variables[variable.Name]
				require.True(t, ok, "unknown variable %s", variable.Name)
				check(t, variable)
			}
		})
	}
}

func TestPresetValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		dir          string
		input        preview.Input
		presetAssert map[string]assertPreset
	}{
		{
			name:  "preset failure",
			dir:   "presetfail",
			input: preview.Input{},
			presetAssert: map[string]assertPreset{
				"invalid_parameters": aPreWithDiags().
					errorDiagnostics("Parameter no_default: Required parameter not provided"),
				"valid_preset": aPre().
					value("has_default", "changed").
					value("no_default", "custom value").
					noDiagnostics(),
				"prebuild_instance_zero": aPre().noDiagnostics().prebuildCount(0),
				"not_prebuild":           aPre().noDiagnostics().prebuildCount(0),
			},
		},
		{
			name:  "preset ok",
			dir:   "presetok",
			input: preview.Input{},
			presetAssert: map[string]assertPreset{
				"valid_preset": aPre().
					value("use_custom_image", "true").
					value("custom_image_url", "docker.io/codercom/test:latest").
					noDiagnostics(),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dirFs := os.DirFS(filepath.Join("testdata", tc.dir))
			output, diags := preview.Preview(context.Background(), tc.input, dirFs)
			if diags.HasErrors() {
				t.Logf("diags: %s", diags)
			}
			require.False(t, diags.HasErrors())
			require.Len(t, diags, 0)

			preview.ValidatePrebuilds(context.Background(), tc.input, output.Presets, dirFs)
			for _, preset := range output.Presets {
				check, ok := tc.presetAssert[preset.Name]
				require.True(t, ok, "unknown preset %s", preset.Name)
				check(t, preset)
				delete(tc.presetAssert, preset.Name)
			}

			require.Len(t, tc.presetAssert, 0, "some presets were not found")
		})
	}
}

type assertVariable func(t *testing.T, variable types.Variable)

func av() assertVariable {
	return func(t *testing.T, v types.Variable) {
		t.Helper()
		assert.Empty(t, v.Diagnostics, "variable should have no diagnostics")
	}
}

func avWithDiags() assertVariable {
	return func(t *testing.T, parameter types.Variable) {}
}

func (a assertVariable) errorDiagnostics(patterns ...string) assertVariable {
	return a.diagnostics(hcl.DiagError, patterns...)
}

func (a assertVariable) warnDiagnostics(patterns ...string) assertVariable {
	return a.diagnostics(hcl.DiagWarning, patterns...)
}

func (a assertVariable) diagnostics(sev hcl.DiagnosticSeverity, patterns ...string) assertVariable {
	shadow := patterns
	return a.extend(func(t *testing.T, v types.Variable) {
		assertDiags(t, sev, v.Diagnostics, shadow...)
	})
}

func (a assertVariable) nullable(n bool) assertVariable {
	return a.extend(func(t *testing.T, v types.Variable) {
		assert.Equal(t, v.Nullable, n, "variable nullable check")
	})
}

func (a assertVariable) typeEq(ty cty.Type) assertVariable {
	return a.extend(func(t *testing.T, v types.Variable) {
		assert.Truef(t, ty.Equals(v.Type), "%q variable type equality check", v.Name)
	})
}

func (a assertVariable) def(def cty.Value) assertVariable {
	return a.extend(func(t *testing.T, v types.Variable) {
		if !assert.Truef(t, def.Equals(v.Default).True(), "%q variable default equality check", v.Name) {
			exp, _ := hclext.AsString(def)
			got, _ := hclext.AsString(v.Default)
			t.Logf("Expected: %s, Value: %s", exp, got)
		}
	})
}

func (a assertVariable) sensitive(s bool) assertVariable {
	return a.extend(func(t *testing.T, v types.Variable) {
		assert.Equal(t, v.Sensitive, s, "variable sensitive check")
	})
}

func (a assertVariable) description(d string) assertVariable {
	return a.extend(func(t *testing.T, v types.Variable) {
		assert.Equal(t, v.Description, d, "variable description check")
	})
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
		assertDiags(t, sev, parameter.Diagnostics, shadow...)
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

//nolint:revive
func (a assertVariable) extend(f assertVariable) assertVariable {
	if a == nil {
		a = func(t *testing.T, v types.Variable) {}
	}

	return func(t *testing.T, v types.Variable) {
		t.Helper()
		(a)(t, v)
		f(t, v)
	}
}

type assertPreset func(t *testing.T, preset types.Preset)

func aPre() assertPreset {
	return func(t *testing.T, preset types.Preset) {
		t.Helper()
		assert.Empty(t, preset.Diagnostics, "preset should have no diagnostics")
	}
}

func aPreWithDiags() assertPreset {
	return func(t *testing.T, parameter types.Preset) {}
}

func (a assertPreset) def(def bool) assertPreset {
	return a.extend(func(t *testing.T, preset types.Preset) {
		require.Equal(t, def, preset.Default)
	})
}

func (a assertPreset) prebuildCount(exp int) assertPreset {
	return a.extend(func(t *testing.T, preset types.Preset) {
		if exp == 0 && preset.Prebuilds == nil {
			return
		}
		require.NotNilf(t, preset.Prebuilds, "prebuild should not be nil, expected %d instances", exp)
		require.Equal(t, exp, preset.Prebuilds.Instances)
	})
}

func (a assertPreset) value(key, value string) assertPreset {
	return a.extend(func(t *testing.T, preset types.Preset) {
		v, ok := preset.Parameters[key]
		require.Truef(t, ok, "preset parameter %q existence check", key)
		assert.Equalf(t, value, v, "preset parameter %q value equality check", key)
	})
}

func (a assertPreset) errorDiagnostics(patterns ...string) assertPreset {
	return a.diagnostics(hcl.DiagError, patterns...)
}

func (a assertPreset) warnDiagnostics(patterns ...string) assertPreset {
	return a.diagnostics(hcl.DiagWarning, patterns...)
}

func (a assertPreset) diagnostics(sev hcl.DiagnosticSeverity, patterns ...string) assertPreset {
	shadow := patterns
	return a.extend(func(t *testing.T, preset types.Preset) {
		t.Helper()

		assertDiags(t, sev, preset.Diagnostics, shadow...)
	})
}

func (a assertPreset) noDiagnostics() assertPreset {
	return a.extend(func(t *testing.T, preset types.Preset) {
		t.Helper()

		assert.Empty(t, preset.Diagnostics, "parameter should have no diagnostics")
	})
}

//nolint:revive
func (a assertPreset) extend(f assertPreset) assertPreset {
	if a == nil {
		a = func(t *testing.T, v types.Preset) {}
	}

	return func(t *testing.T, v types.Preset) {
		t.Helper()
		(a)(t, v)
		f(t, v)
	}
}

func assertDiags(t *testing.T, sev hcl.DiagnosticSeverity, diags types.Diagnostics, patterns ...string) {
	t.Helper()
	checks := make([]string, len(patterns))
	copy(checks, patterns)

DiagLoop:
	for _, diag := range diags {
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
}
