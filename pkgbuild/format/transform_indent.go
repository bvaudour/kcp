package format

import (
	"strings"

	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"mvdan.cc/sh/v3/syntax"
)

// @TODO: unused for now: indentation are made at Print stage.
func IndentFunctions(spaces uint) TransformFunc {
	return func(node *info.NodeInfo) *info.NodeInfo {
		if node.Type != info.Function {
			return node
		}

		printer := syntax.NewPrinter(syntax.Indent(spaces))
		var sb strings.Builder
		if printer.Print(&sb, node.Stmt) != nil {
			return node
		}

		parser := syntax.NewParser(syntax.KeepComments(true))
		f, err := parser.Parse(strings.NewReader(sb.String()), "")
		if err != nil || len(f.Stmts) == 0 {
			return node
		}
		newNode := &info.NodeInfo{
			Type: node.Type,
			Stmt: f.Stmts[0],
		}
		oldBegin, _ := node.Position()
		newBegin, _ := newNode.Position()
		diff := position.Diff(oldBegin, newBegin)
		diff.Update(newNode.Stmt)
		node.Stmt = newNode.Stmt

		return node
	}
}

func getNextArrayElem(element *syntax.ArrayElem, currentPos syntax.Pos) (syntax.Node, bool) {
	var value *syntax.Word
	var comment *syntax.Comment
	if position.Cmp(element.Value.Pos(), currentPos) >= 0 {
		value = element.Value
	}

	for i, c := range element.Comments {
		if position.Cmp(c.Pos(), currentPos) >= 0 {
			comment = &element.Comments[i]
			break
		}
	}

	if value != nil {
		if comment != nil && position.Cmp(comment.Pos(), value.Pos()) < 0 {
			return comment, true
		}
		return value, false
	} else if comment != nil {
		return comment, true
	}
	return nil, false
}

// IndentVariables aligns correctly array variable assignations.
func IndentVariables(node *info.NodeInfo) *info.NodeInfo {
	if node.Type == info.Function {
		return node
	}
	begin, _ := node.InnerPosition()
	l, c, o := begin.Line(), begin.Col(), begin.Offset()
	if c != 1 {
		newBegin := syntax.NewPos(o-(c-1), l, 1)
		diff := position.Diff(begin, newBegin)
		diff.Update(node.Stmt)
	}

	if node.Type != info.ArrayVar {
		return node
	}

	array := node.Stmt.Cmd.(*syntax.CallExpr).Assigns[0].Array
	inc := position.PosDiff{Col: 1, Offset: 1}
	posL := inc.AddTo(array.Lparen)
	currentPos := posL
	var isLastElemComment bool
	for i, element := range array.Elems {
		next, comment := getNextArrayElem(element, currentPos)
		for next != nil {
			isBeginOfLine := currentPos.Col() == posL.Col()
			if !isBeginOfLine {
				currentPos = inc.AddTo(currentPos)
			}
			p := next.Pos()
			d := position.Diff(p, currentPos)
			e := d.Mv(next.End())
			if e.Col() > 80 && !isBeginOfLine {
				currentPos = syntax.NewPos(currentPos.Offset()+posL.Col(), currentPos.Line()+1, posL.Col())
				d = position.Diff(p, currentPos)
			}
			position.Update(d, array.Elems[i:])
			d.UpdateComments(array.Last)
			currentPos = next.End()
			isLastElemComment = comment
			next, comment = getNextArrayElem(element, currentPos)
		}
	}

	if isLastElemComment {
		currentPos = syntax.NewPos(currentPos.Offset()+posL.Col(), currentPos.Line()+1, posL.Col())
	}
	d := position.Diff(array.Rparen, currentPos)
	d.MvAll(&array.Rparen).UpdateComments(array.Last)

	return node
}
