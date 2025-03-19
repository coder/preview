package hclext

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type ReferenceBuilder []hcl.Traverser

func NewReferenceBuilder(path ...string) *ReferenceBuilder {
	return (&ReferenceBuilder{}).AddPath(path...)
}

func (b *ReferenceBuilder) AddPath(path ...string) *ReferenceBuilder {
	for _, p := range path {
		if len(*b) == 0 {
			*b = append(*b, hcl.TraverseRoot{Name: p})
		} else {
			*b = append(*b, hcl.TraverseAttr{Name: p})
		}
	}
	return b
}

func (b *ReferenceBuilder) AddIndex(idx int) *ReferenceBuilder {
	*b = append(*b, hcl.TraverseIndex{Key: cty.NumberIntVal(int64(idx))})
	return b
}

func (b *ReferenceBuilder) AddKey(key string) *ReferenceBuilder {
	*b = append(*b, hcl.TraverseIndex{Key: cty.StringVal(key)})
	return b
}

func (b ReferenceBuilder) Expression() hcl.Expression {
	return &hclsyntax.ScopeTraversalExpr{Traversal: hcl.Traversal(b)}
}

func ReferenceNames(exp hcl.Expression) []string {
	if exp == nil {
		return []string{}
	}
	allVars := exp.Variables()
	vars := make([]string, 0, len(allVars))

	for _, v := range allVars {
		vars = append(vars, CreateDotReferenceFromTraversal(v))
	}

	return vars
}

func CreateDotReferenceFromTraversal(traversals ...hcl.Traversal) string {
	var refParts []string

	for _, x := range traversals {
		for _, p := range x {
			switch part := p.(type) {
			case hcl.TraverseRoot:
				refParts = append(refParts, part.Name)
			case hcl.TraverseAttr:
				refParts = append(refParts, part.Name)
			case hcl.TraverseIndex:
				if part.Key.Type().Equals(cty.String) {
					refParts = append(refParts, fmt.Sprintf("[%s]", part.Key.AsString()))
				} else if part.Key.Type().Equals(cty.Number) {
					idx, _ := part.Key.AsBigFloat().Int64()
					refParts = append(refParts, fmt.Sprintf("[%d]", idx))
				} else {
					refParts = append(refParts, fmt.Sprintf("[?? %q]", part.Key.Type().FriendlyName()))
				}
			}
		}
	}
	return strings.Join(refParts, ".")
}
