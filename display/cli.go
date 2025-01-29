package display

import (
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
			if tag.IsKnown() {
				k, v := tag.AsStrings()
				tableWriter.AppendRow(table.Row{k, v, ""})
				continue
			}

			k := tag.SafeKeyString()
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
