package scanner

import (
	"io"
	"strings"

	"github.com/bvaudour/kcp/runes"
)

type tokenState int
type tokenType int

const (
	tsUnknown tokenState = iota
	tsEnd
	tsVar
	tsVarSingle
	tsVarMultiple
	tsFunc
	tsFuncBody
	tsTrailing

	ttUnknown tokenType = iota
	ttSpace
	ttComment
	ttName
	ttNameType
	ttBody
	ttValue
	ttVarEnd
)

type token struct {
	state  tokenState
	last   rune
	source io.RuneReader
	err    error
}

func (t *token) next() bool {
	if t.err == nil {
		t.last, _, t.err = t.source.ReadRune()
	}
	if t.err != nil {
		t.state = tsEnd
	}
	return t.err == nil
}

func (t *token) nextString(cb runes.BlockCheckerFunc, ignoreFirst ...bool) (string, bool) {
	var buffer strings.Builder
	ok := true
	if len(ignoreFirst) > 0 && ignoreFirst[0] {
		buffer.WriteRune(t.last)
		ok = t.next()
	}
	for ok && t.err == nil {
		if ok, t.err = cb(t.last); ok {
			buffer.WriteRune(t.last)
			ok = t.next()
		}
	}
	if t.state != tsEnd && t.err != nil {
		t.state = tsEnd
	}
	ok = t.err == nil || t.err == io.EOF
	return buffer.String(), ok
}

func (t *token) nextSpace() (string, bool) {
	cb := runes.Checker2Block(runes.IsSpace)
	return t.nextString(cb)
}

func (t *token) nextBlank() (string, bool) {
	cb := runes.Checker2Block(runes.IsBlank)
	return t.nextString(cb)
}

func (t *token) nextLine() (string, bool) {
	cb := runes.RChecker2Block(runes.IsNL)
	return t.nextString(cb)
}

func (t *token) nextDelimited(endChar runes.CheckerFunc) (s string, ok bool) {
	d, cb := runes.NewDelimiterChecker(endChar)
	if s, ok = t.nextString(cb); ok {
		if ok = d.IsClosed(); !ok {
			t.err, t.state = runes.ErrUnendedToken, tsEnd
		}
	}
	return
}

func (t *token) init() (s string, ok bool) {
	if ok = t.next(); ok {
		if s, ok = t.nextSpace(); ok {
			t.state = tsUnknown
		}
	}
	return
}

func (t *token) readUnknown() (s string, tt tokenType, ok bool) {
	switch {
	case runes.IsNL(t.last) || t.err == io.EOF:
		tt, ok = ttSpace, true
		if t.err != nil {
			t.state = tsEnd
		}
	case runes.IsComment(t.last):
		if s, ok = t.nextLine(); ok {
			tt = ttComment
		}
	case runes.IsAlpha(t.last):
		cb := runes.Checker2Block(runes.IsAlphaNum)
		if s, ok = t.nextString(cb); !ok {
			return
		}
		if ok = t.err == nil; !ok {
			t.err, t.state = runes.ErrUnendedToken, tsEnd
			return
		}
		tt = ttName
		if runes.IsAffectation(t.last) {
			t.state = tsVar
		} else {
			t.state = tsFunc
		}
	default:
		t.err, t.state = runes.ErrInvalidToken, tsEnd
	}
	return
}

func (t *token) readVarType() (s string, tt tokenType, ok bool) {
	s, tt, ok = "=(", ttNameType, true
	t.state = tsVarMultiple
	if !t.next() || t.last != '(' {
		s, t.state = "=", tsVarSingle
	} else if !t.next() {
		if t.err == io.EOF {
			t.err = runes.ErrUnendedToken
		}
		tt = ttUnknown
	}
	return
}

func (t *token) readVarSingle() (s string, tt tokenType, ok bool) {
	tt, ok = ttValue, true
	switch {
	case t.err == io.EOF:
		t.state = tsEnd
	case runes.IsNL(t.last):
		t.state = tsUnknown
	case runes.IsSpace(t.last) || runes.IsComment(t.last):
		t.state = tsTrailing
	default:
		endChar := func(r rune) bool { return runes.IsComment(r) || runes.IsBlank(r) }
		if s, ok = t.nextDelimited(endChar); !ok {
			tt = ttUnknown
		} else {
			t.readVarSingle()
		}
	}
	return
}

func (t *token) readVarMultiple() (s string, tt tokenType, ok bool) {
	switch {
	case t.last == ')':
		s, tt, ok = ")", ttVarEnd, true
		if t.next() {
			if runes.IsNL(t.last) {
				t.state = tsUnknown
			} else {
				t.state = tsTrailing
			}
		}
		return
	case runes.IsComment(t.last):
		if s, ok = t.nextLine(); ok {
			tt = ttComment
		}
	case runes.IsBlank(t.last):
		if s, ok = t.nextBlank(); ok {
			tt = ttSpace
		}
	default:
		endChar := func(r rune) bool { return r == ')' || runes.IsComment(r) || runes.IsBlank(r) }
		if s, ok = t.nextDelimited(endChar); ok {
			tt = ttValue
		}
	}
	if ok {
		if ok = t.err == nil; !ok {
			tt = ttUnknown
			t.err = runes.ErrUnendedToken
		}
	}
	return
}

func (t *token) readFuncType() (s string, tt tokenType, ok bool) {
	last := rune(' ')
	cb := func(r rune) (ok bool, err error) {
		switch {
		case runes.IsBlank(r):
			ok = true
		case r == '(':
			if ok = last == ' '; ok {
				last = r
			} else {
				err = runes.ErrInvalidToken
			}
		case r == ')':
			if ok = last == '('; ok {
				last = r
			} else {
				err = runes.ErrInvalidToken
			}
		case r == '{':
			if last != ')' {
				err = runes.ErrInvalidToken
			}
		default:
			err = runes.ErrInvalidToken
		}
		return
	}
	if s, ok = t.nextString(cb); ok {
		if ok = t.err == nil; ok {
			tt, t.state = ttSpace, tsFuncBody
		} else {
			t.err = runes.ErrUnendedToken
		}
	}
	return
}

func (t *token) readFuncBody() (s string, tt tokenType, ok bool) {
	endChar := func(r rune) bool { return runes.IsComment(r) || runes.IsBlank(r) }
	if s, ok = t.nextDelimited(endChar); ok {
		tt = ttBody
		switch {
		case t.err == io.EOF:
			t.state = tsEnd
		case runes.IsNL(t.last):
			t.state = tsUnknown
		case runes.IsSpace(t.last) || runes.IsComment(t.last):
			t.state = tsTrailing
		}
	}
	return
}

func (t *token) readTrailing() (s string, tt tokenType, ok bool) {
	switch {
	case runes.IsSpace(t.last):
		if s, ok = t.nextSpace(); ok {
			tt = ttSpace
		}
	case runes.IsComment(t.last):
		if s, ok = t.nextLine(); ok {
			tt = ttComment
		}
	default:
		t.err, t.state = runes.ErrInvalidToken, tsEnd
	}
	if t.err == nil && runes.IsNL(t.last) {
		t.state = tsUnknown
	}
	return
}

func (t *token) text() (s string, tt tokenType, ok bool) {
	switch t.state {
	case tsUnknown:
		return t.readUnknown()
	case tsVar:
		return t.readVarType()
	case tsVarSingle:
		return t.readVarSingle()
	case tsVarMultiple:
		return t.readVarMultiple()
	case tsFunc:
		return t.readFuncType()
	case tsFuncBody:
		return t.readFuncBody()
	case tsTrailing:
		return t.readTrailing()
	}
	return
}
