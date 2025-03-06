package hclext

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

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
