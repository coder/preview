package types

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

func ReferenceNames(exp hcl.Expression) []string {
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
				refParts = append(refParts, fmt.Sprintf("[%s]", part.Key.AsString()))
			}
		}
	}
	return strings.Join(refParts, ".")
}
