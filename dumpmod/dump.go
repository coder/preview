package dumpmod

import (
	"bufio"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/aquasecurity/trivy/pkg/iac/terraform"
	tfcontext "github.com/aquasecurity/trivy/pkg/iac/terraform/context"
	"github.com/awalterschulze/gographviz"
	gotree "github.com/d6o/GoTree"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type ctxMeta struct {
	Block *terraform.Block
	Name  string
}

func GraphvizDump(mods terraform.Modules) (*gographviz.Graph, error) {
	g := gographviz.NewGraph()

	ctxPerBlock := make(map[*hcl.EvalContext]ctxMeta)
	for _, mod := range mods {
		modRoot := fmt.Sprintf("%q", mod.ModulePath())

		err := g.AddNode(fmt.Sprintf("%q", mod.RootPath()), modRoot, map[string]string{})
		if err != nil {
			return nil, fmt.Errorf("add module node: %w", err)
		}

		for _, block := range mod.GetBlocks() {
			//for _, cb := range block.AllBlocks() {
			cb := block
			meta := ctxMeta{
				Block: cb,
				Name:  fmt.Sprintf("%q", block.FullName()),
			}
			ctxPerBlock[cb.Context().Inner()] = meta

			err = g.AddNode(modRoot, meta.Name, map[string]string{})
			if err != nil {
				return nil, fmt.Errorf("add ctx node: %w", err)
			}
			//}
		}
	}

	ctxName := func(ctx *tfcontext.Context) string {
		if meta, ok := ctxPerBlock[ctx.Inner()]; ok {
			return meta.Name
		}
		return fmt.Sprintf("\"%p\"", ctx.Inner())
	}

	for _, mod := range mods {
		modRoot := fmt.Sprintf("%q", mod.ModulePath())
		for _, block := range mod.GetBlocks() {
			//rctx := root(cb.Context())
			//if rctx != nil {
			//	rn := fmt.Sprintf("%q", ctxName(rctx))
			//	if !g.IsNode(rn) {
			//		err = g.AddNode(modRoot, rn, map[string]string{})
			//		if err != nil {
			//			return nil, fmt.Errorf("add root ctx node: %w", err)
			//		}
			//	}
			//}

			cn := ctxName(block.Context())
			if !g.IsNode(cn) {
				return nil, fmt.Errorf("node %q not found", cn)
			}

			par := block.Context().Parent()
			pn := ctxName(par)
			lpn := pn

			for !g.IsNode(lpn) {
				g.AddNode(modRoot, lpn, map[string]string{})
				par = par.Parent()
				if par == nil {
					break
				}
				lpn = ctxName(par)
				g.AddEdge(lpn, pn, true, map[string]string{})
			}

			err := g.AddEdge(pn, cn, true, map[string]string{})
			if err != nil {
				return nil, fmt.Errorf("add ctx edge: %w", err)
			}
		}
	}
	var _ = g
	fmt.Println("di" + g.String())

	return g, nil
}

func Dump(files map[string]*hcl.File, mods terraform.Modules) error {
	var out strings.Builder

	for _, mod := range mods {
		// TODO: Use module path
		out.WriteString(fmt.Sprintf("# --- %s ---\n", mod.RootPath()))
		for _, block := range mod.GetBlocks() {
			m := block.GetMetadata()

			lines, err := readLinesFromFile(m.Range().GetFS(), m.Range().GetLocalFilename(), m.Range().GetStartLine(), m.Range().GetEndLine())
			if err != nil {
				return fmt.Errorf("read file %q: %w", m.Range().GetLocalFilename(), err)
			}
			out.WriteString(fmt.Sprintf("# Ctx: %s\n", ctxName(block.Context())))
			out.WriteString(fmt.Sprintf("# %s\n", treeNode(block.Context())))

			out.WriteString(strings.Join(lines, "\n") + "\n\n")
		}
	}

	fmt.Println(dumpContexts())
	fmt.Println(out.String())
	return nil
}

var (
	cnt      = 0
	ctxNames = make(map[*tfcontext.Context]string)
)

func dumpContexts() string {
	trees := make(map[*tfcontext.Context]gotree.Tree)
	for ctx, _ := range ctxNames {
		if ctx == nil {
			continue
		}
		par := ctx.Parent()

		ptree, ok := trees[par]
		if !ok {
			ptree = gotree.New(treeNode(par))
			trees[par] = ptree
		}

		ctxTree, ok := trees[ctx]
		if !ok {
			ctxTree = gotree.New(treeNode(ctx))
			trees[ctx] = ctxTree
		}

		ptree.AddTree(ctxTree)
	}

	rootTree := gotree.New("Contexts")
	for _, tree := range trees {
		rootTree.AddTree(tree)
	}
	return rootTree.Print()
}

func treeNode(ctx *tfcontext.Context) string {
	var str strings.Builder
	str.WriteString(ctxName(ctx))
	if ctx == nil {
		str.WriteString("\n<no context>")
		return str.String()
	}
	if ctx.Inner() == nil {
		str.WriteString("\n<no inner>")
		return str.String()
	}
	if ctx.Inner().Variables == nil {
		str.WriteString("\n<no variables>")
		return str.String()
	}

	if len(ctx.Inner().Variables) == 0 {
		return str.String()
	}
	str.WriteString("\n")

	v := cty.ObjectVal(ctx.Inner().Variables)
	v = removeUnknownValues(v)

	d, err := ctyjson.Marshal(v, v.Type())
	if err != nil {
		panic(err)
	}
	str.WriteString(string(d))
	return str.String()
}

func root(ctx *tfcontext.Context) *tfcontext.Context {
	if ctx.Parent() != nil {
		return root(ctx.Parent())
	}
	return ctx
}

func ctxName(ctx *tfcontext.Context) string {
	if name, ok := ctxNames[ctx]; ok {
		return name
	}
	ctxNames[ctx] = fmt.Sprintf("[%p] %d", ctx, cnt)
	cnt++
	return ctxNames[ctx]
}

func readLinesFromFile(fsys fs.FS, path string, from, to int) ([]string, error) {
	slashedPath := strings.TrimPrefix(filepath.ToSlash(path), "/")

	file, err := fsys.Open(slashedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from result filesystem: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rawLines := make([]string, 0, to-from+1)

	for lineNum := 0; scanner.Scan() && lineNum < to; lineNum++ {
		if lineNum >= from-1 {
			rawLines = append(rawLines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return rawLines, nil
}

func removeUnknownValues(val cty.Value) cty.Value {
	if !val.IsKnown() {
		// If the value itself is unknown, return cty.NilVal (or an empty value of the same type)
		return cty.StringVal("unknown")
	}

	switch {
	case val.Type().IsObjectType() || val.Type().IsMapType():
		newMap := make(map[string]cty.Value)
		for it := val.ElementIterator(); it.Next(); {
			key, elem := it.Element()
			if elem.IsWhollyKnown() {
				newMap[key.AsString()] = removeUnknownValues(elem)
			}
		}
		return cty.ObjectVal(newMap)

	case val.Type().IsTupleType() || val.Type().IsListType() || val.Type().IsSetType():
		var newList []cty.Value
		for it := val.ElementIterator(); it.Next(); {
			_, elem := it.Element()
			if elem.IsWhollyKnown() {
				newList = append(newList, removeUnknownValues(elem))
			}
		}

		if val.Type().IsTupleType() {
			return cty.TupleVal(newList)
		}

		if len(newList) == 0 {
			ty := val.Type().ElementType()
			return cty.ListValEmpty(ty)
		}
		return cty.ListVal(newList)
	default:
		// If it's a primitive type, return as is
		return val
	}
}
