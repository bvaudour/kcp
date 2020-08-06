package pkgbuild

import (
	"fmt"
	"io"
	"strings"

	"github.com/bvaudour/kcp/pkgbuild/atom"
	"github.com/bvaudour/kcp/pkgbuild/format"
	"github.com/bvaudour/kcp/pkgbuild/scanner"
	"github.com/bvaudour/kcp/pkgbuild/standard"
	"github.com/bvaudour/kcp/position"
	"github.com/bvaudour/kcp/runes"
)

//PKGBUILD is the result of a decoded PKGBUILD file.
type PKGBUILD struct {
	atoms atom.Slice
	info  *atom.InfoList
}

//RecomputePosition recomputes the position
//of all atoms of the PKGBUILD starting at L.1,C.0.
func (p *PKGBUILD) RecomputePositions() {
	var pos position.Position
	pos.Line = 1
	for _, a := range p.atoms {
		_, pos = atom.RecomputeAllPositions(a, pos)
		pos = pos.Next('\n')
	}
}

//RecomputeValue recomputes the real value
//of all variables.
func (p *PKGBUILD) RecomputeValues() {
	p.info.RecomputeValues()
}

//RecomputeInfo recomputes the informations of
//the named variables.
//If recomputeValues is provided, the values of the
//variables are recomputed.
func (p *PKGBUILD) RecomputeInfos(recomputeValues ...bool) {
	p.info.UpdateAll(p.atoms)
	if len(recomputeValues) > 0 && recomputeValues[0] {
		p.RecomputeValues()
	}
}

//GetInfos returns the informations of all named atoms
//whose the type matches one of the given types.
//If no type is provided, it’s returns all informations.
func (p *PKGBUILD) GetInfos(types ...atom.AtomType) []*atom.Info {
	if len(types) == 0 {
		types = []atom.AtomType{atom.VarArray, atom.VarString, atom.Function}
	}
	return p.info.Filter(atom.NewNameMatcher(types...))
}

//GetInfo returns the first found information with the given name
//and optionnal type, or false if no info found.
func (p *PKGBUILD) GetInfo(name string, types ...atom.AtomType) (*atom.Info, bool) {
	cb := atom.CheckName(name)
	if len(types) > 0 {
		cb = atom.NamedCheckAll(atom.CheckName(name), atom.NewNameMatcher(types...))
	}
	return p.info.FilterFirst(cb)
}

//GetVariables returns the list of all variable names.
func (p *PKGBUILD) GetVariables() (out []string) {
	done := make(map[string]bool)
	infos := p.info.Variables()
	for _, e := range infos {
		name := e.Name()
		if !done[name] {
			done[name] = true
			out = append(out, name)
		}
	}
	return
}

//GetFunctions returns the list of all function names.
func (p *PKGBUILD) GetFunctions() (out []string) {
	done := make(map[string]bool)
	infos := p.info.Variables()
	for _, e := range infos {
		name := e.Name()
		if !done[name] {
			done[name] = true
			out = append(out, name)
		}
	}
	return
}

//GetValues returns all values indexed by the name of the variable.
func (p *PKGBUILD) GetValues() map[string]string {
	return p.info.GetValues()
}

//GetArrayValues is same as GetValues but it returns all values by variable.
func (p *PKGBUILD) GetArrayValues() map[string][]string {
	out := make(map[string][]string)
	for _, v := range p.info.Variables() {
		out[v.Name()] = v.ArrayValue()
	}
	return out
}

//GetValue returns the real value of a variable or
//an empty string if the variable doesn’t exist.
func (p *PKGBUILD) GetValue(name string) string {
	return p.info.GetValue(name)
}

//GetArrayValue returns the real array value of a variable or
//an empty array if the variable doesn’t exist.
func (p *PKGBUILD) GetArrayValue(name string) []string {
	var out []string
	for _, v := range p.info.Variables() {
		if v.Name() == name {
			out = v.ArrayValue()
		}
	}
	return out
}

//HasValue returns true if it is the name of a variable
//and if it has a value.
func (p *PKGBUILD) HasValue(name string) bool {
	return p.info.HasValue(name)
}

//ContainsVariable checks if the given name is a variable name.
func (p *PKGBUILD) ContainsVariable(name string) bool {
	cb := atom.NamedCheckAll(
		atom.NewNameMatcher(atom.VarArray, atom.VarString),
		atom.CheckName(name),
	)
	_, ok := p.info.FilterFirst(cb)
	return ok
}

