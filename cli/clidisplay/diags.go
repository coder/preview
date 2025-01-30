package clidisplay

import (
	"io"
	"log"

	"github.com/hashicorp/hcl/v2"
)

func WriteDiagnostics(out io.Writer, files map[string]*hcl.File, diags hcl.Diagnostics) {
	wr := hcl.NewDiagnosticTextWriter(out, files, 80, true)
	werr := wr.WriteDiagnostics(diags)
	if werr != nil {
		log.Printf("diagnostic writer: %s", werr.Error())
	}
}
