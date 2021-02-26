package scanner

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/bvaudour/kcp/pkgbuild/atom"
	"github.com/bvaudour/kcp/position"
	"github.com/bvaudour/kcp/runes"
)

//Scanner is an interface which wrap
//needed methods to scan a PKGBUILD:
//- Scan() analyzes a part of a PKGBUILD and returns false is the analyze failed
//- Error() returns the error after a scan failed
//- Atom() returns the last scanned atom
type Scanner interface {
	Scan() bool
	Error() error
	Atom() atom.Atom
}

type scbase struct {
	token
	pos     position.Position
	buffer  strings.Builder
	element atom.Atom
}

func (sc *scbase) raw() string {
	return sc.buffer.String()
}

func (sc *scbase) write(s string) position.Position {
	sc.buffer.WriteString(s)
	sc.pos = sc.pos.NextString(s)
	return sc.pos
}

func (sc *scbase) writeRune(r rune) position.Position {
	sc.buffer.WriteRune(r)
	sc.pos = sc.pos.Next(r)
	return sc.pos
}

func (sc *scbase) writeLast() position.Position {
	return sc.writeRune(sc.last)
}

func (sc *scbase) nextToken() (s string, tt tokenType, ok bool) {
	s, tt, ok = sc.text()
	sc.write(s)
	return
}

func (sc *scbase) setAtom(a atom.Atom, begin position.Position) bool {
	a.SetRaw(sc.raw())
	a.SetPositions(begin, sc.pos)
	sc.element = a
	return true
}

func (sc *scbase) setTrailing(a atom.Atom, begin position.Position) bool {
	if sc.state != tsTrailing {
		return sc.setAtom(a, begin)
	}
	raw := sc.raw()
	end, bc := sc.pos, sc.pos
	s, tt, ok := sc.nextToken()
	switch {
	case !ok:
		return false
	case sc.state != tsTrailing:
		return sc.setAtom(a, begin)
	case tt == ttSpace:
		bc = sc.pos
		if s, _, ok = sc.nextToken(); !ok {
			return false
		}
	}

	a.SetRaw(raw)
	a.SetPositions(begin, end)

	c := atom.NewComment()
	c.SetRaw(s)
	c.SetPositions(bc, sc.pos)

	g := atom.NewGroup()
	g.Childs = atom.Slice{a, c}
	return sc.setAtom(g, begin)
}

func (sc *scbase) getFunction(bn position.Position, name string) (a *atom.AtomFunc, ok bool) {
	var s string
	en := sc.pos
	if s, _, ok = sc.nextToken(); !ok {
		return
	}
	bb := sc.pos
	if s, _, ok = sc.nextToken(); ok {
		a = atom.NewFunction()
		a.SetName(name)
		a.SetPositions(bn, en)
		a.SetBody(s)
		a.SetBodyPositions(bb, sc.pos)
	}
	return
}

func (sc *scbase) passFunction() (ok bool) {
	if _, _, ok = sc.nextToken(); !ok {
		return
	}
	if _, _, ok = sc.nextToken(); !ok {
		return
	}
	var s string
	s, ok = sc.nextLine()
	sc.write(s)
	return
}

func (sc *scbase) getVariable(bn position.Position, name string) (a *atom.AtomVar, ok bool) {
	var s string
	en := sc.pos
	if s, _, ok = sc.nextToken(); !ok {
		return
	}
	bv := sc.pos
	var values atom.Slice
	if sc.state == tsVarSingle {
		a = atom.NewStringVar()
		if s, _, ok = sc.nextToken(); !ok {
			return
		}
		var f atom.ValueFormatter
		if f, ok = atom.NewValueFormatter(s); !ok {
			return
		}
		v := atom.NewValue()
		v.SetRaw(s)
		v.SetFormat(f)
		v.SetPositions(bv, sc.pos)
		values.Push(v)
	} else {
		a = atom.NewArrayVar()
		var tt tokenType
		s, tt, ok = sc.nextToken()
		for ok && tt != ttVarEnd {
			switch tt {
			case ttComment:
				v := atom.NewComment()
				v.SetRaw(s)
				v.SetPositions(bv, sc.pos)
				values.Push(v)
			case ttValue:
				var f atom.ValueFormatter
				if f, ok = atom.NewValueFormatter(s); !ok {
					return
				}
				v := atom.NewValue()
				v.SetRaw(s)
				v.SetFormat(f)
				v.SetPositions(bv, sc.pos)
				values.Push(v)
			}
			bv = sc.pos
			s, tt, ok = sc.nextToken()
		}
		if !ok {
			return
		}
	}
	a.SetName(name)
	a.SetNamePositions(bn, en)
	a.SetValues(values...)
	return
}

func (sc *scbase) passVariable() (ok bool) {
	if _, _, ok = sc.nextToken(); !ok {
		return
	}
	if sc.state == tsVarSingle {
		if _, _, ok = sc.nextToken(); !ok {
			return
		}
	} else {
		var tt tokenType
		_, tt, ok = sc.nextToken()
		for ok && tt != ttVarEnd {
			_, tt, ok = sc.nextToken()
		}
		if !ok {
			return
		}
	}
	var s string
	s, ok = sc.nextLine()
	sc.write(s)
	return
}

func (sc *scbase) Error() error {
	switch sc.err {
	case io.EOF, nil:
		return nil
	case runes.ErrInvalidToken:
		sc.writeLast()
	case runes.ErrUnendedToken:
	default:
		return sc.err
	}
	return fmt.Errorf(errorAt, sc.pos, sc.err, sc.buffer)
}