//ContainsFunction checks if the given name is a function name.
func (p *PKGBUILD) ContainsFunction(name string) bool {
	cb := atom.NamedCheckAll(
		atom.NewNameMatcher(atom.Function),
		atom.CheckName(name),
	)
	_, ok := p.info.FilterFirst(cb)
	return ok
}

//Contains returns true if the given name is a variable or
//a function.
func (p *PKGBUILD) Contains(name string) bool {
	cb := atom.NamedCheckAll(
		atom.NewNameMatcher(atom.VarArray, atom.VarString, atom.Function),
		atom.CheckName(name),
	)
	_, ok := p.info.FilterFirst(cb)
	return ok
}

//GetFullVersion returns the full version of the PKGBUILD,
//including the pkgrel and the eventual epoch.
func (p *PKGBUILD) GetFullVersion() string {
	s := fmt.Sprintf("%s-%s", p.GetValue(standard.PKGVER), p.GetValue(standard.PKGREL))
	if p.HasValue(standard.EPOCH) {
		s = fmt.Sprintf("%s:%s", s, p.GetValue(standard.EPOCH))
	}
	return s
}

//Encode writes the PKGBUILD to the given writer.
//Its returns the number of wroten bytes and an error
//if write failed.
func (p *PKGBUILD) Encode(w io.Writer) (size int, err error) {
	if len(p.atoms) == 0 {
		return
	}
	var pos position.Position
	pos.Line++
	var s int
	for _, a := range p.atoms {
		begin, end := a.GetPositions()
		prefix := pos.Blank(begin)
		pos = end
		if s, err = fmt.Fprint(w, prefix); err != nil {
			return
		}
		size += s
		if s, err = fmt.Fprint(w, a.GetRaw()); err != nil {
			return
		}
		size += s
	}
	if s, err = fmt.Fprintln(w); err == nil {
		size += s
	}
	return
}

//Decode parses the reader in the PKGBUILD
//and returns an error if parse failed.
//Note that all actual atoms are erased before parsing.
//To keep the atoms, you should use the Scan method
//with flag keepInfo.
func (p *PKGBUILD) Decode(r io.Reader) error {
	return p.Scan(scanner.New(r))
}

//Debug returns a string represention of the PKGBUILD
//(for debugging only).
func (p *PKGBUILD) Debug() string {
	b := new(strings.Builder)
	for i, a := range p.atoms {
		fmt.Fprintf(b, "(%d) %s\n", i, atom.Debug(a))
		fmt.Fprintln(b, "++++++++++++++++++++++++++")
	}
	return b.String()
}

//String returns the string representation of the PKGBUILD
//as we could see on the file.
func (p *PKGBUILD) String() string {
	var s strings.Builder
	p.Encode(&s)
	return s.String()
}

//New returns an empty PKGBUILD.
func New() *PKGBUILD {
	return &PKGBUILD{
		info: atom.NewInfoList(),
	}
}

//Clone makes a deep copy of the PKGBUILD.
func (p *PKGBUILD) Clone() *PKGBUILD {
	c := New()
	c.atoms = p.atoms.Clone()
	c.RecomputeInfos(true)
	return c
}

//Reset removes all infos and atoms of the PKGBUILD.
func (p *PKGBUILD) Reset() {
	p.atoms = p.atoms[:0]
	p.RecomputeInfos(true)
}

//Scan uses the scanner to add atoms to the PKGBUILD.
//If no keepInfo or keepInfo is false, all actual data
//are resetted.
func (p *PKGBUILD) Scan(sc scanner.Scanner, keepInfo ...bool) error {
	if len(keepInfo) == 0 || !keepInfo[0] {
		p.Reset()
	}
	atoms, err := scanner.ScanAll(sc)
	if err == nil {
		p.atoms.Push(atoms...)
		p.RecomputeInfos(true)
	}
	return err
}

//Decode parses the given reader and returns a PKGBUILD
//or an error if parse failed.
func Decode(r io.Reader) (p *PKGBUILD, err error) {
	p = New()
	if err = p.Decode(r); err != nil {
		p = nil
	}
	return
}

//DecodeFast is same as Decode but it ignores the
//comments and the blank lines.
func DecodeFast(r io.Reader) (p *PKGBUILD, err error) {
	p = New()
	sc := scanner.NewFastScanner(r)
	if err = p.Scan(sc, true); err != nil {
		p = nil
	}
	return
}

