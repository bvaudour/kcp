package atom

import (
	"fmt"
	"strings"

	"github.com/bvaudour/kcp/position"
	"github.com/bvaudour/kcp/runes"
)

//Type is a type of atom.
type AtomType int

const (
	Unknown AtomType = iota
	Blank
	Comment
	Name
	Value
	Body
	Function
	VarString
	VarArray
	Group
)

//Atom represents an atomic part of a PKGBUILD.
type Atom interface {
	GetType() AtomType
	GetRaw() string
	GetPositions() (begin, end position.Position)
	SetType(AtomType)
	SetRaw(string)
	SetPositions(begin, end position.Position)
	Clone() Atom
}

//AtomNamed represents an atom of type declaration (function or variable).
type AtomNamed interface {
	Atom
	GetName() string
	GetNamePositions() (begin, end position.Position)
	SetName(string)
	SetNamePositions(begin, end position.Position)
}

type AtomCheckerFunc func(Atom) bool
type NamedCheckerFunc func(AtomNamed) bool

//TypeSet is a set of atoms’ types.
type TypeSet map[AtomType]bool

func NewTypeSet(types ...AtomType) TypeSet {
	ts := make(TypeSet)
	for _, t := range types {
		ts[t] = true
	}
	return ts
}

//Match checks if the given atom has a type contained in the set.
func (ts TypeSet) Match(a Atom) bool {
	return ts[a.GetType()]
}

//MatchNamed checks if the given atom declaration has a type contained in the set.
func (ts TypeSet) MatchNamed(a AtomNamed) bool {
	return ts.Match(a)
}

//RevCheck returns the inverse of the given callback.
func RevCheck(cb AtomCheckerFunc) AtomCheckerFunc {
	return func(a Atom) bool { return !cb(a) }
}

//RevNameCheck returns the inverse of the given callback.
func RevNameCheck(cb NamedCheckerFunc) NamedCheckerFunc {
	return func(a AtomNamed) bool { return !cb(a) }
}

//NewMatcher returns a callback checking if an atom has a supported type.
func NewMatcher(types ...AtomType) AtomCheckerFunc {
	ts := NewTypeSet(types...)
	return ts.Match
}

//NewNameMatcher returns a callback checking if an atom declaration has a supported type.
func NewNameMatcher(types ...AtomType) NamedCheckerFunc {
	ts := NewTypeSet(types...)
	return ts.MatchNamed
}

//NewRevMatcher returns a callback checking if an atom has not a supported type.
func NewRevMatcher(types ...AtomType) AtomCheckerFunc {
	return RevCheck(NewMatcher(types...))
}

//NewRevNameMatcher returns a callback checking if an atom declaration has not a supported type.
func NewRevNameMatcher(types ...AtomType) NamedCheckerFunc {
	return RevNameCheck(NewNameMatcher(types...))
}

//CheckName returns a callback to check if an atom declaration has the given name.
func CheckName(name string) NamedCheckerFunc {
	return func(n AtomNamed) bool {
		return n.GetName() == name
	}
}

//AtomCheckAll returns a callback checker which applies all given checkers.
func AtomCheckAll(checkers ...AtomCheckerFunc) AtomCheckerFunc {
	return func(a Atom) bool {
		for _, f := range checkers {
			if !f(a) {
				return false
			}
		}
		return true
	}
}

//NamedCheckAll returns a callback checker which applies all given checkers.
func NamedCheckAll(checkers ...NamedCheckerFunc) NamedCheckerFunc {
	return func(n AtomNamed) bool {
		for _, f := range checkers {
			if !f(n) {
				return false
			}
		}
		return true
	}
}

type Slice []Atom

//Push append the given atoms at the end of the slice.
func (l *Slice) Push(atoms ...Atom) {
	*l = append(*l, atoms...)
}

//PushFront append the given atoms at the beginning of the slice.
func (l *Slice) PushFront(atoms ...Atom) {
	*l = append(atoms, (*l)...)
}

