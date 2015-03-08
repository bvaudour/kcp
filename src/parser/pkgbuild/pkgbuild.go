package pkgbuild

import (
	"strings"
)

func Parse(lines []string) Pkgbuild {
	p := newPkgbuild()
	var c *Container
	var end int
	for i, l := range lines {
		switch {
		case c == nil:
			c, end = newContainer(l)
			c.Begin = i
			if c.Type == TC_BLANKCOMMENT && len(p) == 0 {
				c.Type = TC_HEADER
			}
		case c.Type == TC_UNKNOWN:
			c2, e := newContainer(l)
			if c2.Type == TC_UNKNOWN {
				c.Values = append(c.Values, c2.Values...)
			} else {
				p.Insert(c)
				c, end = c2, e
			}
		case c.Type == TC_HEADER || c.Type == TC_BLANKCOMMENT:
			c2, e := newContainer(l)
			if c2.Type == TC_BLANKCOMMENT {
				c.Values = append(c.Values, c2.Values...)
			} else {
				p.Insert(c)
				c, end = c2, e
			}
		case c.Type == TC_VARIABLE || c.Type == TC_UVARIABLE:
			c.Append(TD_VARIABLE, splitString(l)...)
			if strings.HasSuffix(strings.TrimSpace(l), ")") {
				end = 0
			}
		default:
			c.Append(TD_FUNC, l)
			end += strings.Count(l, "{")
			end -= strings.Count(l, "}")
		}
		if end == 0 && c.Type != TC_BLANKCOMMENT && c.Type != TC_HEADER && c.Type != TC_UNKNOWN {
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
	return newPkgbuild()
}

func Unparse(p Pkgbuild) []string {
	return p.Lines()
}

func UnparseInFile(p Pkgbuild, f string) error {
	return writeFile(f, Unparse(p))
}