//DecodeVars is same as Decode but it decodes only
//variables declarations.
func DecodeVars(r io.Reader) (p *PKGBUILD, err error) {
	p = New()
	sc := scanner.NewVarScanner(r)
	if err = p.Scan(sc, true); err != nil {
		p = nil
	}
	return
}

//ReadVersion reads the given reader
//and returns the full version of the PKGBUILD.
func ReadVersion(r io.Reader) (s string) {
	p := New()
	if err := p.Scan(scanner.NewVarScanner(r), true); err == nil {
		s = p.GetFullVersion()
	}
	return
}

//Format applies transformations to the PKGBUILD.
//If no transformations are given it uses the following transformations:
//- Remove all blank lines except the first
//- Remove all comment lines and trailing comments,
//- Remove all useless spaces,
//- Format the values of the variables (including remove inner comments),
//- Reorder functions and variables,
//- Add blank line before functions
func (p *PKGBUILD) Format(formatters ...format.Formatter) {
	if len(formatters) == 0 {
		formatters = []format.Formatter{
			format.RemoveBlankLinesExceptFirst,
			format.RemoveCommentLines,
			format.RemoveTrailingComments,
			format.RemoveExtraSpaces,
			format.BeautifulValues,
			format.ReorderFuncsAndVars,
			format.AddBlankLineBeforeFunctions,
		}
	}
	p.atoms = format.Format(formatters...)(p.atoms)
	p.RecomputePositions()
	p.RecomputeInfos(true)
}

func newLine(pos ...position.Position) atom.Atom {
	a := atom.NewBlank()
	if len(pos) > 0 {
		a.SetPositions(pos[0], pos[0])
	}
	return a
}

func newComment(comment string, pos ...position.Position) atom.Atom {
	a := atom.NewComment()
	raw := "#" + comment
	a.SetRaw(raw)
	if len(pos) > 0 {
		begin := pos[0]
		end := begin.NextString(raw)
		a.SetPositions(begin, end)
	}
	return a
}

//ContainsInfo returns true if the info is valid and comes
//from the PKGBUILD.
func (p *PKGBUILD) ContainsInfo(info *atom.Info) bool {
	i := info.Index()
	if i < 0 || i >= len(p.atoms) {
		return false
	}
	a := p.atoms[i]
	n := info.AtomNamed
	if a == n {
		return true
	} else if a.GetType() != atom.Group {
		return false
	}
	for _, c := range a.(*atom.AtomGroup).Childs {
		if c == n {
			return true
		}
	}
	return false
}

//AppendBlankLine adds a blank line at the end of the PKGBUILD.
func (p *PKGBUILD) AppendBlankLine() (ok bool) {
	pos := position.Position{Line: 1}
	if last := len(p.atoms) - 1; last >= 0 {
		pos = atom.GetEnd(p.atoms[last]).Next('\n')
	}
	p.atoms.Push(newLine(pos))
	return true
}

//PrependBlankLine adds a blank line at the beginning of the PKGBUILD.
func (p *PKGBUILD) PrependBlankLine(recomputePositions ...bool) (ok bool) {
	p.atoms.PushFront(newLine())
	if len(recomputePositions) > 0 && recomputePositions[0] {
		p.RecomputePositions()
	}
	p.RecomputeInfos()
	return true
}

//InsertBlankLineBefore adds a blank line before the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) InsertBlankLineBefore(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms.Insert(info.Index(), newLine())
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//InsertBlankLineAfter adds a blank line after the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) InsertBlankLineAfter(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms.Insert(info.Index()+1, newLine())
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//AppendComment adds a comment at the end of the PKGBUILD.
func (p *PKGBUILD) AppendComment(comment string) (ok bool) {
	begin := position.Position{Line: 1}
	if last := len(p.atoms) - 1; last >= 0 {
		begin = atom.GetEnd(p.atoms[last]).Next('\n')
	}
	p.atoms.Push(newComment(comment, begin))
	return true
}

//PrependComment adds a comment at the beginning of the PKGBUILD.
func (p *PKGBUILD) PrependComment(comment string, recomputePositions ...bool) (ok bool) {
	p.atoms.PushFront(newComment(comment))
	p.RecomputeInfos()
	if len(recomputePositions) > 0 && recomputePositions[0] {
		p.RecomputePositions()
	}
	return true
}

