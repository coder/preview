package preview

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type attributeParser struct {
	block *terraform.Block
	diags hcl.Diagnostics
}

func newAttributeParser(block *terraform.Block) *attributeParser {
	return &attributeParser{
		block: block,
		diags: make(hcl.Diagnostics, 0),
	}
}

func (a *attributeParser) attr(key string) *expectedAttribute {
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
	attr := a.p.block.GetAttribute(a.Key)
	if attr.IsNil() {
		r := a.p.block.HCLBlock().Body.MissingItemRange()
		a.error(hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Missing required attribute %q", a.Key),
				// This is the error word for word from 'terraform apply'
				Detail:  fmt.Sprintf("The argument %q is required, but no definition is found.", a.Key),
				Subject: &r,
				Extra:   nil,
			},
		})
	}

	return a
}

func (a *expectedAttribute) tryString() string {
	attr := a.p.block.GetAttribute(a.Key)
	if attr.IsNil() {
		return ""
	}

	if attr.Type() != cty.String {
		return ""
	}

	return attr.Value().AsString()
}

func (a *expectedAttribute) string() string {
	attr := a.p.block.GetAttribute(a.Key)
	if attr.IsNil() {
		return ""
	}

	if attr.Type() != cty.String {
		a.expectedTypeError(attr, "string")
		return ""
	}

	return attr.Value().AsString()
}

func (a *expectedAttribute) expectedTypeError(attr *terraform.Attribute, expectedType string) {
	var fn string
	if attr.IsNil() || attr.Type().Equals(cty.NilType) {
		fn = "nil"
	} else {
		fn = attr.Type().FriendlyName()
	}

	a.error(hcl.Diagnostics{
		{
			Severity:   hcl.DiagError,
			Summary:    "Invalid attribute type",
			Detail:     fmt.Sprintf("The attribute %q must be of type %q, found type %q", attr.Name(), expectedType, fn),
			Subject:    &attr.HCLAttribute().Range,
			Context:    &a.p.block.HCLBlock().DefRange,
			Expression: attr.HCLAttribute().Expr,

			EvalContext: a.p.block.Context().Inner(),
		},
	})
}
