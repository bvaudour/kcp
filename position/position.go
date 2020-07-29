package position

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

//Position represents the position of a character in a file.
type Position struct {
	Offset int
	Line   int
	Column int
}

//Increment is an incrementer of positions
type Increment = Position

//IsValid returns true if the position is valid.
//To be valid, a position should have lines > 0 and columns/offset ≥ 0.
func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column >= 0 && p.Offset >= 0
}

//Next retuns the position next to the given character.
func (p Position) Next(r rune) Position {
	w := utf8.RuneLen(r)
	if w <= 0 {
		return p
	}
	p.Offset += w
	if r == '\n' {
		p.Line++
		p.Column = 0
	} else {
		p.Column++
	}
	return p
}

//Prev returns the position before the given character.
func (p Position) Prev(r rune) Position {
	w := utf8.RuneLen(r)
	if r == '\n' || w <= 0 {
		return p
	}
	p.Offset -= w
	p.Column--
	return p
}

//NextString returns the cursor position next to the given string.
func (p Position) NextString(s string) Position {
	for _, r := range s {
		p = p.Next(r)
	}
	return p
}

//String returns a string representation of the position
//on the form L{line},C{column}
func (p Position) String() string {
	return fmt.Sprintf("L%d,C%d", p.Line, p.Column)
}

//Increment returns a new position with line, column and offset
//incremented by the given increment.
func (p Position) Increment(inc Increment) Position {
	p.Line += inc.Line
	p.Column += inc.Column
	p.Offset += inc.Offset
	return p
}

//IncrementPosition returns a new position with line, column and offset
//incremented by the given increments.
func (p Position) IncrementPosition(incLine, incColumn, incOffset int) Position {
	return p.Increment(New(incLine, incColumn, incOffset))
}

//New returns a position with the given line, column and offset.
func New(line, column, offset int) Position {
	return Position{
		Line:   line,
		Column: column,
		Offset: offset,
	}
}

//Cmp compares 2 positions. It returns :
//- -1 if p1 before p2
//- 1 if p1 after p2
//- 0 otherwise
func (p1 Position) Cmp(p2 Position) int {
	switch {
	case p1.Line < p2.Line:
		return -1
	case p1.Line > p2.Line:
		return 1
	case p1.Column < p2.Column:
		return -1
	case p1.Column > p2.Column:
		return 1
	}
	return 0
}

func (p1 Position) Eq(p2 Position) bool {
	return p1.Cmp(p2) == 0
}

func (p1 Position) Ne(p2 Position) bool {
	return p1.Cmp(p2) != 0
}

func (p1 Position) Ge(p2 Position) bool {
	return p1.Cmp(p2) >= 0
}

func (p1 Position) Le(p2 Position) bool {
	return p1.Cmp(p2) <= 0
}

func (p1 Position) Gt(p2 Position) bool {
	return p1.Cmp(p2) > 0
}

func (p1 Position) Lt(p2 Position) bool {
	return p1.Cmp(p2) < 0
}

func (p Position) Between(p1, p2 Position) bool {
	return p.Ge(p1) && p.Le(p2)
}

//Diff returns p1 - p2
func (p1 Position) Diff(p2 Position) Increment {
	return New(
		p1.Line-p2.Line,
		p1.Column-p2.Column,
		p1.Offset-p2.Offset,
	)
}

//Blank returns a string corresponding to the space
//between p1 and p2.
func (p1 Position) Blank(p2 Position) string {
	inc := p2.Diff(p1)
	if inc.Line == 0 && inc.Column > 0 {
		return strings.Repeat(" ", inc.Column)
	} else if inc.Line > 0 {
		s := strings.Repeat("\n", inc.Line)
		if p2.Column > 0 {
			s += strings.Repeat(" ", p2.Column)
		}
		return s
	}
	return ""
}
