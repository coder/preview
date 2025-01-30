package preview

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"
)

type attributeParser struct {
	eval    *terraform.Evaluator
	content *hcl.BodyContent
	diags   hcl.Diagnostics
}

func newAttributeParser(content *hcl.BodyContent, eval *terraform.Evaluator) *attributeParser {
	return &attributeParser{
		content: content,
		diags:   make(hcl.Diagnostics, 0),
		eval:    eval,
	}
}

func (a *attributeParser) Attr(key string) *expectedAttribute {
	return &expectedAttribute{
		Key: key,
		p:   a,
	}
}

type expectedAttribute struct {
	Key  string
	diag hcl.Diagnostics
	p    *attributeParser
}

func (a *expectedAttribute) error(diag hcl.Diagnostics) *expectedAttribute {
	if a.diag != nil {
		return a // already have an error, don't overwrite
	}

	a.p.diags = a.p.diags.Extend(diag)
	a.diag = diag
	return a
}

func (a *expectedAttribute) required() *expectedAttribute {
	_, ok := a.p.content.Attributes[a.Key]
	if !ok {
		a.error(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Missing required attribute %q", a.Key),
				Subject:  &a.p.content.MissingItemRange,
				Extra:    nil,
			},
		})
	}

	return a
}

func (a *expectedAttribute) string() string {
	attr, ok := a.p.content.Attributes[a.Key]
	if !ok {
		return ""
	}

	val, diags := a.p.eval.EvaluateExpr(attr.Expr, cty.String)
	if diags.HasErrors() {
		a.error(diags)
		return ""
	}

	return val.AsString()
}
