package pkgbuild

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"codeberg.org/bvaudour/kcp/pkgbuild/env"
	"codeberg.org/bvaudour/kcp/pkgbuild/format"
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"codeberg.org/bvaudour/kcp/pkgbuild/standard"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

// PKGBUILD represents a parsed PKGBUILD file.
type PKGBUILD struct {
	info.NodeInfoList
	variable bool
	file     *syntax.File
	env      expand.WriteEnviron
}

func (p *PKGBUILD) parseVariables(r io.Reader) (err error) {
	var file *syntax.File
	if file, err = syntax.NewParser(syntax.KeepComments(false)).Parse(r, "PKGBUILD"); err != nil {
		return
	}

	environ := env.New()
	for _, stmt := range file.Stmts {
		cmd := stmt.Cmd
		if node, ok := cmd.(*syntax.CallExpr); ok && len(node.Args) == 0 {
			for _, assign := range node.Assigns {
				if err = env.ParseAssignation(environ, assign); err != nil {
					return
				}
			}
		}
	}

	p.variable, p.file, p.env = true, file, environ
	return
}

func (p *PKGBUILD) parseComplete(r io.Reader) (err error) {
	var file *syntax.File
	if file, err = syntax.NewParser(syntax.KeepComments(true)).Parse(r, "PKGBUILD"); err != nil {
		return
	}

	environ := env.New()
	var infos info.NodeInfoList
	for i, stmt := range file.Stmts {
		var node *info.NodeInfo
		if node, err = info.New(i, stmt, environ); err != nil {
			return
		}
		infos = append(infos, node)
	}

	p.file, p.NodeInfoList, p.env = file, infos, environ
	return
}

func (p *PKGBUILD) Env() expand.WriteEnviron {
	return p.env
}

// Decode parses the given reader and returns a PKGBUILD
// or an error if parse failed.
func Decode(r io.Reader) (p *PKGBUILD, err error) {
	p = new(PKGBUILD)
	err = p.parseComplete(r)

	return
}

// DecodeVars is same as Decode but it decodes only
// variables declarations.
func DecodeVars(r io.Reader) (p *PKGBUILD, err error) {
	p = new(PKGBUILD)
	err = p.parseVariables(r)

	return
}

// ReadVersion reads the given reader
// and returns the full version of the PKGBUILD.
func ReadVersion(r io.Reader) (s string) {
	if p, err := DecodeVars(r); err == nil {
		s = p.GetFullVersion()
	}

	return
}

// Debug returns a string represention of the PKGBUILD
// (for debugging only).
func (p *PKGBUILD) Debug(w io.Writer) {
	if p.file != nil {
		p.file.Stmts = p.Stmts()
	}
	syntax.DebugPrint(w, p.file)
}

func (p *PKGBUILD) Format(formater format.Formater) {
	var comments []syntax.Comment
	if p.file != nil {
		comments = p.file.Last
	}

	p.NodeInfoList, comments = formater.Format(p.NodeInfoList, comments)
	if p.file != nil {
		p.file.Stmts, p.file.Last = p.Stmts(), comments
	}
}

// Encode writes the PKGBUILD to the given writer.
// Its returns an error if write failed.
func (p *PKGBUILD) Encode(w io.Writer, formater ...format.Formater) error {
	if len(formater) > 0 {
		p.Format(formater[0])
	} else if p.file != nil {
		p.file.Stmts = p.Stmts()
	}
	if p.file == nil {
		return nil
	}
	var begin syntax.Pos
	if len(p.NodeInfoList) > 0 {
		begin, _ = p.NodeInfoList[0].Position()
	} else if len(p.file.Last) > 0 {
		begin = p.file.Last[0].Pos()
	}
	if begin.IsValid() && begin.Line() > 1 {
		bytes := make([]byte, uint(begin.Line())-1)
		for i := range bytes {
			bytes[i] = '\n'
		}
		if _, err := w.Write(bytes); err != nil {
			return err
		}
	}

	return syntax.NewPrinter().Print(w, p.file)
}

// String returns the string representation of the PKGBUILD
// as we could see on the file.
func (p *PKGBUILD) String() string {
	var s strings.Builder
	p.Encode(&s)

	return s.String()
}

// GetValues returns all values indexed by the name of the variable.
func (p *PKGBUILD) GetValues() (out map[string]string) {
	out = make(map[string]string)

	if !p.variable {
		return
	}

	p.env.Each(func(n string, v expand.Variable) bool {
		out[n] = env.ToString(p.env, v).Str
		return true
	})

	return out
}

// GetArrayValues is same as GetValues but it returns all values by variable.
func (p *PKGBUILD) GetArrayValues() (out map[string][]string) {
	out = make(map[string][]string)

	if !p.variable {
		return
	}

	p.env.Each(func(n string, v expand.Variable) bool {
		out[n] = env.ToIndexed(p.env, v).List
		return true
	})

	return
}

