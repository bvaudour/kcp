package format

import (
	"sort"
	"strings"

	"github.com/bvaudour/kcp/pkgbuild/atom"
	"github.com/bvaudour/kcp/pkgbuild/standard"
)

type ntree struct {
	name   string
	childs []*ntree
}

type fOrder map[string]atom.Slice

func (o fOrder) keys() (keys []string) {
	for k := range o {
		keys = append(keys, k)
	}
	fs := make(map[string]int)
	for i, f := range standard.GetFunctions() {
		fs[f] = i
	}
	less := func(i, j int) bool {
		ki, kj := keys[i], keys[j]
		ii, ei := fs[ki]
		ij, ej := fs[kj]
		switch {
		case ei:
			if ej {
				return ii < ij
			}
			return true
		case ej:
			return false
		}
		return strings.Compare(ki, kj) < 0
	}
	sort.Slice(keys, less)
	return
}

func (o fOrder) order() (atoms atom.Slice) {
	for _, k := range o.keys() {
		atoms.Push(o[k]...)
	}
	return
}

type vBlock struct {
	name    string
	depends map[string]bool
	block   atom.Slice
}

func orderVkeys(keys []string) {
	vs := make(map[string]int)
	for i, v := range standard.GetVariables() {
		vs[v] = i
	}
	less := func(i, j int) bool {
		ki, kj := keys[i], keys[j]
		ii, ei := vs[ki]
		ij, ej := vs[kj]
		switch {
		case ei:
			if ej {
				return ii < ij
			}
			return true
		case ej:
			return false
		}
		return strings.Compare(ki, kj) < 0
	}
	sort.Slice(keys, less)
}

func newVBlock(info *atom.Info, block atom.Slice) *vBlock {
	out := vBlock{
		name:    info.Name(),
		block:   block,
		depends: make(map[string]bool),
	}
	vv, _ := info.Variable()
	for d := range vv.GetDepends() {
		out.depends[d] = true
	}
	return &out
}

func (v *vBlock) dependList() (depends []string) {
	for d, ok := range v.depends {
		if ok {
			depends = append(depends, d)
		}
	}
	orderVkeys(depends)
	return
}

type vOrder map[string][]*vBlock

func (o vOrder) keys() (keys []string) {
	for k := range o {
		keys = append(keys, k)
	}
	orderVkeys(keys)
	return
}

func (o vOrder) first(name string) *vBlock {
	return o[name][0]
}

func (o vOrder) pop(name string) bool {
	if len(o[name]) < 2 {
		delete(o, name)
		return true
	}
	o[name] = o[name][1:]
	return false
}

func (o vOrder) sequence(name string, done map[string]bool) (seq []string) {
	done[name] = true
	b := o.first(name)
	for d := range done {
		delete(b.depends, d)
	}
	if len(b.depends) == 0 {
		return []string{name}
	}
	for _, d := range b.dependList() {
		if done[d] {
			continue
		}
		if _, ok := o[d]; !ok {
			continue
		}
		seq = append(seq, o.sequence(d, done)...)
	}
	seq = append(seq, name)
	return
}

func (o vOrder) order() (atoms atom.Slice) {
	keys, done := o.keys(), make(map[string]bool)
	for len(keys) > 0 {
		seq := o.sequence(keys[0], done)
		for _, k := range seq {
			b := o.first(k)
			atoms = append(atoms, b.block...)
			if o.pop(k) {
				idx := -1
				for i, ki := range keys {
					if ki == k {
						idx = i
						break
					}
				}
				if idx >= 0 {
					ktmp := make([]string, len(keys)-1)
					copy(ktmp[:idx], keys[:idx])
					copy(ktmp[idx:], keys[idx+1:])
					keys = ktmp
				}
			}
		}
	}
	return
}