//InsertCommentBefore adds a comment line before the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) InsertCommentBefore(comment string, info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms.Insert(info.Index(), newComment(comment))
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//InsertCommentAfter adds a comment line after the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) InsertCommentAfter(comment string, info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms.Insert(info.Index()+1, newComment(comment))
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//SetTrailingComment adds or update the trailing comment of the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) SetTrailingComment(comment string, info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		n := info.AtomNamed
		c := newComment(comment, atom.GetEnd(n).Next(' '))
		a := atom.NewGroup()
		a.Childs = atom.Slice{n, c}
		a.SetPositions(atom.GetBegin(n), atom.GetEnd(c))
		p.atoms[info.Index()] = a
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//SetTrailingComment removes the trailing comment of the atom with the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) RemoveTrailingComment(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms[info.Index()] = info.AtomNamed
		p.RecomputeInfos()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//RemoveInfo remove the atom concerned by the given info.
//It returns false if the info is not valid.
func (p *PKGBUILD) RemoveInfo(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = p.ContainsInfo(info); ok {
		p.atoms.Remove(info.Index())
		p.RecomputeInfos(info.IsVar())
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

func (p *PKGBUILD) recomputeInfoPos(info *atom.Info, oldEnd, newEnd position.Position) {
	atom.SetEnd(info, newEnd)
	a := p.atoms[info.Index()]
	if a == info.AtomNamed {
		return
	}
	g := a.(*atom.AtomGroup)
	nf := false
	p0, p1 := oldEnd, newEnd
	for _, c := range g.Childs {
		if !nf {
			nf = c == info.AtomNamed
			continue
		}
		inc := p1.Diff(p0)
		b, e := c.GetPositions()
		if inc.Column != 0 && p0.Line != b.Line {
			inc.Column = 0
		}
		p0 = e
		_, p1 = atom.IncrementPosition(c, inc)
	}
	atom.SetEnd(g, p1)
}

//SetArrayVar forces the variable with the given info to be an array variable.
//It returns false if the info is not valid or if it’s not a variable.
func (p *PKGBUILD) SetArrayVar(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = info.IsStringVar() && p.ContainsInfo(info); ok {
		info.SetType(atom.VarArray)
		a := info.AtomNamed.(*atom.AtomVar)
		v, ok := a.GetValue()
		if !ok {
			v = atom.NewValue()
		}
		a.SetValues(v)
		_, e := a.GetNamePositions()
		_, e = atom.MoveAtom(v, e.NextString("=("))
		oe, ne := atom.GetEnd(a), e.Next(')')
		p.recomputeInfoPos(info, oe, ne)
		a.RecomputRaw()
		p.RecomputeValues()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

//SetStringVar forces the variable with the given info to be an string variable.
//It returns false if the info is not valid or if it’s not a variable.
func (p *PKGBUILD) SetStringVar(info *atom.Info, recomputePositions ...bool) (ok bool) {
	if ok = info.IsArrayVar() && p.ContainsInfo(info); ok {
		info.SetType(atom.VarString)
		a := info.AtomNamed.(*atom.AtomVar)
		v, ok := a.GetValue()
		if !ok {
			v = atom.NewValue()
		}
		a.SetValues(v)
		_, e := a.GetNamePositions()
		_, e = atom.MoveAtom(v, e.NextString("="))
		oe, ne := atom.GetEnd(a), e
		p.recomputeInfoPos(info, oe, ne)
		a.RecomputRaw()
		p.RecomputeValues()
		if len(recomputePositions) > 0 && recomputePositions[0] {
			p.RecomputePositions()
		}
	}
	return
}

func setValues(a *atom.AtomVar, values ...string) (p position.Position, ok bool) {
	var f atom.ValueFormatter
	_, p = a.GetNamePositions()
	p = p.Next('=')
	if a.GetType() == atom.VarString {
		c := atom.NewValue()
		if len(values) > 0 {
			v := strings.TrimSpace(values[0])
			if f, ok = atom.NewValueFormatter(v); !ok {
				return
			}
			c.SetRaw(v)
			c.SetFormat(f)
		}
		_, p = atom.RecomputePosition(c, p)
		a.SetValues(c)
		return
	}
	p = p.Next('(')
	var childs atom.Slice
	ok = true
	for i, v := range values {
		raw := v
		if v = strings.TrimSpace(v); len(v) == 0 {
			p = p.NextString(raw)
			continue
		}
		c := atom.NewValue()
		if f, ok = atom.NewValueFormatter(v); !ok {
			return
		}
		c.SetRaw(v)
		c.SetFormat(f)
		if i > 0 {
			p = p.Next(' ')
		}
		_, p = atom.RecomputePosition(c, p)
		childs.Push(c)
	}
	p = p.Next(')')
	a.SetValues(childs...)
	return
}

//SetValue set the variable of the given info with the given values.
//It returns false if the info is not valid or if it’s not a variable.
func (p *PKGBUILD) SetValue(info *atom.Info, values ...string) (ok bool) {
	if ok = info.IsVar() && p.ContainsInfo(info); !ok {
		return
	}
	a := info.AtomNamed.(*atom.AtomVar)
	oe := atom.GetEnd(a)
	var ne position.Position
	if ne, ok = setValues(a, values...); !ok {
		return
	}
	p.recomputeInfoPos(info, oe, ne)
	a.RecomputRaw()
	p.RecomputeValues()
	return
}

//AddVariable adds a variable with the given name, type and values at the
//end of the PKGBUILD.
//Its returns the associated info or false if add failed.
func (p *PKGBUILD) AddVariable(name string, array bool, values ...string) (info *atom.Info, ok bool) {
	if ok = len(name) > 0 && runes.CheckString(name, runes.IsAlphaNum); !ok {
		return
	}
	var a *atom.AtomVar
	if array {
		a = atom.NewArrayVar()
	} else {
		a = atom.NewStringVar()
	}
	a.SetName(name)
	b := position.Position{Line: 1}
	if l := len(p.atoms); l > 0 {
		b = atom.GetEnd(p.atoms[l-1]).Next('\n')
	}
	e := b.NextString(name)
	a.SetNamePositions(b, e)
	if e, ok = setValues(a, values...); !ok {
		return
	}
	a.SetPositions(b, e)
	a.RecomputRaw()
	p.atoms.Push(a)
	p.RecomputeInfos(true)
	return p.info.Get(a)
}

//AddFunction adds a variable with the given name and body at the
//end of the PKGBUILD.
//Its returns the associated info or false if add failed.
func (p *PKGBUILD) AddFunction(name string, body string) (info *atom.Info, ok bool) {
	raw := fmt.Sprintf("%s() %", name, body)
	b := position.Position{Line: 1}
	if l := len(p.atoms); l > 0 {
		b = atom.GetEnd(p.atoms[l-1]).Next('\n')
	}
	sc := scanner.NewFuncScanner(strings.NewReader(raw), b)
	if ok = sc.Scan(); !ok {
		return
	}
	a := sc.Atom()
	p.atoms.Push(a)
	p.RecomputeInfos()
	return p.info.Get(a)
}

//AddRaw parse the raw string into atoms and add them at the end
//of the PKGBUILD
//Its returns the associated infos or false if add failed.
func (p *PKGBUILD) AddRaw(raw string) (infos []*atom.Info, ok bool) {
	b := position.Position{Line: 1}
	if l := len(p.atoms); l > 0 {
		b = atom.GetEnd(p.atoms[l-1]).Next('\n')
	}
	sc := scanner.New(strings.NewReader(raw), b)
	atoms, err := scanner.ScanAll(sc)
	if ok = err == nil; !ok {
		return
	}
	p.atoms.Push(atoms...)
	p.RecomputeInfos(true)
	for _, a := range atoms {
		if info, exist := p.info.Get(a); exist {
			infos = append(infos, info)
		}
	}
	return
}

//GetIndex returns the info of the atom at the specified index.
//If index is invalid, ok is false.
//If it is a variable or a function, it returns the associated info.
//If it is a blank line, isBlank is true.
//If it is a comment line, isBlank is false.
func (p *PKGBUILD) GetIndex(i int) (info *atom.Info, isBlank bool, ok bool) {
	if ok = i >= 0 && i < len(p.atoms); !ok {
		return
	}
	var isNamed bool
	if info, isNamed = p.info.GetByIndex(i); !isNamed {
		isBlank = p.atoms[i].GetType() == atom.Blank
	}
	return
}

//RemoveIndex removes the atom at the specified index.
//It returns false if index is not valid.
func (p *PKGBUILD) RemoveIndex(i int, recomputePositions ...bool) bool {
	info, _, ok := p.GetIndex(i)
	if !ok {
		return false
	}
	p.atoms.Remove(i)
	recomputeVar := info != nil && info.IsVar()
	p.RecomputeInfos(recomputeVar)
	if len(recomputePositions) > 0 && recomputePositions[0] {
		p.RecomputePositions()
	}
	return true
}

//Len returns the number of the root atoms.
func (p *PKGBUILD) Len() int {
	return len(p.atoms)
}
