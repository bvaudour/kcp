package position

import (
	"reflect"

	"mvdan.cc/sh/v3/syntax"
)

func isNil(e any) bool {
	if e == nil {
		return true
	}

	switch reflect.TypeOf(e).Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
		return reflect.ValueOf(e).IsNil()
	}
	return false
}

func IncLine(pos syntax.Pos, lines ...uint) syntax.Pos {
	line := uint(1)
	if len(lines) > 0 && lines[0] > 0 {
		line = lines[0]
	}

	return syntax.NewPos(pos.Offset()+line, pos.Line()+line, 1)
}

type PosDiff struct {
	From   syntax.Pos
	Ignore syntax.Node
	Line   int
	Col    int
	Offset int
}

func Diff(oldPos, newPos syntax.Pos) *PosDiff {
	return &PosDiff{
		From:   oldPos,
		Line:   int(newPos.Line()) - int(oldPos.Line()),
		Col:    int(newPos.Col()) - int(oldPos.Col()),
		Offset: int(newPos.Offset()) - int(oldPos.Offset()),
	}
}

func (diff *PosDiff) IsNil() bool {
	return diff == nil || (diff.Line == 0 && diff.Col == 0 && diff.Offset == 0)
}

func (diff PosDiff) Clone() PosDiff {
	cloned := diff
	return cloned
}

func (diff *PosDiff) AddTo(pos syntax.Pos) syntax.Pos {
	if !pos.IsValid() || diff.IsNil() {
		return pos
	}
	newOffset := uint(max(0, int(pos.Offset())+diff.Offset))
	newLine := uint(max(0, int(pos.Line())+diff.Line))
	newCol := uint(max(0, int(pos.Col())+diff.Col))

	return syntax.NewPos(newOffset, newLine, newCol)
}

// mv is a private helper method to apply a diff to a position `pos`
// only if it is after `fromPos`. It correctly handles the column adjustment:
// if `pos` is on a different line than `fromPos`, the column diff is reset to zero.
func (diff *PosDiff) Mv(pos syntax.Pos) syntax.Pos {
	if diff.IsNil() || !pos.IsValid() {
		return pos
	}
	if !diff.From.IsValid() {
		diff.From = pos
	}
	if Cmp(pos, diff.From) < 0 {
		return pos
	}

	d := diff.Clone()
	if pos.Line() != diff.From.Line() {
		d.Col = 0
	}

	return d.AddTo(pos)
}

func (diff *PosDiff) MvAll(positions ...*syntax.Pos) *PosDiff {
	for _, p := range positions {
		*p = diff.Mv(*p)
	}

	return diff
}

