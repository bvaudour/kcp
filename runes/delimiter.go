package runes

import (
	"fmt"
)

var (
	closer = map[rune]rune{
		'(':  ')',
		'[':  ']',
		'{':  '}',
		'\'': '\'',
		'"':  '"',
	}
	opener = (func(m map[rune]rune) map[rune]rune {
		out := make(map[rune]rune)
		for k, v := range m {
			out[v] = k
		}
		return out
	})(closer)
)

//Delimiter is a structure which
//stores the state of the analyze
//of a string rune-by-rune.
type Delimiter struct {
	escape       bool
	variable     bool
	longVariable bool
	comment      bool
	variableName []rune
	quote        rune
	openDelim    []rune
}

//IsEscapeOpen returns true if the
//last analyzed rune is the escape rune (\).
func (d Delimiter) IsEscapeOpen() bool {
	return d.escape
}

//IsDelimOpen returns true if
//there are not closed delimiters ('(', '{' or '[').
func (d Delimiter) IsDelimOpen() bool {
	return len(d.openDelim) > 0
}

//IsQuoteOpen returns true
//if previous open quote (' or ") was not closed.
func (d Delimiter) IsQuoteOpen() bool {
	return IsQuote(d.quote)
}

//IsCommentOpen returns true if comment (#) is not ended.
func (d Delimiter) IsCommentOpen() bool {
	return d.comment
}

//IsShortVariableOpen returns true
//if a variable ($name) began to be parsed
//but was not flushed.
func (d Delimiter) IsShortVariableOpen() bool {
	return d.variable && !d.longVariable
}

//IsSLongVariableOpen returns true
//if a variable (${name}) began to be parsed
//but was not flushed.
func (d Delimiter) IsLongVariablOpen() bool {
	return d.longVariable
}

//IsVariableOpen returns true
//if a variable began to be parsed
//but was not flushed.
func (d Delimiter) IsVariableOpen() bool {
	return d.variable
}

//IsUnended returns true
//if the parsign was not finished.
func (d Delimiter) IsUnended() bool {
	return d.IsEscapeOpen() || d.IsDelimOpen() || d.IsQuoteOpen() || d.IsLongVariablOpen()
}

//IsClosed returns true
//if the string stopping parsing returns
//a valid string
func (d Delimiter) IsClosed() bool {
	return !d.IsUnended()
}

//IsFullyClosed is same as IsClosed
//but checks also variable and comment are flushed.
func (d Delimiter) IsFullyClosed() bool {
	return d.IsClosed() && !d.variable && !d.comment
}

//GetVariableName returns the current value of
//the parsed variable name.
func (d Delimiter) GetVariableName() string {
	return string(d.variableName)
}

//Parse reads the given rune and change the state of the delimiter.
//It flushed parsed runes, distinguishing variable name part and
//raw string part, and ok is false if parse failed.
func (d *Delimiter) Parse(r rune) (vname, out []rune, ok bool) {
	ok = true
	var needReparse, needCleanVar bool
	switch {
	case d.comment:
		d.comment = !IsNL(r)
	case d.variable:
		switch {
		case IsAlphaNum(r):
			d.variableName = append(d.variableName, r)
		case IsComment(r) && d.longVariable:
			d.variableName = append(d.variableName, r)
		case r == '{' && len(d.variableName) == 0:
			d.longVariable = true
		case len(d.variableName) == 0:
			if d.longVariable {
				out = []rune("${")
				if !d.IsQuoteOpen() {
					d.openDelim = append(d.openDelim, '{')
				}
			} else {
				out = []rune{'$'}
			}
			out = append(out, d.variableName...)
			needReparse, needCleanVar = true, true
		case !d.longVariable:
			needReparse = true
			fallthrough
		case r == '}':
			vname = append(vname, d.variableName...)
			needCleanVar = true
		default:
			out = []rune("${")
			out = append(out, d.variableName...)
			if !d.IsQuoteOpen() {
				d.openDelim = append(d.openDelim, '{')
			}
			needReparse, needCleanVar = true, true
		}
	case d.escape:
		d.escape = false
		if IsEscapable(r, d.quote) {
			out = []rune{r}
		} else {
			out = []rune{'\\', r}
		}
	case IsEscape(r):
		d.escape = true
	case d.IsQuoteOpen():
		switch {
		case r == d.quote:
			d.quote = 0
		case d.quote == '"' && IsVariable(r):
			d.variable = true
		default:
			out = []rune{r}
		}
	case IsQuote(r):
		d.quote = r
	case IsComment(r):
		d.comment = true
	case IsVariable(r):
		d.variable = true
	case IsOpen(r):
		d.openDelim = append(d.openDelim, r)
		out = []rune{r}
	case IsClose(r):
		i := len(d.openDelim) - 1
		if i < 0 || r != closer[d.openDelim[i]] {
			ok = false
		} else {
			d.openDelim = d.openDelim[:i]
			out = []rune{r}
		}
	default:
		out = []rune{r}
	}
	if needCleanVar {
		d.variable, d.longVariable = false, false
		d.variableName = d.variableName[:0]
	}
	if needReparse {
		var nextOut []rune
		_, nextOut, ok = d.Parse(r)
		out = append(out, nextOut...)
	}
	if !ok {
		vname, out = nil, nil
	}
	return
}

//Next is same as parses but returns only the (no-)failure of
//the parsing.
func (d *Delimiter) Next(r rune) bool {
	_, _, ok := d.Parse(r)
	return ok
}

//String returns the string representation of the delimiter state
//(for debug only).
func (d *Delimiter) String() string {
	v := ""
	if d.variable {
		v = "$"
		if d.longVariable {
			v = "${"
		}
	}
	v += string(d.variableName)
	return fmt.Sprintf(
		`D{
	escaped: %v,
	comment: %v,
	varOpen: %v,
	longOpen: %v,
	variable: "%s",
	quote: '%c',
	brackets: "%c",
	closed: %v,
}`,
		d.escape,
		d.comment,
		d.variable,
		d.longVariable,
		v,
		d.quote,
		d.openDelim,
		d.IsClosed(),
	)
}

//NewDelimiterChecker returns a new Delimiter plus
//an associated blockchecker callback.
//endChar gives a required condition to stop the parsing
//(the other condition is the delimiter is closed).
func NewDelimiterChecker(endChar CheckerFunc) (d *Delimiter, f BlockCheckerFunc) {
	d = new(Delimiter)
	f = func(r rune) (ok bool, err error) {
		if endChar(r) && d.IsClosed() {
			return
		}
		ok = d.Next(r)
		if !ok {
			err = ErrInvalidToken
		}
		return
	}
	return
}
