package format

import (
	"fmt"
	"io"
	"strings"

	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"mvdan.cc/sh/v3/syntax"
)

func Print(w io.Writer, nodes info.NodeInfoList, comments []syntax.Comment) error {
	var lastPos syntax.Pos
	for _, node := range nodes {
		begin, end := node.Position()
		d := int(begin.Line()) - int(lastPos.Line()) - 1
		if d > 0 {
			if _, err := fmt.Fprint(w, strings.Repeat("\n", d)); err != nil {
				return err
			}
		}
		indentSize := uint(4)
		if node.Type == info.ArrayVar {
			arrayExpr := node.Stmt.Cmd.(*syntax.CallExpr).Assigns[0].Array
			indentSize = arrayExpr.Lparen.Col()
		}
		pr := syntax.NewPrinter(syntax.Indent(indentSize))
		if err := pr.Print(w, node.Stmt); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
		lastPos = end
	}

	for _, comment := range comments {
		pos := comment.Hash
		l := int(pos.Line()) - int(lastPos.Line()) - 1
		c := int(pos.Col()) - 1
		s := ""
		if l > 0 {
			s = strings.Repeat("\n", l)
		}
		if c > 0 {
			s = strings.Repeat(" ", c)
		}
		s += "#"
		if _, err := fmt.Fprint(w, s); err != nil {
			return err
		}
		lastPos = comment.End()
	}

	return nil
}