//Insert appends the given atoms at the specified index of the slice.
//If the index is negative, it is counted from the end.
//For example, -1 is the last index of the slice.
func (l *Slice) Insert(idx int, atoms ...Atom) {
	s := len(*l)
	if idx < 0 {
		idx += s
	}
	if idx <= 0 {
		l.PushFront(atoms...)
	} else if idx >= s {
		l.Push(atoms...)
	}
	s2 := len(atoms)
	ns := make(Slice, s+s2)
	copy(ns[:idx], (*l)[:idx])
	copy(ns[idx:idx+s2], atoms)
	copy(ns[idx+s2:], (*l)[idx:])
	*l = ns
}

//Pop removes the last element of the slice and returns it.
//If the slice is empty, it returns false.
func (l *Slice) Pop() (a Atom, exists bool) {
	i := len(*l) - 1
	if exists = i >= 0; exists {
		a = (*l)[i]
		*l = (*l)[:i]
	}
	return
}

//PopFront removes the first element of the slice and returns it.
//If the slice is empty, it returns false.
func (l *Slice) PopFront() (a Atom, exists bool) {
	s := len(*l)
	if exists = s > 0; exists {
		a = (*l)[0]
		*l = (*l)[1:]
	}
	return
}

//Pop removes the element at the given index of the slice and returns it.
//If the slice is empty or the index is not valid, it returns false.
func (l *Slice) Remove(idx int) (a Atom, exists bool) {
	s := len(*l)
	if idx < 0 {
		idx += s
	}
	if exists = idx >= 0 && idx < s; exists {
		a = (*l)[idx]
		ns := make(Slice, s-1)
		copy(ns[:idx], (*l)[:idx])
		copy(ns[idx:], (*l)[idx+1:])
		*l = ns
	}
	return
}

//Search returns all atoms matching the given checker.
//If recursive option is given and is true, it makes a deep search
//in searching in the atom containing other atoms.
func (l Slice) Search(cb AtomCheckerFunc, optRecursive ...bool) (result Slice) {
	recursive := len(optRecursive) > 0 && optRecursive[0]
	for _, a := range l {
		if cb(a) {
			result.Push(a)
		}
		if recursive {
			childs := getChilds(a)
			result.Push(childs.Search(cb, recursive)...)
		}
	}
	return
}

//SearchFirst is same as Search but it returns only the first found atom.
//It returns false if no atom was found.
func (l Slice) SearchFirst(cb AtomCheckerFunc, optRecursive ...bool) (result Atom, exists bool) {
	recursive := len(optRecursive) > 0 && optRecursive[0]
	for _, a := range l {
		if cb(a) {
			return a, true
		}
		if recursive {
			childs := getChilds(a)
			if result, exists = childs.SearchFirst(cb, recursive); exists {
				return
			}
		}
	}
	return
}

//Filter returns all atoms matching one of the given types.
func (l Slice) Filter(types ...AtomType) Slice {
	return l.Search(NewMatcher(types...))
}

//FilterFirst returns the firt atom matching one of the given types.
//It returns false if no atom was found.
func (l Slice) FilterFirst(types ...AtomType) (Atom, bool) {
	return l.SearchFirst(NewMatcher(types...))
}

//FilterRecursive is same as Filter but it makes a deep search
//in searching in the atom containing other atoms.
func (l Slice) FilterRecursive(types ...AtomType) Slice {
	return l.Search(NewMatcher(types...), true)
}

//FilterFirstRecursive is same as FilterFirst but it makes a deep search
//in searching in the atom containing other atoms.
func (l Slice) FilterFirstRecursive(types ...AtomType) (Atom, bool) {
	return l.SearchFirst(NewMatcher(types...), true)
}

//Map apply a transfarmation function to each atom of the slice
//and returns the result of this transformation.
func (l Slice) Map(cb func(Atom) Atom) Slice {
	out := make(Slice, len(l))
	for i, a := range l {
		out[i] = cb(a)
	}
	return out
}

//Clone makes a copy of the slice.
func (l Slice) Clone() Slice {
	return l.Map(func(a Atom) Atom { return a.Clone() })
}

type atom struct {
	t AtomType
	r string
	b position.Position
	e position.Position
}

func (a *atom) GetType() AtomType                      { return a.t }
func (a *atom) GetRaw() string                         { return a.r }
func (a *atom) GetPositions() (b, e position.Position) { return a.b, a.e }
func (a *atom) SetType(t AtomType)                     { a.t = t }
func (a *atom) SetRaw(r string)                        { a.r = r }
func (a *atom) SetPositions(b, e position.Position)    { a.b, a.e = b, e }

