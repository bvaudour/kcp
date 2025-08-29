package format

import (
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
		if standard.IsStandardFunction(node.Name) {
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
	done := make(map[string]bool)
	var result info.NodeInfoList
	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]
		r := depends[node.Id]
		var missing []string
		for _, d := range r.depends {
			if !done[d] {
				missing = append(missing, d)
			}
		}
		if len(nodes) == 0 || len(missing) == 0 {
			result = append(result, node)
			done[node.Name] = true
		}
		var before, after info.NodeInfoList
		d := make(map[string]bool)
		for k := range done {
			d[k] = true
		}
		for _, next := range nodes {
			if d[next.Name] || !slices.Contains(missing, next.Name) {
				after = append(after, next)
			} else {
				before = append(before, next)
			}
			d[next.Name] = true
		}
		if len(before) == 0 {
			result = append(result, node)
			done[node.Name] = true
		} else {
			nodes = append(before, node)
			nodes = append(nodes, after...)
		}
	}

	return result
}

func vOrder(nodes info.NodeInfoList, depends map[int]nodeReorder) info.NodeInfoList {
	nodes = vOrder0(nodes)
	return vOrder1(nodes, depends)
}

func Reorder(nodes info.NodeInfoList) (result info.NodeInfoList) {
	if len(nodes) < 2 {
		return nodes
	}

	rv, rf := make(map[int]nodeReorder), make(map[int]nodeReorder)
	var variables, functions info.NodeInfoList
	begin, _ := nodes[0].Position()
	r := nodeReorder{lines: begin.Line()}
	if nodes[0].Type != info.Function {
		r.depends = getDepends(nodes[0].Stmt.Cmd.(*syntax.CallExpr).Assigns[0])
		rv[nodes[0].Id] = r
		variables = append(variables, nodes[0])
	} else {
		rf[nodes[0].Id] = r
		functions = append(functions, nodes[0])
	}

	for i, node := range nodes[1:] {
		_, end := nodes[i].Position()
		_, begin := node.Position()
		r := nodeReorder{lines: begin.Line() - end.Line()}
		if node.Type != info.Function {
			r.depends = getDepends(node.Stmt.Cmd.(*syntax.CallExpr).Assigns[0])
			rv[node.Id] = r
			variables = append(variables, node)
		} else {
			rf[node.Id] = r
			functions = append(functions, node)
		}
	}

	variables, functions = vOrder(variables, rv), fOrder(functions)
	for i, v := range variables {
		begin, _ := v.Position()
		newPos := syntax.NewPos(0, 1, begin.Col())
		if i > 0 {
			_, end := variables[i-1].Position()
			newPos = syntax.NewPos(end.Offset()+1, end.Line()+1, begin.Col())
		}
		diff := position.Diff(begin, newPos)
		diff.Update(v.Stmt)
	}
	for i, f := range functions {
		begin, _ := f.Position()
		newPos := syntax.NewPos(0, 1, begin.Col())
		if i > 0 {
			_, end := functions[i-1].Position()
			newPos = syntax.NewPos(end.Offset()+1, end.Line()+1, begin.Col())
		} else if l := len(variables); l > 0 {
			_, end := variables[l-1].Position()
			newPos = syntax.NewPos(end.Offset()+1, end.Line()+1, begin.Col())
		}
		diff := position.Diff(begin, newPos)
		diff.Update(f.Stmt)
	}

	return append(variables, functions...)
}

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
