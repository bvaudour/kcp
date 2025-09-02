package format

import (
	"fmt"
	"slices"

	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"codeberg.org/bvaudour/kcp/pkgbuild/standard"
	"git.kaosx.ovh/benjamin/collection"
	"mvdan.cc/sh/v3/syntax"
)

func getArithmeticDepends(arithm syntax.ArithmExpr) (depends []string) {
	switch e := arithm.(type) {
	case *syntax.BinaryArithm:
		depends = append(depends, getArithmeticDepends(e.X)...)
		depends = append(depends, getArithmeticDepends(e.Y)...)
	case *syntax.UnaryArithm:
		depends = append(depends, getArithmeticDepends(e.X)...)
	case *syntax.ParenArithm:
		depends = append(depends, getArithmeticDepends(e.X)...)
	case *syntax.Word:
		depends = append(depends, getWordDepends(e)...)
	}

	return
}

func getPartsDepends(parts []syntax.WordPart) (depends []string) {
	for _, part := range parts {
		switch p := part.(type) {
		case *syntax.ParamExp:
			depends = append(depends, p.Param.Value)
		case *syntax.ArithmExp:
			depends = append(depends, getArithmeticDepends(p.X)...)
		case *syntax.DblQuoted:
			depends = append(depends, getPartsDepends(p.Parts)...)
		}
	}

	return
}

func getWordDepends(word *syntax.Word) (depends []string) {
	return getPartsDepends(word.Parts)
}

func getDepends(assign *syntax.Assign) (depends []string) {
	var values []*syntax.Word
	if assign.Value != nil {
		values = append(values, assign.Value)
	} else {
		values = make([]*syntax.Word, len(assign.Array.Elems))
		for i, elem := range assign.Array.Elems {
			values[i] = elem.Value
		}
	}

	for _, word := range values {
		depends = append(depends, getWordDepends(word)...)
	}

	slices.Sort(depends)
	return slices.Compact(depends)
}

type nodeReorder struct {
	lines   uint
	depends []string
}

func newNr(node *info.NodeInfo, lastPos syntax.Pos) (nr nodeReorder, nextPos syntax.Pos) {
	begin, end := node.Position()
	if lastPos.IsValid() {
		l := int(begin.Line()) - int(lastPos.Line()) - 1
		if l > 0 {
			nr.lines = uint(l)
		}
	}
	if node.Type == info.ArrayVar || node.Type == info.SingleVar {
		assign := node.Stmt.Cmd.(*syntax.CallExpr).Assigns[0]
		nr.depends = getDepends(assign)
	}

	nextPos = end
	return
}

func (nr nodeReorder) String() string {
	return fmt.Sprintf(`{lines: %d, depends: %v}`, nr.lines, nr.depends)
}

func prepareReorder(nodes info.NodeInfoList) (variables, functions info.NodeInfoList, depends map[int]nodeReorder) {
	depends = make(map[int]nodeReorder)
	var nr nodeReorder
	var lastPos syntax.Pos
	for _, node := range nodes {
		nr, lastPos = newNr(node, lastPos)
		depends[node.Id] = nr
		if node.Type == info.Function {
			functions = append(functions, node)
		} else {
			variables = append(variables, node)
		}
	}

	return
}

func fOrder(nodes info.NodeInfoList) info.NodeInfoList {
	var fs, fu info.NodeInfoList
	ms := make(map[string]info.NodeInfoList)
	for _, node := range nodes {
		if standard.IsStandardFunction(node.Name) {
			ms[node.Name] = append(ms[node.Name], node)
		} else {
			fu = append(fu, node)
		}
	}

	names := standard.GetFunctions()
	for _, n := range names {
		if e, ok := ms[n]; ok {
			fs = append(fs, e...)
		}
	}

	return append(fu, fs...)
}

func vOrder0(nodes info.NodeInfoList) info.NodeInfoList {
	var vs, vu info.NodeInfoList
	ms := make(map[string]info.NodeInfoList)
	for _, node := range nodes {
		if standard.IsStandardVariable(node.Name) {
			ms[node.Name] = append(ms[node.Name], node)
		} else {
			vu = append(vu, node)
		}
	}

	names := standard.GetVariables()
	for _, n := range names {
		if e, ok := ms[n]; ok {
			vs = append(vs, e...)
		}
	}

	return append(vs, vu...)
}