func (a *atom) Copy() *atom {
	c := *a
	return &c
}
func (a *atom) Clone() Atom { return a.Copy() }

func newAtom(tpe AtomType) *atom {
	return &atom{t: tpe}
}

//NewComment returns an atom of type comment.
func NewComment() Atom { return newAtom(Comment) }

//NewBlank returns an atom of type blank.
func NewBlank() Atom { return newAtom(Blank) }

type atomNamed struct {
	*atom
	name Atom
}

func (a *atomNamed) GetName() string                            { return a.name.GetRaw() }
func (a *atomNamed) GetNamePositions() (b, e position.Position) { return a.name.GetPositions() }
func (a *atomNamed) SetName(n string)                           { a.name.SetRaw(n) }
func (a *atomNamed) SetNamePositions(b, e position.Position)    { a.name.SetPositions(b, e) }

func (a *atomNamed) Copy() *atomNamed {
	return &atomNamed{
		atom: a.atom.Copy(),
		name: a.name.Clone(),
	}
}
func (a *atomNamed) Clone() Atom { return a.Copy() }

func newAtomNamed(t AtomType) *atomNamed {
	return &atomNamed{
		atom: newAtom(t),
		name: newAtom(Name),
	}
}

//AtomFunc is a function declaration.
type AtomFunc struct {
	*atomNamed
	body Atom
}

func (a *AtomFunc) GetBody() string                            { return a.body.GetRaw() }
func (a *AtomFunc) GetBodyPositions() (b, e position.Position) { return a.body.GetPositions() }
func (a *AtomFunc) SetBody(b string)                           { a.body.SetRaw(b) }
func (a *AtomFunc) SetBodyPositions(b, e position.Position)    { a.body.SetPositions(b, e) }

func (a *AtomFunc) Copy() *AtomFunc {
	return &AtomFunc{
		atomNamed: a.atomNamed.Copy(),
		body:      a.body.Clone(),
	}
}
func (a *AtomFunc) Clone() Atom { return a.Copy() }

//RecomputeRaw recomputes the raw value of the function declaration
func (a *AtomFunc) RecomputeRaw() {
	var buffer strings.Builder
	b, e := a.GetPositions()
	nb, ne := a.GetNamePositions()
	bb, be := a.GetBodyPositions()
	buffer.WriteString(b.Blank(nb))
	buffer.WriteString(a.GetName())
	buffer.WriteString("()")
	ne = ne.NextString("()")
	buffer.WriteString(ne.Blank(bb))
	buffer.WriteString(a.GetBody())
	buffer.WriteString(be.Blank(e))
	a.SetRaw(buffer.String())
}

//FormatSpaces removes useless space before, between and after
//the name and the body of the function declaration.
//If recomputeRaw is true, the raw value of the atom is recomputed.
func (a *AtomFunc) FormatSpaces(recomputeRaw bool) {
	b := GetBegin(a)
	if c := b.Column; c != 0 {
		b = b.IncrementPosition(0, -c, -c)
	}
	name, body := strings.TrimSpace(a.GetName()), strings.TrimSpace(a.GetBody())
	a.SetName(name)
	a.SetBody(body)
	_, p := RecomputePosition(a.name, b)
	p = p.NextString("() ")
	_, e := RecomputePosition(a.body, p)
	a.SetPositions(b, e)
	if recomputeRaw {
		a.SetRaw(fmt.Sprintf("%s() %s", name, body))
	}
}

//NewFunction returns a new function declaration.
func NewFunction() *AtomFunc {
	return &AtomFunc{
		atomNamed: newAtomNamed(Function),
		body:      newAtom(Body),
	}
}

//ValueElement represents an atomic part of a value
//It can be a string of the name of a referenced variable.
type ValueElement struct {
	v string
	n bool
}

//String returns the string representation of the part of the value.
//It can be on the form:
//- part if the element is a raw string
//- ${name} if the element is a reference to a variable.
func (e ValueElement) String() string {
	if e.n {
		return fmt.Sprintf("${%s}", e.v)
	}
	return e.v
}

