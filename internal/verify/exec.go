package verify

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
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

func (e WorkingExecutable) Init(ctx context.Context) error {
	return e.TF.Init(ctx, tfexec.Upgrade(true))
}

func (e WorkingExecutable) Plan(ctx context.Context, outPath string) (bool, error) {
	changes, err := e.TF.Plan(ctx, tfexec.Out(outPath))
	return changes, err
}

func (e WorkingExecutable) Apply(ctx context.Context) ([]byte, error) {
	var out bytes.Buffer
	err := e.TF.ApplyJSON(ctx, &out)
	return out.Bytes(), err
}

func (e WorkingExecutable) ShowPlan(ctx context.Context, planPath string) (*tfjson.Plan, error) {
	return e.TF.ShowPlanFile(ctx, planPath)
}

func (e WorkingExecutable) Show(ctx context.Context) (*tfjson.State, error) {
	return e.TF.Show(ctx)
}

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

// CopyTFFS is copied from os.CopyFS and ignores tfstate and lockfiles.
func CopyTFFS(dir string, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), ".terraform") {
			return nil
		}

		fpath, err := filepath.Localize(path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dir, fpath)
		if d.IsDir() {
			return os.MkdirAll(newPath, 0777)
		}

		// TODO(panjf2000): handle symlinks with the help of fs.ReadLinkFS
		// 		once https://go.dev/issue/49580 is done.
		//		we also need filepathlite.IsLocal from https://go.dev/cl/564295.
		if !d.Type().IsRegular() {
			return &os.PathError{Op: "CopyFS", Path: path, Err: os.ErrInvalid}
		}

		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			return &os.PathError{Op: "Copy", Path: newPath, Err: err}
		}
		return w.Close()
	})
}
