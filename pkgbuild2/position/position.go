package position

import (
	"mvdan.cc/sh/v3/syntax"
)

func sign(i int) int {
	if i > 0 {
		return i / i
	}
	if i < 0 {
		return -i / i
	}
	return 0
}

func Cmp(p1, p2 syntax.Pos) int {
	if c := sign(int(p1.Line()) - int(p2.Line())); c != 0 {
		return c
	}
	return sign(int(p1.Col()) - int(p2.Col()))
}

func IncLine(p syntax.Pos, i int) syntax.Pos {
	if !p.IsValid() || i == 0 {
		return p
	}
	if i > 0 {
		return syntax.NewPos(p.Offset()+uint(i), p.Line()+uint(i), p.Col())
	}
	return syntax.NewPos(p.Offset()-uint(-i), p.Line()-uint(-i), p.Col())
}

func IncCol(p syntax.Pos, i int, l uint) syntax.Pos {
	if !p.IsValid() || i == 0 || p.Line() != l {
		return p
	}
	if i > 0 {
		return syntax.NewPos(p.Offset()+uint(i), p.Line(), p.Col()+uint(i))
	}
	return syntax.NewPos(p.Offset()-uint(-i), p.Line(), p.Col()+uint(-i))
}

func IncLineNode(n syntax.Node, i int) {
	switch n.(type) {
	case *syntax.File:
		e := n.(*syntax.File)
		if e == nil {
			return
		}
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.Stmt:
		e := n.(*syntax.Stmt)
		if e == nil {
			return
		}
		e.Position, e.Semicolon = IncLine(e.Position, i), IncLine(e.Semicolon, i)
		IncLineNode(e.Cmd, i)
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncLine(ee.Hash, i)
		}
		for _, ee := range e.Redirs {
			IncLineNode(ee, i)
		}
	case *syntax.CallExpr:
		e := n.(*syntax.CallExpr)
		if e == nil {
			return
		}
		for _, ee := range e.Args {
			IncLineNode(ee, i)
		}
		for _, ee := range e.Assigns {
			IncLineNode(ee, i)
		}
	case *syntax.Word:
		e := n.(*syntax.Word)
		if e == nil {
			return
		}
		for _, ee := range e.Parts {
			IncLineNode(ee, i)
		}
	case *syntax.Lit:
		e := n.(*syntax.Lit)
		if e == nil {
			return
		}
		e.ValuePos, e.ValueEnd = IncLine(e.ValuePos, i), IncLine(e.ValueEnd, i)
	case *syntax.SglQuoted:
		e := n.(*syntax.SglQuoted)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
	case *syntax.DblQuoted:
		e := n.(*syntax.DblQuoted)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
		for _, ee := range e.Parts {
			IncLineNode(ee, i)
		}
	case *syntax.ParamExp:
		e := n.(*syntax.ParamExp)
		if e == nil {
			return
		}
		e.Dollar, e.Rbrace = IncLine(e.Dollar, i), IncLine(e.Rbrace, i)
		IncLineNode(e.Param, i)
		IncLineNode(e.Index, i)
		if e.Slice != nil {
			IncLineNode(e.Slice.Offset, i)
			IncLineNode(e.Slice.Length, i)
		}
		if e.Repl != nil {
			IncLineNode(e.Repl.Orig, i)
			IncLineNode(e.Repl.With, i)
		}
		if e.Exp != nil {
			IncLineNode(e.Exp.Word, i)
		}
	case *syntax.BinaryArithm:
		e := n.(*syntax.BinaryArithm)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.X, i)
		IncLineNode(e.Y, i)
	case *syntax.UnaryArithm:
		e := n.(*syntax.UnaryArithm)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.X, i)
	case *syntax.ParenArithm:
		e := n.(*syntax.ParenArithm)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncLine(e.Lparen, i), IncLine(e.Rparen, i)
		IncLineNode(e.X, i)
	case *syntax.CmdSubst:
		e := n.(*syntax.CmdSubst)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.ArithmExp:
		e := n.(*syntax.ArithmExp)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
		IncLineNode(e.X, i)
	case *syntax.ProcSubst:
		e := n.(*syntax.ProcSubst)
		if e == nil {
			return
		}
		e.OpPos, e.Rparen = IncLine(e.OpPos, i), IncLine(e.Rparen, i)
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.ExtGlob:
		e := n.(*syntax.ExtGlob)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.Pattern, i)
	case *syntax.Assign:
		e := n.(*syntax.Assign)
		if e == nil {
			return
		}
		IncLineNode(e.Name, i)
		IncLineNode(e.Index, i)
		IncLineNode(e.Value, i)
		IncLineNode(e.Array, i)
	case *syntax.ArrayExpr:
		e := n.(*syntax.ArrayExpr)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncLine(e.Lparen, i), IncLine(e.Rparen, i)
		for _, ee := range e.Elems {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.ArrayElem:
		e := n.(*syntax.ArrayElem)
		if e == nil {
			return
		}
		IncLineNode(e.Index, i)
		IncLineNode(e.Value, i)
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.IfClause:
		e := n.(*syntax.IfClause)
		if e == nil {
			return
		}
		e.Position, e.ThenPos, e.FiPos = IncLine(e.Position, i), IncLine(e.ThenPos, i), IncLine(e.FiPos, i)
		IncLineNode(e.Else, i)
		for _, ee := range e.Cond {
			IncLineNode(ee, i)
		}
		for _, ee := range e.Then {
			IncLineNode(ee, i)
		}
		for i, ee := range e.CondLast {
			e.CondLast[i].Hash = IncLine(ee.Hash, i)
		}
		for i, ee := range e.ThenLast {
			e.ThenLast[i].Hash = IncLine(ee.Hash, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.WhileClause:
		e := n.(*syntax.WhileClause)
		if e == nil {
			return
		}
		e.WhilePos, e.DoPos, e.DonePos = IncLine(e.WhilePos, i), IncLine(e.DoPos, i), IncLine(e.DonePos, i)
		for _, ee := range e.Cond {
			IncLineNode(ee, i)
		}
		for _, ee := range e.Do {
			IncLineNode(ee, i)
		}
		for i, ee := range e.CondLast {
			e.CondLast[i].Hash = IncLine(ee.Hash, i)
		}
		for i, ee := range e.DoLast {
			e.DoLast[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.ForClause:
		e := n.(*syntax.ForClause)
		if e == nil {
			return
		}
		e.ForPos, e.DoPos, e.DonePos = IncLine(e.ForPos, i), IncLine(e.DoPos, i), IncLine(e.DonePos, i)
		IncLineNode(e.Loop, i)
		for _, ee := range e.Do {
			IncLineNode(ee, i)
		}
		for i, ee := range e.DoLast {
			e.DoLast[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.WordIter:
		e := n.(*syntax.WordIter)
		if e == nil {
			return
		}
		e.InPos = IncLine(e.InPos, i)
		IncLineNode(e.Name, i)
		for _, ee := range e.Items {
			IncLineNode(ee, i)
		}
	case *syntax.CStyleLoop:
		e := n.(*syntax.CStyleLoop)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncLine(e.Lparen, i), IncLine(e.Rparen, i)
		IncLineNode(e.Init, i)
		IncLineNode(e.Cond, i)
		IncLineNode(e.Post, i)
	case *syntax.CaseClause:
		e := n.(*syntax.CaseClause)
		if e == nil {
			return
		}
		e.Case, e.In, e.Esac = IncLine(e.Case, i), IncLine(e.In, i), IncLine(e.Esac, i)
		IncLineNode(e.Word, i)
		for _, ee := range e.Items {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.CaseItem:
		e := n.(*syntax.CaseItem)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		for _, ee := range e.Patterns {
			IncLineNode(ee, i)
		}
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncLine(ee.Hash, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.Block:
		e := n.(*syntax.Block)
		if e == nil {
			return
		}
		e.Lbrace, e.Rbrace = IncLine(e.Lbrace, i), IncLine(e.Rbrace, i)
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.Subshell:
		e := n.(*syntax.Subshell)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncLine(e.Lparen, i), IncLine(e.Rparen, i)
		for _, ee := range e.Stmts {
			IncLineNode(ee, i)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncLine(ee.Hash, i)
		}
	case *syntax.BinaryCmd:
		e := n.(*syntax.BinaryCmd)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.X, i)
		IncLineNode(e.Y, i)
	case *syntax.FuncDecl:
		e := n.(*syntax.FuncDecl)
		if e == nil {
			return
		}
		e.Position = IncLine(e.Position, i)
		IncLineNode(e.Name, i)
		IncLineNode(e.Body, i)
	case *syntax.ArithmCmd:
		e := n.(*syntax.ArithmCmd)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
		IncLineNode(e.X, i)
	case *syntax.TestClause:
		e := n.(*syntax.TestClause)
		if e == nil {
			return
		}
		e.Left, e.Right = IncLine(e.Left, i), IncLine(e.Right, i)
		IncLineNode(e.X, i)
	case *syntax.BinaryTest:
		e := n.(*syntax.BinaryTest)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.X, i)
		IncLineNode(e.Y, i)
	case *syntax.UnaryTest:
		e := n.(*syntax.UnaryTest)
		if e == nil {
			return
		}
		e.OpPos = IncLine(e.OpPos, i)
		IncLineNode(e.X, i)
	case *syntax.ParenTest:
		e := n.(*syntax.ParenTest)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncLine(e.Lparen, i), IncLine(e.Rparen, i)
		IncLineNode(e.X, i)
	case *syntax.DeclClause:
		e := n.(*syntax.DeclClause)
		if e == nil {
			return
		}
		IncLineNode(e.Variant, i)
		for _, ee := range e.Args {
			IncLineNode(ee, i)
		}
	case *syntax.LetClause:
		e := n.(*syntax.LetClause)
		if e == nil {
			return
		}
		e.Let = IncLine(e.Let, i)
		for _, ee := range e.Exprs {
			IncLineNode(ee, i)
		}
	case *syntax.TimeClause:
		e := n.(*syntax.TimeClause)
		if e == nil {
			return
		}
		e.Time = IncLine(e.Time, i)
		IncLineNode(e.Stmt, i)
	case *syntax.CoprocClause:
		e := n.(*syntax.CoprocClause)
		if e == nil {
			return
		}
		e.Coproc = IncLine(e.Coproc, i)
		IncLineNode(e.Name, i)
		IncLineNode(e.Stmt, i)
	}
}

func IncColNode(n syntax.Node, i int, l uint) {
	switch n.(type) {
	case *syntax.File:
		e := n.(*syntax.File)
		if e == nil {
			return
		}
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.Stmt:
		e := n.(*syntax.Stmt)
		if e == nil {
			return
		}
		e.Position, e.Semicolon = IncCol(e.Position, i, l), IncCol(e.Semicolon, i, l)
		IncColNode(e.Cmd, i, l)
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncCol(ee.Hash, i, l)
		}
		for _, ee := range e.Redirs {
			IncColNode(ee, i, l)
		}
	case *syntax.CallExpr:
		e := n.(*syntax.CallExpr)
		if e == nil {
			return
		}
		for _, ee := range e.Args {
			IncColNode(ee, i, l)
		}
		for _, ee := range e.Assigns {
			IncColNode(ee, i, l)
		}
	case *syntax.Word:
		e := n.(*syntax.Word)
		if e == nil {
			return
		}
		for _, ee := range e.Parts {
			IncColNode(ee, i, l)
		}
	case *syntax.Lit:
		e := n.(*syntax.Lit)
		if e == nil {
			return
		}
		e.ValuePos, e.ValueEnd = IncCol(e.ValuePos, i, l), IncCol(e.ValueEnd, i, l)
	case *syntax.SglQuoted:
		e := n.(*syntax.SglQuoted)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
	case *syntax.DblQuoted:
		e := n.(*syntax.DblQuoted)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
		for _, ee := range e.Parts {
			IncColNode(ee, i, l)
		}
	case *syntax.ParamExp:
		e := n.(*syntax.ParamExp)
		if e == nil {
			return
		}
		e.Dollar, e.Rbrace = IncCol(e.Dollar, i, l), IncCol(e.Rbrace, i, l)
		IncColNode(e.Param, i, l)
		IncColNode(e.Index, i, l)
		if e.Slice != nil {
			IncColNode(e.Slice.Offset, i, l)
			IncColNode(e.Slice.Length, i, l)
		}
		if e.Repl != nil {
			IncColNode(e.Repl.Orig, i, l)
			IncColNode(e.Repl.With, i, l)
		}
		if e.Exp != nil {
			IncColNode(e.Exp.Word, i, l)
		}
	case *syntax.BinaryArithm:
		e := n.(*syntax.BinaryArithm)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.X, i, l)
		IncColNode(e.Y, i, l)
	case *syntax.UnaryArithm:
		e := n.(*syntax.UnaryArithm)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.X, i, l)
	case *syntax.ParenArithm:
		e := n.(*syntax.ParenArithm)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncCol(e.Lparen, i, l), IncCol(e.Rparen, i, l)
		IncColNode(e.X, i, l)
	case *syntax.CmdSubst:
		e := n.(*syntax.CmdSubst)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.ArithmExp:
		e := n.(*syntax.ArithmExp)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
		IncColNode(e.X, i, l)
	case *syntax.ProcSubst:
		e := n.(*syntax.ProcSubst)
		if e == nil {
			return
		}
		e.OpPos, e.Rparen = IncCol(e.OpPos, i, l), IncCol(e.Rparen, i, l)
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.ExtGlob:
		e := n.(*syntax.ExtGlob)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.Pattern, i, l)
	case *syntax.Assign:
		e := n.(*syntax.Assign)
		if e == nil {
			return
		}
		IncColNode(e.Name, i, l)
		IncColNode(e.Index, i, l)
		IncColNode(e.Value, i, l)
		IncColNode(e.Array, i, l)
	case *syntax.ArrayExpr:
		e := n.(*syntax.ArrayExpr)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncCol(e.Lparen, i, l), IncCol(e.Rparen, i, l)
		for _, ee := range e.Elems {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.ArrayElem:
		e := n.(*syntax.ArrayElem)
		if e == nil {
			return
		}
		IncColNode(e.Index, i, l)
		IncColNode(e.Value, i, l)
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.IfClause:
		e := n.(*syntax.IfClause)
		if e == nil {
			return
		}
		e.Position, e.ThenPos, e.FiPos = IncCol(e.Position, i, l), IncCol(e.ThenPos, i, l), IncCol(e.FiPos, i, l)
		IncColNode(e.Else, i, l)
		for _, ee := range e.Cond {
			IncColNode(ee, i, l)
		}
		for _, ee := range e.Then {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.CondLast {
			e.CondLast[i].Hash = IncCol(ee.Hash, i, l)
		}
		for i, ee := range e.ThenLast {
			e.ThenLast[i].Hash = IncCol(ee.Hash, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.WhileClause:
		e := n.(*syntax.WhileClause)
		if e == nil {
			return
		}
		e.WhilePos, e.DoPos, e.DonePos = IncCol(e.WhilePos, i, l), IncCol(e.DoPos, i, l), IncCol(e.DonePos, i, l)
		for _, ee := range e.Cond {
			IncColNode(ee, i, l)
		}
		for _, ee := range e.Do {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.CondLast {
			e.CondLast[i].Hash = IncCol(ee.Hash, i, l)
		}
		for i, ee := range e.DoLast {
			e.DoLast[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.ForClause:
		e := n.(*syntax.ForClause)
		if e == nil {
			return
		}
		e.ForPos, e.DoPos, e.DonePos = IncCol(e.ForPos, i, l), IncCol(e.DoPos, i, l), IncCol(e.DonePos, i, l)
		IncColNode(e.Loop, i, l)
		for _, ee := range e.Do {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.DoLast {
			e.DoLast[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.WordIter:
		e := n.(*syntax.WordIter)
		if e == nil {
			return
		}
		e.InPos = IncCol(e.InPos, i, l)
		IncColNode(e.Name, i, l)
		for _, ee := range e.Items {
			IncColNode(ee, i, l)
		}
	case *syntax.CStyleLoop:
		e := n.(*syntax.CStyleLoop)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncCol(e.Lparen, i, l), IncCol(e.Rparen, i, l)
		IncColNode(e.Init, i, l)
		IncColNode(e.Cond, i, l)
		IncColNode(e.Post, i, l)
	case *syntax.CaseClause:
		e := n.(*syntax.CaseClause)
		if e == nil {
			return
		}
		e.Case, e.In, e.Esac = IncCol(e.Case, i, l), IncCol(e.In, i, l), IncCol(e.Esac, i, l)
		IncColNode(e.Word, i, l)
		for _, ee := range e.Items {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.CaseItem:
		e := n.(*syntax.CaseItem)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		for _, ee := range e.Patterns {
			IncColNode(ee, i, l)
		}
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Comments {
			e.Comments[i].Hash = IncCol(ee.Hash, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.Block:
		e := n.(*syntax.Block)
		if e == nil {
			return
		}
		e.Lbrace, e.Rbrace = IncCol(e.Lbrace, i, l), IncCol(e.Rbrace, i, l)
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.Subshell:
		e := n.(*syntax.Subshell)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncCol(e.Lparen, i, l), IncCol(e.Rparen, i, l)
		for _, ee := range e.Stmts {
			IncColNode(ee, i, l)
		}
		for i, ee := range e.Last {
			e.Last[i].Hash = IncCol(ee.Hash, i, l)
		}
	case *syntax.BinaryCmd:
		e := n.(*syntax.BinaryCmd)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.X, i, l)
		IncColNode(e.Y, i, l)
	case *syntax.FuncDecl:
		e := n.(*syntax.FuncDecl)
		if e == nil {
			return
		}
		e.Position = IncCol(e.Position, i, l)
		IncColNode(e.Name, i, l)
		IncColNode(e.Body, i, l)
	case *syntax.ArithmCmd:
		e := n.(*syntax.ArithmCmd)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
		IncColNode(e.X, i, l)
	case *syntax.TestClause:
		e := n.(*syntax.TestClause)
		if e == nil {
			return
		}
		e.Left, e.Right = IncCol(e.Left, i, l), IncCol(e.Right, i, l)
		IncColNode(e.X, i, l)
	case *syntax.BinaryTest:
		e := n.(*syntax.BinaryTest)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.X, i, l)
		IncColNode(e.Y, i, l)
	case *syntax.UnaryTest:
		e := n.(*syntax.UnaryTest)
		if e == nil {
			return
		}
		e.OpPos = IncCol(e.OpPos, i, l)
		IncColNode(e.X, i, l)
	case *syntax.ParenTest:
		e := n.(*syntax.ParenTest)
		if e == nil {
			return
		}
		e.Lparen, e.Rparen = IncCol(e.Lparen, i, l), IncCol(e.Rparen, i, l)
		IncColNode(e.X, i, l)
	case *syntax.DeclClause:
		e := n.(*syntax.DeclClause)
		if e == nil {
			return
		}
		IncColNode(e.Variant, i, l)
		for _, ee := range e.Args {
			IncColNode(ee, i, l)
		}
	case *syntax.LetClause:
		e := n.(*syntax.LetClause)
		if e == nil {
			return
		}
		e.Let = IncCol(e.Let, i, l)
		for _, ee := range e.Exprs {
			IncColNode(ee, i, l)
		}
	case *syntax.TimeClause:
		e := n.(*syntax.TimeClause)
		if e == nil {
			return
		}
		e.Time = IncCol(e.Time, i, l)
		IncColNode(e.Stmt, i, l)
	case *syntax.CoprocClause:
		e := n.(*syntax.CoprocClause)
		if e == nil {
			return
		}
		e.Coproc = IncCol(e.Coproc, i, l)
		IncColNode(e.Name, i, l)
		IncColNode(e.Stmt, i, l)
	}
}