//Parse returns the computed string of the element
//using the given mapping (value of a known variable).
//If the element is a raw string, it returns it.
//Else it returns the value of the referenced variable
//(or an empty string if the variable doesn’t exist).
func (e ValueElement) Parse(depends map[string]string) string {
	if e.n {
		return depends[e.v]
	}
	return e.v
}

//Format formats the element according to the given quote rune.
//It applies a transformation in escaping characters when needed.
func (e ValueElement) Format(q rune) string {
	s := e.String()
	if e.n {
		return s
	}
	var buffer strings.Builder
	for _, r := range s {
		if runes.IsEscapable(r, q) {
			buffer.WriteRune('\\')
		}
		buffer.WriteRune(r)
	}
	return buffer.String()
}

//ValueFormatter represent parts of a value.
type ValueFormatter []ValueElement

//NewValueFormatter splits the given value
//in order to distinguish which part is a raw string
//and which part is a reference to a variable.
//It returns false if the parse failed.
func NewValueFormatter(s string) (f ValueFormatter, ok bool) {
	var buffer strings.Builder
	d := new(runes.Delimiter)
	var name, raw []rune
	ok = true
	for _, r := range s {
		if name, raw, ok = d.Parse(r); !ok {
			break
		}
		if len(name) > 0 {
			if buffer.Len() > 0 {
				f = append(f, ValueElement{v: buffer.String()})
				buffer.Reset()
			}
			f = append(f, ValueElement{v: string(name), n: true})
		}
		buffer.WriteString(string(raw))
	}
	if ok = d.IsClosed(); ok {
		var v string
		if d.IsShortVariableOpen() {
			v = d.GetVariableName()
			if len(v) == 0 {
				buffer.WriteRune('$')
			}
		}
		if buffer.Len() > 0 {
			f = append(f, ValueElement{v: buffer.String()})
		}
		if len(v) > 0 {
			f = append(f, ValueElement{v: v, n: true})
		}
	}
	return
}

//String returns the string representation of the formatter.
func (f ValueFormatter) String() string {
	out := make([]string, len(f))
	for i, e := range f {
		out[i] = e.String()
	}
	return strings.Join(out, "")
}

//Debug returns a string representation of the formatter (for debugging only).
func (f ValueFormatter) Debug() string {
	b := new(strings.Builder)
	fmt.Fprintln(b, "[")
	for i, e := range f {
		fmt.Fprintf(b, "  %d: %s (%v),\n", i, e.v, e.n)
	}
	fmt.Fprintln(b, "]")
	return b.String()
}

//Parse returns the computed value according to the value
//of the referenced variables.
func (f ValueFormatter) Parse(depends map[string]string) string {
	out := make([]string, len(f))
	for i, e := range f {
		out[i] = e.Parse(depends)
	}
	return strings.Join(out, "")
}

//GetDepends returns the list of the required variables’ names
//to fully compute the real value.
func (f ValueFormatter) GetDepends() (depends []string) {
	for _, e := range f {
		if e.n {
			depends = append(depends, e.v)
		}
	}
	return
}

//Clone returns a copy of the formatter.
func (f ValueFormatter) Clone() ValueFormatter {
	c := make(ValueFormatter, len(f))
	copy(c, f)
	return f
}

//HasDep returns true if the formmatter contains referenced variables.
func (f ValueFormatter) HasDep() bool {
	for _, e := range f {
		if e.n {
			return true
		}
	}
	return false
}

//Format formats the value according to the given quote rune.
//It escapes chars when needed.
func (f ValueFormatter) Format(q rune) string {
	var b strings.Builder
	if q != 0 {
		b.WriteRune(q)
	}
	for _, e := range f {
		b.WriteString(e.Format(q))
	}
	if q != 0 {
		b.WriteRune(q)
	}
	return b.String()
}

//AtomValue is an atom containing a value of a variable declaration.
type AtomValue struct {
	*atom
	f ValueFormatter
}

//GetValue returns the string representation of the formatted value
func (a *AtomValue) GetValue() string { return a.f.String() }

//GetParsed returns the real value, replacing the referenced variables by their values.
func (a *AtomValue) GetParsed(depends map[string]string) string { return a.f.Parse(depends) }