func (sc *scbase) Atom() atom.Atom {
	return sc.element
}

type scFull struct {
	scbase
}

func (sc *scFull) Scan() (ok bool) {
	sc.element = nil
	sc.buffer.Reset()
	var prefix, s string
	if prefix, ok = sc.init(); !ok {
		return
	}
	begin := sc.pos
	p := sc.write(prefix)
	var tt tokenType
	if s, tt, ok = sc.nextToken(); !ok {
		return
	}
	switch tt {
	case ttSpace:
		ok = sc.setAtom(atom.NewBlank(), begin)
	case ttComment:
		ok = sc.setAtom(atom.NewComment(), begin)
	case ttName:
		var a atom.Atom
		if sc.state == tsFunc {
			a, ok = sc.getFunction(p, s)
		} else {
			a, ok = sc.getVariable(p, s)
		}
		if ok {
			ok = sc.setTrailing(a, begin)
		}
	}
	if sc.state == tsUnknown {
		sc.writeRune('\n')
	}
	return
}

type scSlim struct {
	scbase
}

func (sc *scSlim) Scan() (ok bool) {
	sc.element = nil
	sc.buffer.Reset()
	var prefix, s string
	if prefix, ok = sc.init(); !ok {
		return
	}
	begin := sc.pos
	p := sc.write(prefix)
	var tt tokenType
	if s, tt, ok = sc.nextToken(); !ok {
		return
	}
	if tt != ttName {
		if sc.state == tsUnknown {
			sc.writeRune('\n')
		}
		return sc.Scan()
	}
	var a atom.Atom
	if sc.state == tsFunc {
		a, ok = sc.getFunction(p, s)
	} else {
		a, ok = sc.getVariable(p, s)
	}
	if ok {
		sc.setAtom(a, begin)
		s, ok = sc.nextLine()
		sc.write(s)
	}
	if sc.state == tsUnknown {
		sc.writeRune('\n')
	}
	return
}

type scVar struct {
	scbase
}

func (sc *scVar) Scan() (ok bool) {
	sc.element = nil
	sc.buffer.Reset()
	var prefix, s string
	if prefix, ok = sc.init(); !ok {
		return
	}
	begin := sc.pos
	p := sc.write(prefix)
	var tt tokenType
	if s, tt, ok = sc.nextToken(); !ok {
		return
	}
	if tt != ttName {
		if sc.state == tsUnknown {
			sc.writeRune('\n')
		}
		return sc.Scan()
	}
	var a atom.Atom
	if sc.state == tsFunc {
		if ok = sc.passFunction(); !ok {
			return
		}
		if sc.state == tsUnknown {
			sc.writeRune('\n')
		}
		return sc.Scan()
	} else {
		a, ok = sc.getVariable(p, s)
	}
	if ok {
		sc.setAtom(a, begin)
		s, ok = sc.nextLine()
		sc.write(s)
	}
	if sc.state == tsUnknown {
		sc.writeRune('\n')
	}
	return
}

type scFunc struct {
	scbase
}

func (sc *scFunc) Scan() (ok bool) {
	sc.element = nil
	sc.buffer.Reset()
	var prefix, s string
	if prefix, ok = sc.init(); !ok {
		return
	}
	begin := sc.pos
	p := sc.write(prefix)
	var tt tokenType
	if s, tt, ok = sc.nextToken(); !ok {
		return
	}
	if tt != ttName {
		if sc.state == tsUnknown {
			sc.writeRune('\n')
		}
		return sc.Scan()
	}
	var a atom.Atom
	if sc.state == tsVar {
		if ok = sc.passVariable(); !ok {
			return
		}
		if sc.state == tsUnknown {
			sc.writeRune('\n')
		}
		return sc.Scan()
	} else {
		a, ok = sc.getFunction(p, s)
	}
	if ok {
		sc.setAtom(a, begin)
		s, ok = sc.nextLine()
		sc.write(s)
	}
	if sc.state == tsUnknown {
		sc.writeRune('\n')
	}
	return
}

func (sc *scbase) new(r io.Reader, initialPos ...position.Position) {
	sc.source = bufio.NewReader(r)
	if len(initialPos) > 0 {
		sc.pos = initialPos[0]
	} else {
		sc.pos.Line++
	}
}

//New returns a full scanner of the given reader
//If a position is provided, the scan position start here.
//Otherwise the position is the initial position (line 1, column/offset 0).
func New(r io.Reader, initialPos ...position.Position) Scanner {
	sc := new(scFull)
	sc.new(r, initialPos...)
	return sc
}

//NewFastScanner is a scanner which ignores blank lines and comments.
func NewFastScanner(r io.Reader, initialPos ...position.Position) Scanner {
	sc := new(scSlim)
	sc.new(r, initialPos...)
	return sc
}

//NewVarScanner is a scanner which keeps only variable atoms.
func NewVarScanner(r io.Reader, initialPos ...position.Position) Scanner {
	sc := new(scVar)
	sc.new(r, initialPos...)
	return sc
}

//NewFuncScanner is a scanner which keeps only function atoms.
func NewFuncScanner(r io.Reader, initialPos ...position.Position) Scanner {
	sc := new(scFunc)
	sc.new(r, initialPos...)
	return sc
}

//ScanAll apply the Scan method of the scanner up to the end
//or until an error occurs.
//It returns the scanned atoms or an eventual error if scan failed.
func ScanAll(sc Scanner) (atoms atom.Slice, err error) {
	for sc.Scan() {
		atoms.Push(sc.Atom())
	}
	err = sc.Error()
	return
}
