//go:build js && wasm

package main

import (
	"syscall/js"
)

func main() {
	// Create a channel to keep the Go program alive
	done := make(chan struct{}, 0)

	// Expose the Go function `fibonacciSum` to JavaScript
	js.Global().Set("go_preview", js.FuncOf(Hello))

	// Block the program from exiting
	<-done
}

func Hello(this js.Value, p []js.Value) any {
	return js.ValueOf("Hello, World!")
}
