//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"syscall/js"

	"github.com/spf13/afero"

	"github.com/coder/preview"
	"github.com/coder/preview/types"
)

func main() {
	// Create a channel to keep the Go program alive
	done := make(chan struct{}, 0)

	// Expose the Go function `fibonacciSum` to JavaScript
	js.Global().Set("go_preview", js.FuncOf(tfpreview))
	js.Global()

	// Block the program from exiting
	<-done
}

func tfpreview(this js.Value, p []js.Value) any {
	tf, err := fileTreeFS(p[0])
	if err != nil {
		return err
	}

	output, diags := preview.Preview(context.Background(), preview.Input{
		PlanJSONPath:    "",
		PlanJSON:        nil,
		ParameterValues: nil,
		Owner:           types.WorkspaceOwner{},
	}, tf)

	data, _ := json.Marshal(map[string]any{
		"output": output,
		"diags":  diags,
	})
	return js.ValueOf(string(data))
}

func fileTreeFS(value js.Value) (fs.FS, error) {
	data := js.Global().Get("JSON").Call("stringify", value).String()
	var filetree map[string]any
	if err := json.Unmarshal([]byte(data), &filetree); err != nil {
		return nil, err
	}

	mem := afero.NewMemMapFs()
	loadTree(mem, filetree)

	return afero.NewIOFS(mem), nil
}

func loadTree(mem afero.Fs, fileTree map[string]any, path ...string) {
	dir := filepath.Join(path...)
	err := mem.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("error creating directory %q: %v\n", dir, err)
	}
	for k, v := range fileTree {
		switch vv := v.(type) {
		case string:
			fn := filepath.Join(dir, k)
			f, err := mem.Create(fn)
			if err != nil {
				fmt.Printf("error creating file %q: %v\n", fn, err)
				continue
			}
			_, _ = f.WriteString(vv)
			f.Close()
		case map[string]any:
			loadTree(mem, vv, append(path, k)...)
		default:
			fmt.Printf("unknown type %T for %q\n", v, k)
		}
	}
}
