package main

import (
	"errors"
	"log"
	"os"

	"github.com/coder/preview/cli"
	"github.com/hashicorp/hcl/v2"
)

func main() {
	log.SetOutput(os.Stderr)
	root := &cli.RootCmd{}
	cmd := root.Root()

	err := cmd.Invoke().WithOS().Run()
	if err != nil {
		var diags hcl.Diagnostics
		if errors.As(err, &diags) {
			var files map[string]*hcl.File
			if root.Files != nil {
				files = root.Files
			}
			wr := hcl.NewDiagnosticTextWriter(os.Stderr, files, 80, true)
			werr := wr.WriteDiagnostics(diags)
			if werr != nil {
				log.Printf("diagnostic writer: %s", werr.Error())
			}
		}
		log.Fatal(err.Error())
		os.Exit(1)
	}
}
