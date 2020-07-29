package atom

import (
	"sort"

	"github.com/bvaudour/kcp/position"
)

//GetNamed returns the named atom of the given atom if possible.
//If the atom is a named atom, it returns it.
//If the atom is a group, it returns the first named atom child and put
//isParent to true.
//Otherwise, exists is false.
func GetNamed(a Atom) (n AtomNamed, exists bool, isParent bool) {
	switch a.GetType() {
	case Function:
		n, exists = a.(*AtomFunc)
	case VarArray, VarString:
		n, exists = a.(*AtomVar)
	case Group:
		{
			var e *AtomGroup
			if e, isParent = a.(*AtomGroup); isParent {
				for _, c := range e.Childs {
					if n, exists, _ = GetNamed(c); exists {
						break
					}
				}
			}
		}
	}
	return
}

//Info packs properties of a named atom.
type Info struct {
	AtomNamed
	index  int
	parent Atom
	values []string
}

//NewInfo returns the infos of the given atom.
//If the atom is not a named atom or doesn’t contain
//a named atom, it returns false.
func NewInfo(a Atom) (info *Info, ok bool) {
	var n AtomNamed
	var isParent bool
	if n, ok, isParent = GetNamed(a); ok {
		info = &Info{AtomNamed: n}
		if isParent {
			info.parent = a
		}
	}
	return
}

//InfoList is a list of named infos in a slice of atoms.
type InfoList struct {
	infos  map[Atom]*Info
	values map[string]string
}

//NewInfoList returns an empty info list.
func NewInfoList() *InfoList {
	return &InfoList{
		infos:  make(map[Atom]*Info),
		values: make(map[string]string),
	}
}

//Keys returns the ilst of atoms where infos are found.
func (l *InfoList) Keys() Slice {
	var keys Slice
	for n := range l.infos {
		keys.Push(n)
	}
	sort.Slice(keys, func(i, j int) bool {
		ki, kj := keys[i], keys[j]
		return l.infos[ki].index < l.infos[kj].index
	})
	return keys
}

//Filter returns the infos of the atoms which pass le callback.
func (l *InfoList) Filter(cb NamedCheckerFunc) []*Info {
	var infos []*Info
	for _, n := range l.Keys() {
		info := l.infos[n]
		if cb == nil || cb(info.AtomNamed) {
			infos = append(infos, info)
		}
	}
	return infos
}

//FilterFirst returns the first atom which pass the callback
//or false if no atom was found.
func (l *InfoList) FilterFirst(cb NamedCheckerFunc) (info *Info, exists bool) {
	for _, n := range l.Keys() {
		i := l.infos[n]
		if cb == nil || cb(i.AtomNamed) {
			return i, true
		}
	}
	return
}

//Variables returns all infos of variable atoms.
func (l *InfoList) Variables() []*Info {
	return l.Filter(NewNameMatcher(VarArray, VarString))
}

//Functions returns all infos of function atoms.
func (l *InfoList) Functions() []*Info {
	return l.Filter(NewNameMatcher(Function))
}

//GetNamed returns the info of the given named atom
//or false if not found.
func (l *InfoList) GetNamed(n AtomNamed) (info *Info, exists bool) {
	for _, i := range l.infos {
		if n == i.AtomNamed {
			return i, true
		}
	}
	return
}

//Get is same as GetNamed but checks the container.
func (l *InfoList) Get(a Atom) (info *Info, exists bool) {
	info, exists = l.infos[a]
	return
}

//GetDeep returns the info of the atom wheter it is
//a named atom or a container, or false
//if not found.
func (l *InfoList) GetDeep(a Atom) (info *Info, exists bool) {
	if info, exists = l.Get(a); !exists {
		if n, ok, _ := GetNamed(a); ok {
			info, exists = l.GetNamed(n)
		}
	}
	return
}

//GetValues returns the list of the variable values
//indexed by the name of the variables.
func (l *InfoList) GetValues() map[string]string {
	out := make(map[string]string)
	for k, v := range l.values {
		out[k] = v
	}
	return out
}

//GetValue returns the value of the given variable name
//or an empty string if not found.
func (l *InfoList) GetValue(name string) string {
	return l.values[name]
}

//GetByIndex returns the info on the given index.
func (l *InfoList) GetByIndex(idx int) (info *Info, exists bool) {
	for _, k := range l.Keys() {
		i := l.infos[k]
		if exists = i.index == idx; exists {
			info = i
			return
		}
	}
	return
}

//HasValue returns true if the name is a variable name
//and if it has a value.
func (l *InfoList) HasValue(name string) bool {
	_, ok := l.values[name]
	return ok
}

