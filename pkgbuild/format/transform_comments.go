package format

import (
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"mvdan.cc/sh/v3/syntax"
)

func RemoveOuterComments(node *info.NodeInfo) *info.NodeInfo {
	toPos, _ := node.Position()
	fromPos, _ := node.InnerPosition()
	diff := position.Diff(fromPos, toPos)

	node.Stmt.Comments = node.Stmt.Comments[:0]
	diff.Update(node.Stmt)

	return node
}

func commentsDiff(comments []syntax.Comment) (diff *position.PosDiff) {
	if l := len(comments); l > 0 {
		diff = position.Diff(comments[l-1].End(), comments[0].Pos())
	}
	return
}

func RemoveInnerComments(node *info.NodeInfo) *info.NodeInfo {
	if node.Type != info.ArrayVar {
		return node
	}

	values := node.Stmt.Cmd.(*syntax.CallExpr).Assigns[0].Array
	diff := commentsDiff(values.Last)
	diff.Update(node.Stmt)
	for _, v := range values.Elems {
		diff = commentsDiff(v.Comments)
		diff.Update(node.Stmt)
	}

	return node
}
