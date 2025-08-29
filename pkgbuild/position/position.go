package position

import (
	"cmp"
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

// String returns the position representation.
func String(p syntax.Pos) string {
	return fmt.Sprintf("%d(%d)", p.Line(), p.Col())
}

// RangeString returns the represention of a positions’ range.
func RangeString(begin, end syntax.Pos) string {
	return fmt.Sprintf("%s-%s", String(begin), String(end))
}

// Cmp compares 2 positions.
func Cmp(p1, p2 syntax.Pos) int {
	if c := cmp.Compare(p1.Line(), p2.Line()); c != 0 {
		return c
	}
	return cmp.Compare(p1.Col(), p2.Col())
}
