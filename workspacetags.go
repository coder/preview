package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/coder/preview/types"
)

func workspaceTags(modules terraform.Modules, files map[string]*hcl.File) (types.TagBlocks, hcl.Diagnostics) {
	diags := make(hcl.Diagnostics, 0)
	tagBlocks := make(types.TagBlocks, 0)

	for _, mod := range modules {
		blocks := mod.GetDatasByType("coder_workspace_tags")
		for _, block := range blocks {
			evCtx := block.Context().Inner()

			tagsAttr := block.GetAttribute("tags")
			if tagsAttr.IsNil() {
				r := block.HCLBlock().Body.MissingItemRange()
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Missing required argument",
					Detail:      `"tags" attribute is required by coder_workspace_tags blocks`,
					Subject:     &r,
					EvalContext: evCtx,
				})
				continue
			}

			tagsValue := tagsAttr.Value()
			if !tagsValue.Type().IsObjectType() {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Incorrect type for \"tags\" attribute",
					// TODO: better error message for types
					Detail:      fmt.Sprintf(`"tags" attribute must be an 'Object', but got %q`, tagsValue.Type().FriendlyName()),
					Subject:     &tagsAttr.HCLAttribute().NameRange,
					Context:     &tagsAttr.HCLAttribute().Range,
					Expression:  tagsAttr.HCLAttribute().Expr,
					EvalContext: block.Context().Inner(),
				})
				continue
			}

			//tagsObj, ok := tagsAttr.HCLAttribute().Expr.(*hclsyntax.ObjectConsExpr)
			//if !ok {
			//	diags = diags.Append(&hcl.Diagnostic{
			//		Severity: hcl.DiagError,
			//		Summary:  "Incorrect type for \"tags\" attribute",
			//		// TODO: better error message for types
			//		Detail:      fmt.Sprintf(`"tags" attribute must be an 'ObjectConsExpr', but got %T`, tagsAttr.HCLAttribute().Expr),
			//		Subject:     &tagsAttr.HCLAttribute().NameRange,
			//		Context:     &tagsAttr.HCLAttribute().Range,
			//		Expression:  tagsAttr.HCLAttribute().Expr,
			//		EvalContext: block.Context().Inner(),
			//	})
			//	continue
			//}

			var tags []types.Tag
			tagsValue.ForEachElement(func(key cty.Value, val cty.Value) (stop bool) {
				r := tagsAttr.HCLAttribute().Expr.Range()
				tag, tagDiag := newTag(&r, files, key, val)
				if tagDiag != nil {
					diags = diags.Append(tagDiag)
					return false
				}

				tags = append(tags, tag)

				return false
			})
			//for _, item := range tagsObj.Items {
			//	tag, tagDiag := newTag(tagsObj, files, item, evCtx)
			//	if tagDiag != nil {
			//		diags = diags.Append(tagDiag)
			//		continue
			//	}
			//
			//	tags = append(tags, tag)
			//}
			tagBlocks = append(tagBlocks, types.TagBlock{
				Tags:  tags,
				Block: block,
			})
		}
	}

	return tagBlocks, diags
}

// newTag creates a workspace tag from its hcl expression.
func newTag(srcRange *hcl.Range, files map[string]*hcl.File, key, val cty.Value) (types.Tag, *hcl.Diagnostic) {
	//key, kdiags := expr.KeyExpr.Value(evCtx)
	//val, vdiags := expr.ValueExpr.Value(evCtx)

	// TODO: ???

	//if kdiags.HasErrors() {
	//	key = cty.UnknownVal(cty.String)
	//}
	//if vdiags.HasErrors() {
	//	val = cty.UnknownVal(cty.String)
	//}

	if key.IsKnown() && key.Type() != cty.String {
		return types.Tag{}, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid key type for tags",
			Detail:   fmt.Sprintf("Key must be a string, but got %s", key.Type().FriendlyName()),
			//Subject:  &r,
			Context: srcRange,
			//Expression:  expr.KeyExpr,
			//EvalContext: evCtx,
		}
	}

	if val.IsKnown() && val.Type() != cty.String {
		fr := "<nil>"
		if !val.Type().Equals(cty.NilType) {
			fr = val.Type().FriendlyName()
		}
		//r := expr.ValueExpr.Range()
		return types.Tag{}, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value type for tag",
			Detail:   fmt.Sprintf("Value must be a string, but got %s", fr),
			//Subject:     &r,
			Context: srcRange,
			//Expression:  expr.ValueExpr,
			//EvalContext: evCtx,
		}
	}

	tag := types.Tag{
		Key: types.HCLString{
			Value: key,
			//ValueDiags: kdiags,
			//ValueExpr:  expr.KeyExpr,
		},
		Value: types.HCLString{
			Value: val,
			//ValueDiags: vdiags,
			//ValueExpr:  expr.ValueExpr,
		},
	}

	//ks, err := source(expr.KeyExpr.Range(), files)
	//if err == nil {
	//	src := string(ks)
	//	tag.Key.Source = &src
	//}
	//
	//vs, err := source(expr.ValueExpr.Range(), files)
	//if err == nil {
	//	src := string(vs)
	//	tag.Value.Source = &src
	//}

	return tag, nil
}