func vOrder1(nodes info.NodeInfoList, depends map[int]nodeReorder) info.NodeInfoList {
	done := collection.NewSet[string]()
	var result info.NodeInfoList
	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]
		r := depends[node.Id]
		var missing []string
		for _, d := range r.depends {
			if !done.Contains(d) {
				missing = append(missing, d)
			}
		}
		if len(nodes) == 0 || len(missing) == 0 {
			result = append(result, node)
			done.Add(node.Name)
			continue
		}
		var before, after info.NodeInfoList
		d := collection.NewSet(done.ToSlice()...)
		for _, next := range nodes {
			if d.Contains(next.Name) || !slices.Contains(missing, next.Name) {
				after = append(after, next)
			} else {
				before = append(before, next)
			}
			d.Add(next.Name)
		}
		if len(before) == 0 {
			result = append(result, node)
			done.Add(node.Name)
		} else {
			before = append(before, node)
			before = append(before, after...)
			nodes = before
		}
	}

	return result
}

func vOrder(nodes info.NodeInfoList, depends map[int]nodeReorder) info.NodeInfoList {
	nodes = vOrder0(nodes)
	return vOrder1(nodes, depends)
}

// Reorder reorders variable assignations and function declarations in a PKGBUILD.
func Reorder(nodes info.NodeInfoList) (result info.NodeInfoList) {
	if len(nodes) < 2 {
		return nodes
	}

	variables, functions, depends := prepareReorder(nodes)

	functions = fOrder(functions)
	variables = vOrder(variables, depends)
	result = append(result, variables...)
	result = append(result, functions...)

	var lastPos syntax.Pos
	initialPos, _ := nodes[0].Position()
	for _, node := range result {
		nr := depends[node.Id]
		var newPos syntax.Pos
		begin, _ := node.Position()
		if !lastPos.IsValid() {
			newPos = initialPos
		} else {
			newPos = syntax.NewPos(lastPos.Offset()+nr.lines+1, lastPos.Line()+nr.lines+1, begin.Col())
		}
		diff := position.Diff(begin, newPos)
		diff.Update(node.Stmt)
		_, lastPos = node.Position()
	}

	return
}

// FormatBlankLines removes useless blank lines and add one blank line before
// each function declaration.
func FormatBlankLines(keepFirstBlank bool) TransformListFunc {
	return func(nodes info.NodeInfoList) info.NodeInfoList {
		if len(nodes) == 0 {
			return nodes
		}

		newPos := syntax.NewPos(0, 1, 1)
		begin, _ := nodes[0].Position()
		if begin.Line() > 1 && keepFirstBlank {
			newPos = syntax.NewPos(1, 2, 1)
		}
		diff := position.Diff(begin, newPos)
		diff.Update(nodes[0].Stmt)
		_, currentPos := nodes[0].Position()

		for _, node := range nodes[1:] {
			newPos = syntax.NewPos(currentPos.Offset()+1, currentPos.Line()+1, 1)
			if node.Type == info.Function {
				newPos = syntax.NewPos(currentPos.Offset()+2, currentPos.Line()+2, 1)
			}
			begin, _ := node.Position()
			diff := position.Diff(begin, newPos)
			diff.Update(node.Stmt)
			_, currentPos = node.Position()
		}

		return nodes
	}
}

// RemoveHeader removes the first comment of the first node if needed.
func RemoveHeader(nodes info.NodeInfoList) info.NodeInfoList {
	if len(nodes) == 0 {
		return nodes
	}

	beginComment, _ := nodes[0].Position()
	begin, _ := nodes[0].InnerPosition()
	diff := position.Diff(begin, beginComment)
	var comments []syntax.Comment
	for _, c := range nodes[0].Stmt.Comments {
		if position.Cmp(c.Pos(), begin) > 0 {
			comments = append(comments, c)
		}
	}
	nodes[0].Stmt.Comments = comments

	for _, node := range nodes {
		diff.Update(node.Stmt)
	}

	return nodes
}

// RemoveDuplicates remove duplicates variables and functions.
// It keeps the last element.
func RemoveDuplicates(nodes info.NodeInfoList) info.NodeInfoList {
	duplicates := nodes.GetDuplicates()
	ids := collection.NewSet[int]()
	for _, nd := range duplicates {
		for _, n := range nd[:len(nd)-1] {
			ids.Add(n.Id)
		}
	}

	var newNodes info.NodeInfoList
	l := len(nodes) - 1
	var diff *position.PosDiff
	for i, n := range nodes {
		if diff != nil {
			diff.Update(n.Stmt)
		}
		if !ids.Contains(n.Id) {
			newNodes = append(newNodes, n)
			continue
		}
		if i < l {
			toPos, _ := n.Position()
			fromPos, _ := nodes[i+1].Position()
			diff = position.Diff(fromPos, toPos)
		}
	}

	return newNodes
}
