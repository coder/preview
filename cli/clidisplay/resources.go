package clidisplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/coder/preview/types"
	"github.com/coder/terraform-provider-coder/v2/provider"
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
			}

			k := tag.KeyString()
			refs := tag.References()
			tableWriter.AppendRow(table.Row{k, "??", strings.Join(refs, "\n")})
		}
	}
	_, _ = fmt.Fprintln(writer, tableWriter.Render())
	return diags
}

func Parameters(writer io.Writer, params []types.Parameter, files map[string]*hcl.File) {
	tableWriter := table.NewWriter()
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Options.SeparateColumns = false
	row := table.Row{"Parameter"}
	tableWriter.AppendHeader(row)
	for _, p := range params {
		strVal := p.Value.AsString()
		selections := []string{strVal}
		if p.FormType == provider.ParameterFormTypeMultiSelect {
			_ = json.Unmarshal([]byte(strVal), &selections)
		}

		dp := p.DisplayName
		if p.DisplayName == "" {
			dp = p.Name
		}

		tableWriter.AppendRow(table.Row{
			fmt.Sprintf("(%s) %s: %s\n%s", dp, p.Name, p.Description, formatOptions(selections, p.Options)),
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

func Presets(writer io.Writer, presets []types.Preset, files map[string]*hcl.File) {
	tableWriter := table.NewWriter()
	tableWriter.SetStyle(table.StyleLight)
	tableWriter.Style().Options.SeparateColumns = false
	row := table.Row{"Preset"}
	tableWriter.AppendHeader(row)
	for _, p := range presets {
		tableWriter.AppendRow(table.Row{
			fmt.Sprintf("%s\n%s", p.Name, formatPresetParameters(p.Parameters)),
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

func formatPresetParameters(presetParameters map[string]string) string {
	var str strings.Builder
	for presetParamName, PresetParamValue := range presetParameters {
		_, _ = str.WriteString(fmt.Sprintf("%s = %s\n", presetParamName, PresetParamValue))
	}
	return str.String()
}

func formatOptions(selected []string, options []*types.ParameterOption) string {
	var str strings.Builder
	sep := ""
	found := false

	for _, opt := range options {
		_, _ = str.WriteString(sep)
		prefix := "[ ]"
		if slices.Contains(selected, opt.Value.AsString()) {
			prefix = "[X]"
			found = true
		}

		_, _ = str.WriteString(fmt.Sprintf("%s %s (%s)", prefix, opt.Name, opt.Value.AsString()))
		if opt.Description != "" {
			_, _ = str.WriteString(fmt.Sprintf("\n    %s", maxLength(opt.Description, 25)))
		}
		sep = "\n"
	}
	if !found {
		_, _ = str.WriteString(sep)
		_, _ = str.WriteString(fmt.Sprintf("= %s", selected))
	}
	return str.String()
}

func maxLength(s string, m int) string {
	if len(s) > m {
		return s[:m] + "..."
	}
	return s
}
