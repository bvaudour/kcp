package info

import (
	"fmt"
	"slices"
	"strings"

	"codeberg.org/bvaudour/kcp/pkgbuild/env"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

// NodeType represents a type of PKGBUILD node.
type NodeType uint

const (
	Unknown NodeType = iota
	SingleVar
	ArrayVar
	Function
)

func TypeOf(stmt *syntax.Stmt) NodeType {
	switch stmt.Cmd.(type) {
	case *syntax.CallExpr:
		n := stmt.Cmd.(*syntax.CallExpr)
		if len(n.Args) > 0 || len(n.Assigns) > 1 || len(n.Assigns) == 0 {
			return Unknown
		}
		assign := n.Assigns[0]
		if assign.Index != nil || assign.Append {
			return Unknown
		}
		if assign.Array != nil {
			return ArrayVar
		} else {
			return SingleVar
		}
	case *syntax.FuncDecl:
		return Function
	}

	return Unknown
}

// NodeInfo represents the informations of a node.
type NodeInfo struct {
	Id     int
	Type   NodeType
	Name   string
	Value  string
	Values []string
	Stmt   *syntax.Stmt
}

func newInfo(idx int, stmt *syntax.Stmt, t NodeType, environ expand.WriteEnviron) (node *NodeInfo, err error) {
	node = new(NodeInfo)
	pr := syntax.NewPrinter()
	var sb strings.Builder

	switch t {
	case Function:
		n := stmt.Cmd.(*syntax.FuncDecl)
		if err = pr.Print(&sb, n.Body); err != nil {
			return
		}
		node.Name, node.Value = n.Name.Value, sb.String()
	case SingleVar:
		n := stmt.Cmd.(*syntax.CallExpr).Assigns[0]
		if err = env.ParseAssignation(environ, n); err != nil {
			return
		}
		node.Name = n.Name.Value
		if n.Value != nil {
			node.Value, _ = env.Literal(environ, n.Value)
		}
	case ArrayVar:
		n := stmt.Cmd.(*syntax.CallExpr).Assigns[0]
		var values []string
		for _, elem := range n.Array.Elems {
			var v string
			v, _ = env.Literal(environ, elem.Value)
			values = append(values, v)
		}
		node.Name, node.Values = n.Name.Value, values
	default:
		err = fmt.Errorf(errInvalidSyntax, position.RangeString(stmt.Pos(), stmt.End()))
		return
	}
	node.Id, node.Type, node.Stmt = idx, t, stmt

	return
}

// New returns a node info from a statement.
func New(idx int, stmt *syntax.Stmt, environ expand.WriteEnviron) (*NodeInfo, error) {
	return newInfo(idx, stmt, TypeOf(stmt), environ)
}

// InnerPosition returns the position of the node, excluding comments.
func (info *NodeInfo) InnerPosition() (begin, end syntax.Pos) {
	return info.Stmt.Pos(), info.Stmt.End()
}

// Position returns the position of the node, including comments.
func (info *NodeInfo) Position() (begin, end syntax.Pos) {
	begin, end = info.InnerPosition()

	for _, c := range info.Stmt.Comments {
		bc, ec := c.Pos(), c.End()
		if position.Cmp(bc, begin) < 0 {
			begin = bc
		}
		if position.Cmp(ec, end) > 0 {
			end = ec
		}
	}

	return
}

func comments(n syntax.Node) (out []syntax.Comment) {
	switch e := n.(type) {
	case *syntax.ArithmCmd:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.ArithmExp:
		if e == nil || e.X == nil {
			return
		}
		return comments(e.X)
	case *syntax.ArrayElem:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Index),
			comments(e.Value),
			e.Comments,
		)
	case *syntax.ArrayExpr:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Elems),
			e.Last,
		)
	case *syntax.Assign:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Index),
			comments(e.Value),
			comments(e.Array),
		)
	case *syntax.BinaryArithm:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.X),
			comments(e.Y),
		)
	case *syntax.BinaryCmd:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.X),
			comments(e.Y),
		)
	case *syntax.BinaryTest:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.X),
			comments(e.Y),
		)
	case *syntax.Block:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.BraceExp:
		if e == nil {
			return
		}
		return scomments(e.Elems)
	case *syntax.CStyleLoop:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Init),
			comments(e.Cond),
			comments(e.Post),
		)
	case *syntax.CallExpr:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Assigns),
			scomments(e.Args),
		)
	case *syntax.CaseClause:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Word),
			scomments(e.Items),
			e.Last,
		)
	case *syntax.CaseItem:
		if e == nil {
			return
		}
		return slices.Concat(
			e.Comments,
			scomments(e.Patterns),
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.CmdSubst:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.CoprocClause:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Name),
			comments(e.Stmt),
		)
	case *syntax.DblQuoted:
		if e == nil {
			return
		}
		return scomments(e.Parts)
	case *syntax.DeclClause:
		if e == nil {
			return
		}
		return scomments(e.Args)
	case *syntax.File:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.ForClause:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Loop),
			scomments(e.Do),
			e.DoLast,
		)
	case *syntax.FuncDecl:
		if e == nil {
			return
		}
		return comments(e.Body)
	case *syntax.IfClause:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Cond),
			e.CondLast,
			scomments(e.Then),
			e.ThenLast,
			comments(e.Else),
			e.Last,
		)
	case *syntax.LetClause:
		if e == nil {
			return
		}
		return scomments(e.Exprs)
	case *syntax.ParamExp:
		if e == nil {
			return
		}
		out = comments(e.Index)
		if e.Slice != nil {
			out = slices.Concat(
				out,
				comments(e.Slice.Offset),
				comments(e.Slice.Length),
			)
		}
		if e.Repl != nil {
			out = slices.Concat(
				out,
				comments(e.Repl.Orig),
				comments(e.Repl.With),
			)
		}
		if e.Exp != nil {
			out = slices.Concat(
				out,
				comments(e.Exp.Word),
			)
		}
		return
	case *syntax.ParenArithm:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.ParenTest:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.ProcSubst:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.Redirect:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Word),
			comments(e.Hdoc),
		)
	case *syntax.Stmt:
		if e == nil {
			return
		}
		return slices.Concat(
			e.Comments,
			comments(e.Cmd),
			scomments(e.Redirs),
		)
	case *syntax.Subshell:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Stmts),
			e.Last,
		)
	case *syntax.TestClause:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.TestDecl:
		if e == nil {
			return
		}
		return slices.Concat(
			comments(e.Description),
			comments(e.Body),
		)
	case *syntax.TimeClause:
		if e == nil {
			return
		}
		return comments(e.Stmt)
	case *syntax.UnaryArithm:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.UnaryTest:
		if e == nil {
			return
		}
		return comments(e.X)
	case *syntax.WhileClause:
		if e == nil {
			return
		}
		return slices.Concat(
			scomments(e.Cond),
			e.CondLast,
			scomments(e.Do),
			e.DoLast,
		)
	case *syntax.Word:
		if e == nil {
			return
		}
		return scomments(e.Parts)
	case *syntax.WordIter:
		if e == nil {
			return
		}
		return scomments(e.Items)
	default:
		return
	}
}

