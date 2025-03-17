package clidisplay

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/coder/preview/types"
)

func WorkspaceTags(writer io.Writer, tags types.TagBlocks) hcl.Diagnostics {
	var diags hcl.Diagnostics

	tableWriter := table.NewWriter()
	tableWriter.SetTitle("Provisioner Tags")
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Options.SeparateColumns = false
	row := table.Row{"Key", "Value", "Refs"}
	tableWriter.AppendHeader(row)
	for _, tb := range tags {
		for _, tag := range tb.Tags {
			if tag.Valid() {
				k, v := tag.AsStrings()
				tableWriter.AppendRow(table.Row{k, v, ""})
				continue
				//diags = diags.Extend(tDiags)
				//if !diags.HasErrors() {
				//	tableWriter.AppendRow(table.Row{k, v, ""})
				//	continue
				//}
			}

			k := tag.KeyString()
			refs := tag.References()
			tableWriter.AppendRow(table.Row{k, "??", strings.Join(refs, "\n")})

			//refs := tb.AllReferences()
			//refsStr := make([]string, 0, len(refs))
			//for _, ref := range refs {
			//	refsStr = append(refsStr, ref.String())
			//}
			//tableWriter.AppendRow(table.Row{unknown, "???", strings.Join(refsStr, "\n")})
		}
	}
	_, _ = fmt.Fprintln(writer, tableWriter.Render())
	return diags
}

func Parameters(writer io.Writer, params []types.Parameter, files map[string]*hcl.File) {
	tableWriter := table.NewWriter()
	//tableWriter.SetTitle("Parameters")
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Options.SeparateColumns = false
	row := table.Row{"Parameter"}
	tableWriter.AppendHeader(row)
	for _, p := range params {
		strVal := p.Value.AsString()
		//value := p.Value.Value
		//
		//if value.IsNull() {
		//	strVal = "null"
		//} else if !p.Value.Value.IsKnown() {
		//	strVal = "unknown"
		//} else if value.Type().Equals(cty.String) {
		//	strVal = value.AsString()
		//} else {
		//	strVal = value.GoString()
		//}

		tableWriter.AppendRow(table.Row{
			fmt.Sprintf("(%s) %s: %s\n%s", p.DisplayName, p.Name, p.Description, formatOptions(strVal, p.Options)),
		})

		if hcl.Diagnostics(p.Diagnostics).HasErrors() {
			var out bytes.Buffer
			WriteDiagnostics(&out, files, hcl.Diagnostics(p.Diagnostics))
			tableWriter.AppendRow(table.Row{out.String()})

		}

		tableWriter.AppendSeparator()
	}
	_, _ = fmt.Fprintln(writer, tableWriter.Render())
}

func formatOptions(selected string, options []*types.ParameterOption) string {
	var str strings.Builder
	sep := ""
	found := false

	for _, opt := range options {
		str.WriteString(sep)
		prefix := "[ ]"
		if opt.Value.AsString() == selected {
			prefix = "[X]"
			found = true
		}

		str.WriteString(fmt.Sprintf("%s %s (%s)", prefix, opt.Name, opt.Value.AsString()))
		if opt.Description != "" {
			str.WriteString(fmt.Sprintf("\n    %s", maxLength(opt.Description, 25)))
		}
		sep = "\n"
	}
	if !found {
		str.WriteString(sep)
		str.WriteString(fmt.Sprintf("= %s", selected))
	}
	return str.String()
}

func maxLength(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