func (diff *PosDiff) upd(node syntax.Node) {
	if diff.IsNil() || isNil(node) || diff.Ignore == node {
		return
	}

	switch x := node.(type) {
	case *syntax.ArithmCmd:
		diff.
			MvAll(&x.Left, &x.Right).
			upd(x.X)
	case *syntax.ArithmExp:
		diff.
			MvAll(&x.Left, &x.Right).
			upd(x.X)
	case *syntax.ArrayElem:
		diff.
			updAll(x.Index, x.Value).
			updComments(x.Comments)
	case *syntax.ArrayExpr:
		diff.
			MvAll(&x.Lparen, &x.Rparen).
			updComments(x.Last)
		updAll(diff, x.Elems)
	case *syntax.Assign:
		diff.updAll(x.Name, x.Index, x.Value, x.Array)
	case *syntax.BinaryArithm:
		diff.
			MvAll(&x.OpPos).
			updAll(x.X, x.Y)
	case *syntax.BinaryCmd:
		diff.
			MvAll(&x.OpPos).
			updAll(x.X, x.Y)
	case *syntax.BinaryTest:
		diff.
			MvAll(&x.OpPos).
			updAll(x.X, x.Y)
	case *syntax.Block:
		diff.
			MvAll(&x.Lbrace, &x.Rbrace).
			updComments(x.Last)
		updAll(diff, x.Stmts)
	case *syntax.BraceExp:
		updAll(diff, x.Elems)
	case *syntax.CallExpr:
		updAll(diff, x.Assigns)
		updAll(diff, x.Args)
	case *syntax.CaseClause:
		diff.
			MvAll(&x.Case, &x.In, &x.Esac).
			updAll(x.Word).
			updComments(x.Last)
		updAll(diff, x.Items)
	case *syntax.CaseItem:
		diff.
			MvAll(&x.OpPos).
			updComments(x.Comments, x.Last)
		updAll(diff, x.Patterns)
		updAll(diff, x.Stmts)
	case *syntax.CmdSubst:
		diff.
			MvAll(&x.Left, &x.Right).
			updComments(x.Last)
		updAll(diff, x.Stmts)
	case *syntax.Comment:
		diff.MvAll(&x.Hash)
	case *syntax.CoprocClause:
		diff.
			MvAll(&x.Coproc).
			updAll(x.Name, x.Stmt)
	case *syntax.CStyleLoop:
		diff.
			MvAll(&x.Lparen, &x.Rparen).
			updAll(x.Init, x.Cond, x.Post)
	case *syntax.DblQuoted:
		diff.MvAll(&x.Left, &x.Right)
		updAll(diff, x.Parts)
	case *syntax.DeclClause:
		diff.updAll(x.Variant)
		updAll(diff, x.Args)
	case *syntax.ExtGlob:
		diff.
			MvAll(&x.OpPos).
			updAll(x.Pattern)
	case *syntax.File:
		updAll(diff, x.Stmts).
			updComments(x.Last)
	case *syntax.ForClause:
		diff.
			MvAll(&x.ForPos, &x.DoPos, &x.DonePos).
			updAll(x.Loop).
			updComments(x.DoLast)
		updAll(diff, x.Do)
	case *syntax.FuncDecl:
		diff.
			MvAll(&x.Position).
			updAll(x.Name, x.Body)
	case *syntax.IfClause:
		diff.
			MvAll(&x.Position, &x.ThenPos, &x.FiPos).
			updAll(x.Else).
			updComments(x.CondLast, x.ThenLast, x.Last)
		updAll(diff, x.Cond)
		updAll(diff, x.Then)
	case *syntax.LetClause:
		diff.MvAll(&x.Let)
		updAll(diff, x.Exprs)
	case *syntax.Lit:
		diff.MvAll(&x.ValuePos, &x.ValueEnd)
	case *syntax.ParamExp:
		subnodes := []syntax.Node{x.Param, x.Index}
		if !isNil(x.Slice) {
			subnodes = append(subnodes, x.Slice.Offset, x.Slice.Length)
		}
		if !isNil(x.Repl) {
			subnodes = append(subnodes, x.Repl.Orig, x.Repl.With)
		}
		if !isNil(x.Exp) {
			subnodes = append(subnodes, x.Exp.Word)
		}
		diff.
			MvAll(&x.Dollar, &x.Rbrace).
			updAll(subnodes...)
	case *syntax.ParenArithm:
		diff.
			MvAll(&x.Lparen, &x.Rparen).
			updAll(x.X)
	case *syntax.ParenTest:
		diff.
			MvAll(&x.Lparen, &x.Rparen).
			updAll(x.X)
	case *syntax.ProcSubst:
		diff.
			MvAll(&x.OpPos, &x.Rparen).
			updComments(x.Last)
		updAll(diff, x.Stmts)
	case *syntax.Redirect:
		diff.
			MvAll(&x.OpPos).
			updAll(x.N, x.Word, x.Hdoc)
	case *syntax.SglQuoted:
		diff.MvAll(&x.Left, &x.Right)
	case *syntax.Stmt:
		diff.
			MvAll(&x.Position, &x.Semicolon).
			updAll(x.Cmd).
			updComments(x.Comments)
		updAll(diff, x.Redirs)
	case *syntax.Subshell:
		diff.
			MvAll(&x.Lparen, &x.Rparen).
			updComments(x.Last)
		updAll(diff, x.Stmts)
	case *syntax.TestClause:
		diff.
			MvAll(&x.Left, &x.Right).
			updAll(x.X)
	case *syntax.TestDecl:
		diff.
			MvAll(&x.Position).
			updAll(x.Description, x.Body)
	case *syntax.TimeClause:
		diff.
			MvAll(&x.Time).
			updAll(x.Stmt)
	case *syntax.UnaryArithm:
		diff.
			MvAll(&x.OpPos).
			updAll(x.X)
	case *syntax.UnaryTest:
		diff.
			MvAll(&x.OpPos).
			updAll(x.X)
	case *syntax.WhileClause:
		diff.
			MvAll(&x.WhilePos, &x.DoPos, &x.DonePos).
			updComments(x.CondLast, x.DoLast)
		updAll(diff, x.Cond)
		updAll(diff, x.Do)
	case *syntax.Word:
		updAll(diff, x.Parts)
	case *syntax.WordIter:
		diff.
			MvAll(&x.InPos).
			updAll(x.Name)
		updAll(diff, x.Items)
	case syntax.ArithmExpr:
		switch y := x.(type) {
		case *syntax.BinaryArithm:
			diff.upd(y)
		case *syntax.ParenArithm:
			diff.upd(y)
		case *syntax.UnaryArithm:
			diff.upd(y)
		case *syntax.Word:
			diff.upd(y)
		}
	case syntax.Command:
		switch y := x.(type) {
		case *syntax.ArithmCmd:
			diff.upd(y)
		case *syntax.BinaryCmd:
			diff.upd(y)
		case *syntax.Block:
			diff.upd(y)
		case *syntax.CallExpr:
			diff.upd(y)
		case *syntax.CaseClause:
			diff.upd(y)
		case *syntax.CoprocClause:
			diff.upd(y)
		case *syntax.DeclClause:
			diff.upd(y)
		case *syntax.ForClause:
			diff.upd(y)
		case *syntax.FuncDecl:
			diff.upd(y)
		case *syntax.IfClause:
			diff.upd(y)
		case *syntax.LetClause:
			diff.upd(y)
		case *syntax.Subshell:
			diff.upd(y)
		case *syntax.TestClause:
			diff.upd(y)
		case *syntax.TimeClause:
			diff.upd(y)
		case *syntax.WhileClause:
			diff.upd(y)
		}
	case syntax.Loop:
		switch y := x.(type) {
		case *syntax.CStyleLoop:
			diff.upd(y)
		case *syntax.WordIter:
			diff.upd(y)
		}
	case syntax.TestExpr:
		switch y := x.(type) {
		case *syntax.BinaryTest:
			diff.upd(y)
		case *syntax.ParenTest:
			diff.upd(y)
		case *syntax.UnaryTest:
			diff.upd(y)
		case *syntax.Word:
			diff.upd(y)
		}
	case syntax.WordPart:
		switch y := x.(type) {
		case *syntax.ArithmExp:
			diff.upd(y)
		case *syntax.CmdSubst:
			diff.upd(y)
		case *syntax.DblQuoted:
			diff.upd(y)
		case *syntax.ExtGlob:
			diff.upd(y)
		case *syntax.Lit:
			diff.upd(y)
		case *syntax.ParamExp:
			diff.upd(y)
		case *syntax.ProcSubst:
			diff.upd(y)
		case *syntax.SglQuoted:
			diff.upd(y)
		}
	}
}

func (diff *PosDiff) updAll(nodes ...syntax.Node) *PosDiff {
	return updAll(diff, nodes)
}

func (diff *PosDiff) updComments(sliceComments ...[]syntax.Comment) *PosDiff {
	for _, l := range sliceComments {
		for i := range l {
			diff.upd(&l[i])
		}
	}

	return diff
}

func updAll[T syntax.Node](diff *PosDiff, nodes []T) *PosDiff {
	for _, node := range nodes {
		diff.upd(node)
	}
	return diff
}

func (diff *PosDiff) Update(nodes ...syntax.Node) {
	diff.updAll(nodes...)
}

func (diff *PosDiff) UpdateComments(comments ...[]syntax.Comment) {
	diff.updComments(comments...)
}

func Update[T syntax.Node](diff *PosDiff, nodes []T) {
	updAll(diff, nodes)
}
