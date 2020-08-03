package color

import (
	"fmt"
)

type Color int

const (
	NoColor Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
)

func (c Color) Light() string {
	if c == NoColor {
		return "\033[1m"
	}
	return fmt.Sprintf("\033[1;3%dm", c)
}

func (c Color) Dark() string {
	if c == NoColor {
		return "\033[m"
	}
	return fmt.Sprintf("\033[3%dm", c)
}

func (c Color) String() string {
	if c == NoColor {
		return c.Dark()
	}
	return c.Light()
}

func (c Color) Format(f string, args ...interface{}) string {
	return fmt.Sprint(c, fmt.Sprintf(f, args...), NoColor)
}

func (c Color) FormatDark(f string, args ...interface{}) string {
	return fmt.Sprint(c.Dark(), fmt.Sprintf(f, args...), NoColor)
}

func (c Color) Colorize(args ...interface{}) string {
	l := len(args)
	elts := make([]interface{}, l+2)
	elts[0] = c.Light()
	copy(elts[1:l+1], args)
	elts[l+1] = NoColor
	return fmt.Sprint(elts...)
}

func (c Color) ColorizeDark(args ...interface{}) string {
	l := len(args)
	elts := make([]interface{}, l+2)
	elts[0] = c.Dark()
	copy(elts[1:l+1], args)
	elts[l+1] = NoColor
	return fmt.Sprint(elts...)
}