//GetDepends returns the needed variable to fully compute the value.
func (a *AtomValue) GetDepends() []string { return a.f.GetDepends() }

//GetFormatted formats a value with or without quote depending
//of the given parameter.
func (a *AtomValue) GetFormatted(quoteNeeded bool) string {
	var q rune
	if quoteNeeded {
		q = '\''
		if a.f.HasDep() || strings.ContainsRune(a.f.String(), '\'') {
			q = '"'
		}
	}
	return a.f.Format(q)
}

//SetFormat puts the format to use to compute and format the value.
func (a *AtomValue) SetFormat(f ValueFormatter) { a.f = f }

func (a *AtomValue) Copy() *AtomValue {
	return &AtomValue{
		atom: a.atom.Copy(),
		f:    a.f.Clone(),
	}
}
func (a *AtomValue) Clone() Atom { return a.Copy() }

//NewValue returns an otom of type value.
func NewValue() *AtomValue {
	return &AtomValue{
		atom: newAtom(Value),
	}
}

//AtomVar is a variable declaration.
type AtomVar struct {
	*atomNamed
	values Slice
}

//SetValues set the values atoms to the given atoms.
//The atoms must be of type value or comment (if it is
//an inner comment of the variable declaration).
func (a *AtomVar) SetValues(values ...Atom) { a.values = values }

//GetValues returns the value otams only (without comments).
func (a *AtomVar) GetValues() []*AtomValue {
	var out []*AtomValue
	for _, v := range a.values {
		if v.GetType() == Value {
			if e, ok := v.(*AtomValue); ok {
				out = append(out, e)
			}
		}
	}
	return out
}

//GetValue returns the first value otom or false
//if the declaration is empty of values.
func (a *AtomVar) GetValue() (e *AtomValue, exists bool) {
	if v, ok := a.values.FilterFirst(Value); ok {
		e, exists = v.(*AtomValue)
	}
	return
}

//GetDepends returns the list of the needed variables
//to compute the real values of the declaration.
func (a *AtomVar) GetDepends() map[string]bool {
	depends := make(map[string]bool)
	for _, v := range a.GetValues() {
		for _, d := range v.GetDepends() {
			depends[d] = true
		}
	}
	return depends
}

//GetArrayValue returns the list of the raw string of the values.
func (a *AtomVar) GetArrayValue() []string {
	values := a.GetValues()
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = v.GetRaw()
	}
	return out
}

//GetArrayFormatted returns the list of the beautified values
//with or without quotes depending of the given parameter.
func (a *AtomVar) GetArrayFormatted(quoteNeeded bool) []string {
	values := a.GetValues()
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = v.GetFormatted(quoteNeeded)
	}
	return out
}

//GetArrayParsed returns the list of real values
//computed with the known values of the depending variables.
func (a *AtomVar) GetArrayParsed(depends map[string]string) []string {
	values := a.GetValues()
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = v.GetParsed(depends)
	}
	return out
}

//GetStringValue is same as GetArrayValue but returns only the first entry
//or an empty string if no value).
func (a *AtomVar) GetStringValue() string {
	if v, ok := a.GetValue(); ok {
		return v.GetRaw()
	}
	return ""
}

//GetStringFormatted is same as GetArrayFormatted but returns only the first entry
//or an empty string if no value).
func (a *AtomVar) GetStringFormatted(quoteNeeded bool) string {
	if v, ok := a.GetValue(); ok {
		return v.GetFormatted(quoteNeeded)
	}
	return ""
}

//GetStringParsed is same as GetArrayParsed but returns only the first entry
//or an empty string if no value).
func (a *AtomVar) GetStringParsed(depends map[string]string) string {
	if v, ok := a.GetValue(); ok {
		return v.GetParsed(depends)
	}
	return ""
}

