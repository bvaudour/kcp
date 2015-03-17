package pkgbuild

import (
	"strings"
)

// Types of runes
func is_quote(r rune) bool {
	return r == '\'' || r == '"'
}

func is_open(r rune) bool {
	return strings.ContainsRune("({[", r)
}

func is_close(r rune) bool {
	return strings.ContainsRune(")}]", r)
}

func is_bracket(r rune) bool {
	return is_open(r) || is_close(r)
}

func is_blank(r rune) bool {
	return strings.ContainsRune(" \n\t\r", r)
}

// Map of correspondances
var m_bracket = map[rune]rune{
	'(': ')',
	'{': '}',
	'[': ']',
}

// Delimiters definition
type delimiter struct {
	q rune
	b []rune
}

func initd() *delimiter {
	return &delimiter{' ', []rune{}}
}

func (d *delimiter) quote_opened() bool {
	return is_quote(d.q)
}

func (d *delimiter) quote_closed() bool {
	return !d.quote_opened()
}

func (d *delimiter) bracket_opened() bool {
	return len(d.b) > 0
}

func (d *delimiter) bracket_closed() bool {
	return !d.bracket_opened()
}

func (d *delimiter) opened() bool {
	return d.quote_opened() || d.bracket_opened()
}

func (d *delimiter) closed() bool {
	return !d.opened()
}

func (d *delimiter) is_quote(r rune) bool {
	return d.quote_opened() && d.q == r
}

func (d *delimiter) is_bracket(r rune) bool {
	return d.bracket_opened() && d.b[len(d.b)-1] == r
}

func (d *delimiter) is_closer(r rune) bool {
	l := len(d.b) - 1
	if l < 0 || !is_close(r) {
		return false
	}
	return m_bracket[d.b[l]] == r
}

func (d *delimiter) set(r rune) bool {
	switch {
	case d.quote_opened():
		if d.is_quote(r) {
			d.q = ' '
			return true
		}
	case is_quote(r):
		d.q = r
		return true
	case d.is_closer(r):
		d.b = d.b[:len(d.b)-1]
		return true
	case is_open(r):
		d.b = append(d.b, r)
		return true
	}
	return false
}

func (d *delimiter) update(s string) {
	ig := false
	var p rune = ' '
	for _, c := range s {
		switch {
		case ig:
			ig = false
		case c == '\\':
			ig = true
		case d.quote_closed() && c == '#' && p == ' ':
			return
		default:
			d.set(c)
		}
		p = c
	}
}
