package format

import (
	"strings"

	"github.com/bvaudour/kcp/pkgbuild/atom"
	"github.com/bvaudour/kcp/pkgbuild/standard"
)

//MapFunc is a function which apply
//transformations to a slice of atoms
//and returns the result of these transformations.
type MapFunc func(atom.Slice) atom.Slice

//Formatter is an interface
//which implements a transformation method
type Formatter interface {
	Map(atom.Slice) atom.Slice
}

type StdFormatter int

const (
	RemoveExtraSpaces StdFormatter = iota
	RemoveBlankLinesExceptFirst
	RemoveAllBlankLines
	RemoveAdjacentDuplicateBlankLines
	RemoveCommentLines
	RemoveTrailingComments
	RemoveVariableComments
	RemoveDuplicateVariables
	RemoveDuplicateFunctions
	FormatValues
	BeautifulValues
	ReorderFuncsAndVars
	AddBlankLineBeforeVariables
	AddBlankLineBeforeFunctions
)

var mstd = map[StdFormatter]MapFunc{
	RemoveExtraSpaces: func(in atom.Slice) (out atom.Slice) {
		cb := func(a atom.Atom) atom.Atom {
			switch a.GetType() {
			case atom.Function:
				a.(*atom.AtomFunc).FormatSpaces(true)
			case atom.VarArray, atom.VarString:
				a.(*atom.AtomVar).FormatSpaces(true)
			case atom.Group:
				a.(*atom.AtomGroup).FormatSpaces(true)
			default:
				raw := strings.TrimSpace(a.GetRaw())
				a.SetRaw(raw)
				atom.RecomputePosition(a)
			}
			return a
		}
		return in.Map(cb)
	},
	RemoveBlankLinesExceptFirst: func(in atom.Slice) (out atom.Slice) {
		check := atom.NewMatcher(atom.Blank)
		for i, a := range in {
			if i == 0 || !check(a) {
				out.Push(a)
			}
		}
		return
	},
	RemoveAllBlankLines: func(in atom.Slice) (out atom.Slice) {
		check := atom.NewRevMatcher(atom.Blank)
		return in.Search(check)
	},
	RemoveAdjacentDuplicateBlankLines: func(in atom.Slice) (out atom.Slice) {
		check := atom.NewMatcher(atom.Blank)
		for i, a := range in {
			if !(i > 0 && check(a) && check(in[i-1])) {
				out.Push(a)
			}
		}
		return
	},
	RemoveCommentLines: func(in atom.Slice) (out atom.Slice) {
		check := atom.NewRevMatcher(atom.Comment)
		return in.Search(check)
	},
	RemoveTrailingComments: func(in atom.Slice) (out atom.Slice) {
		check := atom.NewMatcher(atom.Group)
		cc := atom.NewMatcher(atom.Comment)
		for _, a := range in {
			if check(a) {
				e := a.(*atom.AtomGroup)
				last := len(e.Childs) - 1
				for last >= 0 && cc(e.Childs[last]) {
					e.Childs.Pop()
					last--
				}
				if last < 0 {
					continue
				} else if last == 0 {
					out.Push(e.Childs[0])
					continue
				}
			}
			out.Push(a)
		}
		return
	},
	RemoveVariableComments: func(in atom.Slice) (out atom.Slice) {
		cb := func(a atom.Atom) {
			if e, ok := a.(*atom.AtomVar); ok {
				e.RemoveComments(true)
			}
		}
		out = make(atom.Slice, len(in))
		for i, a := range in {
			if info, ok := atom.NewInfo(a); ok && info.IsVar() {
				cb(info.AtomNamed)
			}
			out[i] = a
		}
		return
	},
	RemoveDuplicateVariables: func(in atom.Slice) (out atom.Slice) {
		done := make(map[string]bool)
		check := func(a atom.Atom) bool {
			info, ok := atom.NewInfo(a)
			if ok && info.IsVar() {
				if name := info.Name(); done[name] {
					return false
				} else {
					done[name] = true
				}
			}
			return true
		}
		out = in.Search(check)
		return
	},
	RemoveDuplicateFunctions: func(in atom.Slice) (out atom.Slice) {
		done := make(map[string]bool)
		check := func(a atom.Atom) bool {
			info, ok := atom.NewInfo(a)
			if ok && info.IsFunc() {
				if name := info.Name(); done[name] {
					return false
				} else {
					done[name] = true
				}
			}
			return true
		}
		out = in.Search(check)
		return
	},
	FormatValues: func(in atom.Slice) (out atom.Slice) {
		cb := func(a atom.Atom) atom.Atom {
			info, ok := atom.NewInfo(a)
			if !ok {
				return a
			}
			e, ok := info.Variable()
			if !ok {
				return a
			}
			name := info.Name()
			qn := standard.IsQuotedVariable(name)
			var t []atom.AtomType
			if standard.IsStandardVariable(name) {
				if standard.IsArrayVariable(name) {
					t = append(t, atom.VarArray)
				} else {
					t = append(t, atom.VarString)
				}
			}
			e.FormatVariables(true, qn, t...)
			return a
		}
		out = in.Map(cb)
		return
	},
	BeautifulValues: func(in atom.Slice) (out atom.Slice) {
		cb := func(a atom.Atom) atom.Atom {
			info, ok := atom.NewInfo(a)
			if !ok {
				return a
			}
			e, ok := info.Variable()
			if !ok {
				return a
			}
			name := info.Name()
			qn := standard.IsQuotedVariable(name)
			var t []atom.AtomType
			if standard.IsStandardVariable(name) {
				if standard.IsArrayVariable(name) {
					t = append(t, atom.VarArray)
				} else {
					t = append(t, atom.VarString)
				}
			}
			e.RemoveComments(false)
			e.FormatVariables(false, qn, t...)
			e.FormatSpaces(true)
			return a
		}
		out = in.Map(cb)
		return
	},
	ReorderFuncsAndVars: func(in atom.Slice) (out atom.Slice) {
		var header, block atom.Slice
		fo, vo := make(fOrder), make(vOrder)
		for _, a := range in {
			info, ok := atom.NewInfo(a)
			if !ok {
				block.Push(a)
				continue
			}
			if len(header) == 0 && len(fo) == 0 && len(vo) == 0 {
				header.Push(block...)
				block = nil
			}
			block.Push(a)
			name := info.Name()
			if info.IsFunc() {
				f := fo[name]
				f.Push(block...)
				fo[name] = f
			} else {
				v := vo[name]
				vo[name] = append(v, newVBlock(info, block))
			}
			block = nil
		}
		out.Push(header...)
		out.Push(vo.order()...)
		out.Push(fo.order()...)
		out.Push(block...)
		return
	},
	AddBlankLineBeforeVariables: func(in atom.Slice) (out atom.Slice) {
		cl := atom.NewMatcher(atom.Blank)
		check := func(a atom.Atom) bool {
			info, ok := atom.NewInfo(a)
			return ok && info.IsVar()
		}
		for i, a := range in {
			if i > 0 && !cl(in[i-1]) && check(a) {
				out.Push(atom.NewBlank())
			}
			out.Push(a)
		}
		return
	},
	AddBlankLineBeforeFunctions: func(in atom.Slice) (out atom.Slice) {
		cl := atom.NewMatcher(atom.Blank)
		check := func(a atom.Atom) bool {
			info, ok := atom.NewInfo(a)
			return ok && info.IsFunc()
		}
		for i, a := range in {
			if i > 0 && !cl(in[i-1]) && check(a) {
				out.Push(atom.NewBlank())
			}
			out.Push(a)
		}
		return
	},
}

func (f StdFormatter) Map(in atom.Slice) atom.Slice {
	if cb, ok := mstd[f]; ok {
		return cb(in)
	}
	return in
}

func Format(formats ...Formatter) MapFunc {
	return func(atoms atom.Slice) atom.Slice {
		for _, f := range formats {
			atoms = f.Map(atoms)
		}
		return atoms
	}
}
