package preview_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coder/preview"
	"github.com/coder/preview/internal/verify"
)

// Test_VerifyE2E will fully evaluate with `terraform apply`
// and verify the output of `preview` against the tfstate. This
// is the e2e test for the preview package.
// It uses the `terraform plan` output, meaning this mimics the
// "Workspace Create" form of preview. Which is arguably the most
// important part of the preview lifecycle.
//
//  1. Terraform versions listed from 'verify.TerraformTestVersions' are
//     installed into a temp directory. If multiple versions exist, a e2e test
//     for each tf version is run.
//  2. For each test directory in `testdata`, the following steps are performed:
//     a. If the directory contains a file named `skipe2e`, skip the test.
//     Some tests are not meant to be run in e2e mode as they require external
//     credentials, or are invalid terraform configurations.
//     b. For each terraform version, the following steps are performed:
//     ---- Core Test ----
//     i. Create a working directory for the terraform version in a temp dir.
//     ii. Copy the test data into the working directory.
//     iii. Run `terraform init`.
//     iv. Run `terraform plan` and save the output to a file.
//     v. Run `terraform apply`.
//     vi. Run `terraform show` to get the state.
//     vii. Run `preview --plan=out.plan` to get the preview state.
//     viii. Compare the preview state with the terraform state.
//     ix. If the preview state is not equal to the terraform state, fail the test.
//     ----           ----
//
// The goal of the test is to compare `tfstate` with the output of `preview`.
// If `preview`'s implementation of terraform is incorrect, the test will fail.
// TODO: Adding varied parameter inputs would be a good idea.
// TODO: Add workspace tag comparisions.
func Test_VerifyE2E(t *testing.T) {
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
			return entry.Name() == "skipe2e"
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
							ParameterValues: map[string]string{},
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