func scomments[T syntax.Node](nodes []T) (out []syntax.Comment) {
	for _, n := range nodes {
		out = slices.Concat(out, comments(n))
	}
	return
}

func (info *NodeInfo) HeaderComments() (out []syntax.Comment) {
	allComments := comments(info.Stmt)
	begin, _ := info.InnerPosition()
	for _, c := range allComments {
		if position.Cmp(c.End(), begin) <= 0 {
			out = append(out, c)
		}
	}

	return
}

func (info *NodeInfo) InlineComments() (out []syntax.Comment) {
	allComments := comments(info.Stmt)
	begin, end := info.InnerPosition()
	for _, c := range allComments {
		if position.Cmp(c.Pos(), begin) >= 0 && position.Cmp(c.End(), end) <= 0 {
			out = append(out, c)
		}
	}

	return
}

func (info *NodeInfo) FooterComments() (out []syntax.Comment) {
	allComments := comments(info.Stmt)
	_, end := info.InnerPosition()
	for _, c := range allComments {
		if position.Cmp(c.Pos(), end) <= 0 {
			out = append(out, c)
		}
	}

	return
}

func (info *NodeInfo) Clone(removeComments ...bool) *NodeInfo {
	cloned := NodeInfo{
		Id:     info.Id,
		Type:   info.Type,
		Name:   info.Name,
		Value:  info.Value,
		Values: make([]string, len(info.Values)),
		Stmt:   info.Stmt,
	}

	copy(cloned.Values, info.Values)

	printer, parser := syntax.NewPrinter(), syntax.NewParser(syntax.KeepComments(len(removeComments) == 0 || !removeComments[0]))
	begin, _ := info.Position()
	var sb strings.Builder
	sb.WriteString(strings.Repeat("\n", int(begin.Line())-1))

	if err := printer.Print(&sb, info.Stmt); err == nil {
		if f, err := parser.Parse(strings.NewReader(sb.String()), ""); err == nil && len(f.Stmts) > 0 {
			cloned.Stmt = f.Stmts[0]
		}
	}

	return &cloned
}

// NodeInfoList represents a list of nodes’ infos.
type NodeInfoList []*NodeInfo

// GetVariables returns the list of all variable names.
func (infos NodeInfoList) GetVariables() (out []string) {
	done := make(map[string]bool)

	for _, info := range infos {
		if info.Type == Function {
			continue
		}
		name := info.Name
		if !done[name] {
			out = append(out, name)
			done[name] = true
		}
	}

	return
}

// GetFunctions returns the list of all function names.
func (infos NodeInfoList) GetFunctions() (out []string) {
	done := make(map[string]bool)

	for _, info := range infos {
		if info.Type != Function {
			continue
		}
		name := info.Name
		if !done[name] {
			out = append(out, name)
			done[name] = true
		}
	}

	return
}

// HasVariable returns true if variable “name” is declared.
func (infos NodeInfoList) HasVariable(name string) bool {
	for _, info := range infos {
		if info.Type != Function && info.Name == name {
			return true
		}
	}

	return false
}

// HasFunction returns true if function “name” is declared.
func (infos NodeInfoList) HasFunction(name string) bool {
	for _, info := range infos {
		if info.Type == Function && info.Name == name {
			return true
		}
	}

	return false
}

// GetInfos returns all nodes with the name “name”.
func (infos NodeInfoList) GetInfos(name string) (out []NodeInfo) {
	for _, info := range infos {
		if info.Name == name {
			out = append(out, *info)
		}
	}
	return
}

func (infos NodeInfoList) Stmts() (stmts []*syntax.Stmt) {
	stmts = make([]*syntax.Stmt, len(infos))
	for i, node := range infos {
		stmts[i] = node.Stmt
	}

	return
}

func (infos NodeInfoList) GetDuplicates() (duplicates map[string]NodeInfoList) {
	names := make(map[string]NodeInfoList)
	for _, info := range infos {
		names[info.Name] = append(names[info.Name], info)
	}

	duplicates = make(map[string]NodeInfoList)
	for name, nodes := range names {
		if len(nodes) > 1 {
			duplicates[name] = nodes
		}
	}

	return
}