//RecomputeRaw recomputes the raw string of the atom.
func (a *AtomVar) RecomputRaw() {
	var buffer strings.Builder
	b, e := a.GetPositions()
	bn, _ := a.GetNamePositions()
	buffer.WriteString(b.Blank(bn))
	buffer.WriteString(a.GetName())
	buffer.WriteRune('=')
	var p position.Position
	if a.GetType() == VarString {
		buffer.WriteString(a.GetStringValue())
		p = b.NextString(buffer.String())
	} else {
		buffer.WriteRune('(')
		p = b.NextString(buffer.String())
		for _, v := range a.values {
			vb, ve := v.GetPositions()
			buffer.WriteString(p.Blank(vb))
			buffer.WriteString(v.GetRaw())
			p = ve
		}
		if l := len(a.values) - 1; l >= 0 && a.values[l].GetType() == Comment {
			buffer.WriteRune('\n')
			p = p.Next('\n')
		}
		buffer.WriteRune(')')
		p = p.Next(')')
	}
	buffer.WriteString(p.Blank(e))
	a.SetRaw(buffer.String())
}

//FormatSpaces removes useless spaces of the atom.
//If recomputeRaw is true, the raw string is recomputed.
func (a *AtomVar) FormatSpaces(recomputeRaw bool) {
	b := GetBegin(a)
	if c := b.Column; c != 0 {
		b = b.IncrementPosition(0, -c, -c)
	}
	name := strings.TrimSpace(a.GetName())
	a.SetName(name)
	_, p := RecomputePosition(a.name, b)
	p = p.Next('=')
	if a.GetType() == VarString {
		if v, ok := a.GetValue(); ok {
			_, p = RecomputePosition(v, p)
		}
	} else {
		p = p.Next('(')
		c := p.Column
		sc := strings.Repeat(" ", c)
		l := len(a.values) - 1
		for i, v := range a.values {
			t := v.GetType()
			vb, ve := v.GetPositions()
			if t == Comment {
				raw := strings.TrimSpace(v.GetRaw())
				v.SetRaw(raw)
				ve = vb.NextString(raw)
			}
			inc := ve.Diff(vb)
			if p.Column > c && (inc.Line != 0 || inc.Column+c > 80) {
				p = p.Prev(' ').Next('\n').NextString(sc)
			}
			_, p = RecomputePosition(v, p)
			if i != l {
				if t == Comment {
					p = p.Next('\n').NextString(sc)
				} else {
					p = p.Next(' ')
				}
			} else if t == Comment {
				p = p.Next('\n')
			}
		}
		p = p.Next(')')
	}
	a.SetPositions(b, p)
	if recomputeRaw {
		a.RecomputRaw()
	}
}

//RemoveComments removes all inner comment of the declaration.
func (a *AtomVar) RemoveComments(recomputeRaw bool) {
	if a.GetType() == VarString {
		a.values = a.values.Filter(Value)
		if recomputeRaw {
			a.RecomputRaw()
		}
		return
	}
	var values Slice
	var inc0 position.Increment
	for i, v := range a.values {
		if v.GetType() != Comment {
			values = append(values, v)
			continue
		}

		var inc position.Increment
		b, e := v.GetPositions()
		if i == 0 {
			_, b = a.GetNamePositions()
			b = b.NextString("=(")
			inc.Offset = b.Offset - e.Offset
		} else {
			e1 := GetEnd(a.values[i-1])
			if e1.Line == b.Line {
				b = e1
				inc.Offset = b.Offset - e.Offset
			} else {
				inc.Line--
				inc.Offset = b.Offset - e.Offset - b.Column - 1
			}
		}
		inc1 := inc
		for _, vn := range a.values[i+1:] {
			e0 := GetEnd(vn)
			_, e1 := IncrementPosition(vn, inc)
			inc1 = e1.Diff(e0)
		}
		inc0 = inc0.Increment(inc1)
	}
	b, e := a.GetPositions()
	e = e.Increment(inc0)
	a.SetPositions(b, e)
	if recomputeRaw {
		a.RecomputRaw()
	}
}

