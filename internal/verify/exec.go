package verify

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/require"
)

type WorkingExecutable struct {
	Executable
	WorkingDir string
	TF         *tfexec.Terraform
}

type Executable struct {
	ExecPath string
	Version  string
	DirPath  string

	ins src.Installable
}

func (e Executable) WorkingDir(dir string) (WorkingExecutable, error) {
	tf, err := tfexec.NewTerraform(dir, e.ExecPath)
	if err != nil {
		return WorkingExecutable{}, fmt.Errorf("create terraform exec: %w", err)
	}

	return WorkingExecutable{
		Executable: e,
		WorkingDir: dir,
		TF:         tf,
	}, nil
}

//func (e WorkingExecutable) Init(ctx context.Context) error {
//	e.TF.Init(ctx, tfexec.Upgrade(true))
//}

// TerraformTestVersions returns a list of Terraform versions to test.
func TerraformTestVersions(ctx context.Context) []src.Installable {
	lv := LatestTerraformVersion(ctx)
	return []src.Installable{
		lv,
	}
}

func InstallTerraforms(ctx context.Context, t *testing.T, installables ...src.Installable) []Executable {
	// All terraform versions are installed in the same root directory
	root := t.TempDir()
	execPaths := make([]Executable, 0, len(installables))

	for _, installable := range installables {
		ex := Executable{
			ins: installable,
		}
		switch tfi := installable.(type) {
		case *releases.ExactVersion:
			ver := tfi.Version.String()
			t.Logf("Installing Terraform %s", ver)
			tfi.InstallDir = filepath.Join(root, ver)

			err := os.Mkdir(tfi.InstallDir, 0o755)
			require.NoErrorf(t, err, "tf install %q", ver)

			ex.Version = ver
			ex.DirPath = tfi.InstallDir
		case *releases.LatestVersion:
			t.Logf("Installing latest Terraform")
			ver := "latest"
			tfi.InstallDir = filepath.Join(root, ver)

			err := os.Mkdir(tfi.InstallDir, 0o755)
			require.NoErrorf(t, err, "tf install %q", ver)

			ex.Version = ver
			ex.DirPath = tfi.InstallDir
		default:
			// We only support the types we know about
			t.Fatalf("unknown installable type %T", tfi)
		}

		execPath, err := installable.Install(ctx)
		require.NoErrorf(t, err, "tf install")
		ex.ExecPath = execPath

		execPaths = append(execPaths, ex)
	}

	return execPaths
}

func LatestTerraformVersion(ctx context.Context) *releases.LatestVersion {
	return &releases.LatestVersion{
		Product: product.Terraform,
	}
}

// TerraformVersions will return all versions that match the constraints plus the
// current latest version.
func TerraformVersions(ctx context.Context, constraints version.Constraints) ([]*releases.ExactVersion, error) {
	if len(constraints) == 0 {
		return nil, fmt.Errorf("no constraints provided, don't fetch everything")
	}

	srcs, err := (&releases.Versions{
		Product:     product.Terraform,
		Enterprise:  nil,
		Constraints: constraints,
		ListTimeout: time.Second * 60,
		Install: releases.InstallationOptions{
			Timeout:                  0,
			Dir:                      "",
			SkipChecksumVerification: false,
		},
	}).List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Terraform versions: %w", err)
	}

	include := make([]*releases.ExactVersion, 0)
	for _, src := range srcs {
		ev, ok := src.(*releases.ExactVersion)
		if !ok {
			return nil, fmt.Errorf("failed to cast src to ExactVersion, type was %T", src)
		}

		include = append(include, ev)
	}

	return include, nil
}

func Apply() {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	workingDir := "/path/to/working/dir"
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	fmt.Println(state.FormatVersion) // "0.1"
}
