//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/coder/preview"
)

func main() {
	// Create a channel to keep the Go program alive
	done := make(chan struct{}, 0)

	// Expose the Go function `fibonacciSum` to JavaScript
	js.Global().Set("go_preview", js.FuncOf(tfpreview))
	js.Global().Set("Loaded", js.FuncOf(loaded))

	// Block the program from exiting
	<-done
}

func tfpreview(this js.Value, p []js.Value) any {
	preview.Preview(context.Background(), preview.Input{}, nil)
}

func loaded(this js.Value, p []js.Value) any {
	return js.ValueOf(true)
}