// GetValue returns the value of the variable “name”
// or an empty string if the variable doesn’t exist.
func (p *PKGBUILD) GetValue(name string) string {
	if p.variable {
		return env.GetDeep(p.env, name).Str
	}
	return ""
}

// GetArrayValue returns the real array value of a variable or
// an empty array if the variable doesn’t exist.
func (p *PKGBUILD) GetArrayValue(name string) []string {
	if p.variable {
		return env.GetDeep(p.env, name).List
	}
	return nil
}

// HasValue returns true if it is the name of a variable
// and if it has a value.
func (p *PKGBUILD) HasValue(name string) bool {
	if p.variable {
		v := env.GetDeep(p.env, name)
		switch v.Kind {
		case expand.Associative:
			return len(v.Map) > 0
		case expand.Indexed:
			return len(v.List) > 0
		default:
			return v.IsSet() && v.Str != ""
		}
	}

	return false
}

// GetFullVersion returns the full version of the PKGBUILD,
// including the pkgrel and the eventual epoch.
func (p *PKGBUILD) GetFullVersion() string {
	s := fmt.Sprintf("%s-%s", p.GetValue(standard.PKGVER), p.GetValue(standard.PKGREL))
	if p.HasValue(standard.EPOCH) {
		s = fmt.Sprintf("%s:%s", p.GetValue(standard.EPOCH), s)
	}

	return s
}

func (p *PKGBUILD) nextId() int {
	next := 1
	for _, node := range p.NodeInfoList {
		next = max(next, node.Id+1)
	}
	return next
}

func (p *PKGBUILD) Add(nodes ...*info.NodeInfo) {
	currentPos := syntax.NewPos(0, 1, 1)
	if l := len(p.NodeInfoList); l > 0 {
		_, end := p.NodeInfoList[l-1].Position()
		currentPos = position.IncLine(end)
	}
	initialPos := currentPos

	id := p.nextId()
	for _, node := range nodes {
		node.Id = id
		id++
		begin, _ := node.Position()
		diff := position.Diff(begin, currentPos)
		diff.Update(node.Stmt)
		_, end := node.Position()
		currentPos = position.IncLine(end)
	}
	p.NodeInfoList = append(p.NodeInfoList, nodes...)
	diff := position.Diff(initialPos, currentPos)
	diff.UpdateComments(p.file.Last)
}

func (p *PKGBUILD) Remove(ids ...int) {
	var nodes info.NodeInfoList
	var diff *position.PosDiff
	for _, node := range p.NodeInfoList {
		if diff != nil {
			diff.Update(node.Stmt)
		}
		if !slices.Contains(ids, node.Id) {
			nodes = append(nodes, node)
			continue
		}
		begin, end := node.Position()
		diff = position.Diff(position.IncLine(end), begin)
	}
	diff.UpdateComments(p.file.Last)
}

func (p *PKGBUILD) FindLast(name string, types ...info.NodeType) (node *info.NodeInfo, index int) {
	index = -1
	for i, n := range p.NodeInfoList {
		if n.Name == name {
			if len(types) > 0 && slices.Contains(types, n.Type) {
				index, node = i, n
			}
		}
	}

	return
}

func (p *PKGBUILD) FindById(id int) (node *info.NodeInfo, index int) {
	index = -1
	for i, n := range p.NodeInfoList {
		if n.Id == id {
			return n, i
		}
	}
	return
}

func (p *PKGBUILD) UpdateValue(id int, value string) {
	node, i := p.FindById(id)
	if i < 0 || node.Type == info.Function {
		return
	}

	value = strings.TrimSpace(value)
	_, e := node.Position()
	ib, ie := node.InnerPosition()
	var v string
	if node.Type == info.ArrayVar {
		v = fmt.Sprintf("%s=(%s)", node.Name, value)
	} else {
		v = fmt.Sprintf("%s=%s", node.Name, value)
	}
	parsed, err := syntax.NewParser().Parse(strings.NewReader(v), "")
	if err != nil || len(parsed.Stmts) == 0 {
		return
	}

	expr, ok := parsed.Stmts[0].Cmd.(*syntax.CallExpr)
	if !(ok && len(expr.Assigns) == 1 && len(expr.Args) == 0) {
		return
	}
	assign := expr.Assigns[0]
	diff := position.Diff(assign.Pos(), ib)
	diff.Update(assign)
	diff = position.Diff(ie, assign.End())
	diff.UpdateComments(node.Stmt.Comments)
	node.Stmt.Cmd = expr

	_, ne := node.Position()
	diff = position.Diff(e, ne)
	for _, n := range p.NodeInfoList[i:] {
		diff.Update(n.Stmt)
	}
	diff.UpdateComments(p.file.Last)
}