//FormatVariables reformats the values of the declaration, with or without quote
//depending to the value of the quoteNeeded parameter.
//If a type is given (should be of VarString or VarArray), the type of
//of the variable is forced.That means if an array variable becones a string variable,
//only the first value is kept.
//If recomputeRaw is true, the raw string of the declaration is recomputed.
func (a *AtomVar) FormatVariables(recomputeRaw, quoteNeeded bool, optType ...AtomType) {
	t := a.GetType()
	if len(optType) > 0 {
		t = optType[0]
	}

	isArray := t != VarString
	if isArray {
		t = VarArray
	}
	a.SetType(t)

	if !isArray {
		v, ok := a.GetValue()
		if !ok {
			v = NewValue()
		}
		_, b := a.GetNamePositions()
		b = b.Next('=')
		raw := v.GetFormatted(quoteNeeded)
		v.SetRaw(raw)
		_, e := RecomputePosition(v, b)
		b = GetBegin(a)
		a.SetValues(v)
		a.SetPositions(b, e)
		if recomputeRaw {
			a.RecomputRaw()
		}
		return
	}

	var inc0 position.Increment
	var sameLine bool
	if l := len(a.values) - 1; l >= 0 {
		ei := GetEnd(a.values[l])
		e := GetEnd(a)
		sameLine = e.Line == ei.Line
	}
	for i, v := range a.values {
		if v.GetType() == Comment {
			continue
		}
		e := v.(*AtomValue)
		raw := e.GetFormatted(quoteNeeded)
		e.SetRaw(raw)
		p0 := GetEnd(e)
		_, p1 := RecomputePosition(e)
		inc := p1.Diff(p0)
		inc1 := inc
		l := p0.Line
		for _, vn := range a.values[i+1:] {
			b0, e0 := vn.GetPositions()
			if b0.Line != l {
				inc.Column = 0
			}
			_, e1 := IncrementPosition(vn, inc)
			inc1 = e1.Diff(e0)
		}
		inc0 = inc0.Increment(inc1)
	}
	if !sameLine {
		inc0.Column = 0
	}
	b, e := a.GetPositions()
	e = e.Increment(inc0)
	a.SetPositions(b, e)
	if recomputeRaw {
		a.RecomputRaw()
	}
}

func (a *AtomVar) Copy() *AtomVar {
	return &AtomVar{
		atomNamed: a.atomNamed.Copy(),
		values:    a.values.Clone(),
	}
}
func (a *AtomVar) Clone() Atom { return a.Copy() }

func newVariable(tpe AtomType) *AtomVar {
	return &AtomVar{
		atomNamed: newAtomNamed(tpe),
	}
}

//NewStringVar returns a variable of type string.
func NewStringVar() *AtomVar {
	return newVariable(VarString)
}

//NewArrayVar returns a variable of type array.
func NewArrayVar() *AtomVar {
	return newVariable(VarArray)
}

//AtomGroup is a group of atoms
//It should contains:
//- one declaration atom (variable or function),
//- one comment.
type AtomGroup struct {
	*atom
	Childs Slice
}

func (a *AtomGroup) Copy() *AtomGroup {
	return &AtomGroup{
		atom:   a.atom.Copy(),
		Childs: a.Childs.Clone(),
	}
}
func (a *AtomGroup) Clone() Atom { return a.Copy() }

//RecomputeRaw recompute the raw string of the group.
func (a *AtomGroup) RecomputRaw() {
	var buffer strings.Builder
	p, e := a.GetPositions()
	for _, c := range a.Childs {
		b0, e0 := c.GetPositions()
		buffer.WriteString(p.Blank(b0))
		buffer.WriteString(c.GetRaw())
		p = e0
	}
	buffer.WriteString(p.Blank(e))
	a.SetRaw(buffer.String())
}

//FormatSpaces removes useless spaces before, between and
//after each childs of the group.
//If recomputeRaw is true, it recomputes the raw string
//of the group.
func (a *AtomGroup) FormatSpaces(recomputeRaw bool) {
	b := GetBegin(a)
	if c := b.Column; c != 0 {
		b = b.IncrementPosition(0, -c, -c)
	}
	p := b
	for i, c := range a.Childs {
		switch c.GetType() {
		case Function:
			c.(*AtomFunc).FormatSpaces(recomputeRaw)
		case VarArray, VarString:
			c.(*AtomVar).FormatSpaces(recomputeRaw)
		default:
			if recomputeRaw {
				raw := strings.TrimSpace(c.GetRaw())
				c.SetRaw(raw)
				RecomputePosition(c)
			}
		}
		if i != 0 {
			p = p.Next(' ')
		}
		_, p = MoveAtom(c, p)
	}
	a.SetPositions(b, p)
	if recomputeRaw {
		a.RecomputRaw()
	}
}

