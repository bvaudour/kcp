package pkgbuild

import (
	"strings"
)

func Parse(lines []string) Pkgbuild {
	p := NewPkgbuild()
	var c *Container
	var end int
	var begin bool
	for i, l := range lines {
		switch {
		case c == nil:
			begin = false
			c, end = NewContainer(l, i)
			if c.Type == TC_BLANKCOMMENT && len(p) == 0 {
				c.Type = TC_HEADER
				c.Name = HEADER
			} else if c.Type == TC_FUNCTION || c.Type == TC_UFUNCTION || c.Type == TC_SFUNCTION {
				if end == 0 && !strings.Contains(l, "{") {
					begin = true
				}
			}
		case c.Type == TC_UNKNOWN:
			c2, e := NewContainer(l, i)
			if c2.Type == TC_UNKNOWN {
				c.Values = append(c.Values, c2.Values...)
			} else {
				p.Insert(c)
				c, end = c2, e
			}
		case c.Type == TC_HEADER || c.Type == TC_BLANKCOMMENT:
			c2, e := NewContainer(l, i)
			if c2.Type == TC_BLANKCOMMENT {
				c.Values = append(c.Values, c2.Values...)
			} else {
				p.Insert(c)
				c, end = c2, e
				if c.Type == TC_FUNCTION || c.Type == TC_UFUNCTION || c.Type == TC_SFUNCTION {
					if end == 0 && !strings.Contains(l, "{") {
						begin = true
					}
				}
			}
		case c.Type == TC_VARIABLE || c.Type == TC_UVARIABLE:
			c2, _ := NewContainer(l, i)
			if c2.Type == TC_BLANKCOMMENT {
				c2.End = i
				p.Insert(c2)
			} else {
				c.Append(TD_VARIABLE, i, splitString(l)...)
				if strings.HasSuffix(strings.TrimSpace(l), ")") {
					end = 0
				}
			}
		default:
			c.Append(TD_FUNC, i, l)
			end += strings.Count(l, "{")
			end -= strings.Count(l, "}")
			if begin && strings.Contains(l, "{") {
				begin = false
			}
		}
		c.End = i
		if !begin && end == 0 && c.Type != TC_BLANKCOMMENT && c.Type != TC_HEADER && c.Type != TC_UNKNOWN {
			p.Insert(c)
			c = nil
		}
	}
	return p
}

func ParseFile(file string) Pkgbuild {
	if lines, e := readFile(file); e == nil {
		return Parse(lines)
	}
	return NewPkgbuild()
}

func Unparse(p Pkgbuild) []string {
	return p.Lines()
}

func UnparseInFile(p Pkgbuild, f string) error {
	return writeFile(f, Unparse(p))
}