//Update updates the info of the given atom with the specified index
//and returns a rune code specific to the type of update:
//- 'A' if the info didn’t exist and was added,
//- 'D' if the info existed and was deleted,
//- 'U' if the info was found and some properties were modified,
//- 'O' if there was no change.
func (l *InfoList) Update(a Atom, i int) (t rune) {
	infoNew, isInfo := NewInfo(a)
	info, exists := l.Get(a)
	if !isInfo {
		if exists {
			info.index = -1
			delete(l.infos, a)
			return 'D'
		}
		return 'O'
	}
	if !exists {
		info, exists = l.GetNamed(infoNew.AtomNamed)
	}
	if !exists {
		infoNew.index = i
		l.infos[a] = infoNew
		return 'A'
	}
	c := false
	old := Slice{info.AtomNamed, info.parent}
	if info.index != i {
		info.index = i
		c = true
	}
	if info.AtomNamed != infoNew.AtomNamed {
		info.AtomNamed = infoNew.AtomNamed
		info.values = info.values[:0]
		c = true
	}
	if info.parent != infoNew.parent {
		info.parent = infoNew.parent
		c = true
	}
	if !c {
		return 'O'
	}
	for _, e := range old {
		if e != nil {
			delete(l.infos, e)
		}
	}
	l.infos[a] = info
	return 'U'
}

//UpdateAll update the infos with the slice of atoms.
func (l *InfoList) UpdateAll(atoms Slice) {
	done := make(map[Atom]bool)
	for i, a := range atoms {
		l.Update(a, i)
		done[a] = true
		if n, ok, ip := GetNamed(a); ok && ip {
			done[n] = true
		}
	}
	for k, i := range l.infos {
		if !done[k] {
			i.index = -1
			i.values = nil
			delete(l.infos, k)
		}
	}
}

//RecomputeValues recomputes all values of all
//variable atoms.
func (l *InfoList) RecomputeValues() {
	l.values = make(map[string]string)
	for _, k := range l.Keys() {
		info := l.infos[k]
		if v, ok := info.Variable(); ok {
			info.values = v.GetArrayParsed(l.values)
			if len(info.values) > 0 {
				l.values[info.GetName()] = info.values[0]
			}
		} else {
			l.values = nil
		}
	}
}

//Begin returns the begin position of the named atom
//or the container.
func (i *Info) Begin() position.Position {
	return GetBegin(i)
}

//End returns the end position of the named atom
//or the container.
func (i *Info) End() position.Position {
	return GetEnd(i)
}

//Index returns the index position of the described
//atom in the slice.
func (i *Info) Index() int {
	return i.index
}

//Name returns the name of the named atom.
func (i *Info) Name() string {
	return i.GetName()
}

//Variable returns the variable atom described
//by the info, or false if it’s not a variable atom.
func (i *Info) Variable() (n *AtomVar, ok bool) {
	n, ok = i.AtomNamed.(*AtomVar)
	return
}

func (i *Info) v() *AtomVar {
	n, _ := i.Variable()
	return n
}

//Function returns the functon atom described
//by the info, or false if it’s not a function atom.
func (i *Info) Function() (n *AtomFunc, ok bool) {
	n, ok = i.AtomNamed.(*AtomFunc)
	return
}

func (i *Info) f() *AtomFunc {
	n, _ := i.Function()
	return n
}

//IsArrayVar returns true if the info concerns
//a variable of type array.
func (i *Info) IsArrayVar() bool {
	return i.GetType() == VarArray
}

//IsStringVar returns true if the info concerns
//a variable of type string.
func (i *Info) IsStringVar() bool {
	return i.GetType() == VarString
}

//IsSVar returns true if the info concerns
//a variable.
func (i *Info) IsVar() bool {
	return i.IsArrayVar() || i.IsStringVar()
}

//IsFunction returns true if the info concerns
//a function.
func (i *Info) IsFunc() bool {
	return i.GetType() == Function
}

//Raw returns the raw string of the described
//named atom.
func (i *Info) Raw() string {
	return i.GetRaw()
}

//StringValue returns the first value of the
//described atom or an empty string if it’s not
//a variable.
func (i *Info) StringValue() string {
	if len(i.values) > 0 {
		return i.values[0]
	}
	return ""
}

//ArrayValue returns the values of the
//described atom or an empty array if it’s not
//a variable.
func (i *Info) ArrayValue() []string {
	return i.values
}

//Body returns the raw content of the body function
//or an empty string if it’s not a function.
func (i *Info) Body() string {
	if i.IsFunc() {
		return i.f().GetBody()
	}
	return ""
}

//StringRawValue returns the first raw value of the
//described atom or an empty string if it’s not
//a variable.
func (i *Info) StringRawValue() string {
	if i.IsVar() {
		return i.v().GetStringValue()
	}
	return ""
}

//ArrayRawValue returns the raw values of the
//described atom or an empty array if it’s not
//a variable.
func (i *Info) ArrayRawValue() []string {
	if i.IsVar() {
		return i.v().GetArrayValue()
	}
	return nil
}