//NewGroup returns a new group atom.
func NewGroup() *AtomGroup {
	return &AtomGroup{
		atom: newAtom(Group),
	}
}

func getChilds(a Atom) Slice {
	var childs Slice
	switch a.GetType() {
	case Function:
		f := a.(*AtomFunc)
		childs = Slice{f.name, f.body}
	case VarArray, VarString:
		v := a.(*AtomVar)
		childs = Slice{v.name}
		childs.Push(v.values...)
	case Group:
		childs = a.(*AtomGroup).Childs
	}
	return childs
}

//IncrementPosition moves the given atom in incrementing the beginning
//and the end positions with the given increment.
//If recursive is needed, it moves the inner atoms too.
//It returns the new begin/end positions of the atom.
func IncrementPosition(a Atom, inc position.Increment, recursive ...bool) (newBegin, newEnd position.Position) {
	oldBegin, oldEnd := a.GetPositions()
	if len(recursive) > 0 && recursive[0] {
		childs := getChilds(a)
		for _, c := range childs {
			b := GetBegin(c)
			ic := inc
			if b.Line != oldBegin.Line {
				ic.Column = 0
			}
			IncrementPosition(c, ic)
		}
	}
	newBegin = oldBegin.Increment(inc)
	if oldBegin.Line != oldEnd.Line {
		inc.Column = 0
	}
	newEnd = oldEnd.Increment(inc)
	a.SetPositions(newBegin, newEnd)
	return
}

//MoveAtom moves the atom to the given position.
//If recursive is needed, it moves the inner atoms too.
//It returns the new begin/end positions of the atom.
func MoveAtom(a Atom, begin position.Position, recursive ...bool) (newBegin, newEnd position.Position) {
	b := GetBegin(a)
	return IncrementPosition(a, begin.Diff(b), recursive...)
}

func recomputePositions(a Atom, recursive bool, optBegin ...position.Position) (newBegin, newEnd position.Position) {
	oldBegin := GetBegin(a)
	newBegin = oldBegin
	if len(optBegin) > 0 {
		newBegin = optBegin[0]
	}
	p := newBegin
	if recursive {
		childs := getChilds(a)
		inc := newBegin.Diff(oldBegin)
		for _, c := range childs {
			b, e := c.GetPositions()
			if b.Line != p.Line {
				inc.Column = 0
			}
			_, p = recomputePositions(c, true, b.Increment(inc))
			inc = p.Diff(e)
		}
	}
	newEnd = newBegin.NextString(a.GetRaw())
	if newEnd.Lt(p) {
		newEnd = p
	}
	a.SetPositions(newBegin, newEnd)
	return
}

//RecomputePosition recomputes the end position of the atom according
//to the raw string.
//If a begin position is given, it moves the atom first before recomputing.
//It returns the new begin/end positions of the atom.
func RecomputePosition(a Atom, optBegin ...position.Position) (newBegin, newEnd position.Position) {
	return recomputePositions(a, false, optBegin...)
}

//RecomputeAllPositions is same as RecomputePosition but it recomputes the inner atoms too.
//It returns the new begin/end positions of the atom.
func RecomputeAllPositions(a Atom, optBegin ...position.Position) (newBegin, newEnd position.Position) {
	return recomputePositions(a, true, optBegin...)
}

//GetBegin returns the begin position of the atom.
func GetBegin(a Atom) position.Position {
	b, _ := a.GetPositions()
	return b
}

//GetEnd returns the end position of the atom.
func GetEnd(a Atom) position.Position {
	_, e := a.GetPositions()
	return e
}

//SetBegin put the begin position of the atom.
func SetBegin(a Atom, b position.Position) position.Position {
	p, e := a.GetPositions()
	a.SetPositions(b, e)
	return p
}

//SetEnd put the end position of the atom.
func SetEnd(a Atom, e position.Position) position.Position {
	b, p := a.GetPositions()
	a.SetPositions(b, e)
	return p
}

//Debug returns a string representation of the atom (for debugging only).
func Debug(a Atom) string {
	c0, c1 := a.GetPositions()
	r := a.GetRaw()
	return fmt.Sprintf(`%s -> %s:
%s`, c0, c1, r)
}
