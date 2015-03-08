package pkgbuild

import (
	"fmt"
	"regexp"
	"strings"
)

// Type of container/data
const (
	T_UNKNOWN int = iota
	T_HEADER
	T_BLANK
	T_COMMENT
	T_VARIABLE
	T_UVARIABLE
	T_FUNCTION
	T_UFUNCTION
)

// Name of container
const (
	HEADER       = "header"
	PKGBASE      = "pkgbase"
	PKGNAME      = "pkgname"
	PKGVER       = "pkgver"
	PKGREL       = "pkgrel"
	ARCH         = "arch"
	URL          = "url"
	EPOCH        = "epoch"
	PKGDESC      = "pkgdesc"
	LICENSE      = "license"
	GROUPS       = "groups"
	DEPENDS      = "depends"
	MAKEDEPENDS  = "makedepends"
	CHECKDEPENDS = "checkdepends"
	OPTDEPENDS   = "optdepends"
	PROVIDES     = "provides"
	CONFLICTS    = "conflicts"
	REPLACES     = "replaces"
	BACKUP       = "backup"
	OPTIONS      = "options"
	INSTALL      = "install"
	CHANGELOG    = "changelog"
	SOURCE       = "source"
	NOEXTRACT    = "noextract"
	MD5SUMS      = "md5sums"
	SHA1SUMS     = "sha1sums"
	SHA256SUMS   = "sha256sums"
	CHECK        = "check"
	PREPARE      = "prepare"
	BUILD        = "build"
	PACKAGE      = "package"
	SPLITPKG     = "package_"
	UNKNOWN      = "<unknown>"
	BLANKCOMMENT = "<blank>"
)

// Type of unparsing
const (
	U_SINGLE int = iota
	U_SINGLE_WITH_QUOTES
	U_OMULTI
	U_MULTI1
	U_MULTI1_WITHOUT_QUOTES
	U_MULTI2
	U_LINES
)

var mUnparse = map[string]int{
	HEADER:       U_LINES,
	PKGBASE:      U_SINGLE,
	PKGNAME:      U_OMULTI,
	PKGVER:       U_SINGLE,
	PKGREL:       U_SINGLE,
	ARCH:         U_MULTI1,
	URL:          U_SINGLE_WITH_QUOTES,
	EPOCH:        U_SINGLE,
	PKGDESC:      U_SINGLE_WITH_QUOTES,
	LICENSE:      U_MULTI1,
	GROUPS:       U_MULTI1,
	DEPENDS:      U_MULTI1,
	MAKEDEPENDS:  U_MULTI1,
	CHECKDEPENDS: U_MULTI1,
	OPTDEPENDS:   U_MULTI2,
	PROVIDES:     U_MULTI1,
	CONFLICTS:    U_MULTI1,
	REPLACES:     U_MULTI1,
	BACKUP:       U_MULTI1,
	OPTIONS:      U_MULTI1_WITHOUT_QUOTES,
	INSTALL:      U_SINGLE_WITH_QUOTES,
	CHANGELOG:    U_SINGLE_WITH_QUOTES,
	SOURCE:       U_MULTI2,
	NOEXTRACT:    U_MULTI1,
	MD5SUMS:      U_MULTI2,
	SHA1SUMS:     U_MULTI2,
	SHA256SUMS:   U_MULTI2,
	CHECK:        U_LINES,
	PREPARE:      U_LINES,
	BUILD:        U_LINES,
	PACKAGE:      U_LINES,
	SPLITPKG:     U_LINES,
	UNKNOWN:      U_LINES,
	BLANKCOMMENT: U_LINES,
}

var lstVar = []string{
	PKGBASE,
	PKGNAME,
	PKGVER,
	PKGREL,
	ARCH,
	URL,
	EPOCH,
	PKGDESC,
	LICENSE,
	GROUPS,
	DEPENDS,
	MAKEDEPENDS,
	CHECKDEPENDS,
	OPTDEPENDS,
	PROVIDES,
	CONFLICTS,
	REPLACES,
	BACKUP,
	OPTIONS,
	INSTALL,
	CHANGELOG,
	SOURCE,
	NOEXTRACT,
	MD5SUMS,
	SHA1SUMS,
	SHA256SUMS,
}

var lstFunc = []string{
	CHECK,
	PREPARE,
	BUILD,
	PACKAGE,
}

var regVar = regexp.MustCompile(`^(\S+)=(.*)$`)
var regFunc = regexp.MustCompile(`^(\S+)\s*\(\s*\)\s*\{.*`)

// Parse string into list of elements
func s2a(s string) []string {
	out := make([]string, 0)
	s = strings.Trim(s, "() ")
	e := make([]rune, 0)
	sep, ignore := "", false
	for _, c := range s {
		switch c {
		case ' ':
			if sep != "" {
				if ignore {
					e = append(e, '\\')
				}
				e = append(e, c)
			} else if len(e) > 0 {
				out = append(out, string(e))
				e = make([]rune, 0)
			}
		case '\'':
			switch sep {
			case string(c):
				if ignore {
					e = append(e, c)
				} else if len(e) > 0 {
					out = append(out, string(e))
					e = make([]rune, 0)
					sep = ""
				}
			case "":
				sep = string(c)
			default:
				e = append(e, c)
			}
		case '"':
			switch sep {
			case string(c):
				if ignore {
					e = append(e, c)
				} else if len(e) > 0 {
					out = append(out, string(e))
					e = make([]rune, 0)
					sep = ""
				}
			case "":
				sep = string(c)
			default:
				e = append(e, c)
			}
		case '\\':
			if ignore {
				e = append(e, c, c)
				ignore = false
			} else {
				ignore = true
			}
		default:
			if ignore {
				e = append(e, '\\')
			}
			e = append(e, c)
		}
		if ignore && c != '\\' {
			ignore = false
		}
	}
	return out
}

// Data description
type Data struct {
	Type  int
	Value string
}

func comments(d []*Data) ([]string, int) {
	out, p := make([]string, 0), -1
	for i, e := range d {
		if e.Type == T_COMMENT {
			out = append(out, e.Value)
		} else if e.Type == T_BLANK {
			if len(out) == 0 {
				out = append(out, e.Value)
			}
		} else {
			p = i
			break
		}
	}
	return out, p
}

func concat(d []*Data) string {
	s := ""
	for _, e := range d {
		s += e.Value
	}
	return s
}

func concatMulti(d []*Data) string {
	s := ""
	for _, e := range d {
		c := quotify(e.Value)
		if c != "" {
			if s != "" {
				s += " "
			}
			s += c
		}
	}
	return s
}

func quotify(s string) string {
	if s == "" {
		return s
	}
	q := ""
	if strings.Contains(s, "$") {
		q = "\""
		s = strings.Replace(s, "\"", "\\\"", -1)
	} else {
		q = "'"
		s = strings.Replace(s, "'", "\\'", -1)
	}
	return fmt.Sprintf("%s%s%s", q, s, q)
}

// Container description
type Container struct {
	Name  string
	Type  int
	Begin int
	End   int
	Data  []*Data
}

func (c *Container) Append(t int, d ...string) {
	for _, e := range d {
		c.Data = append(c.Data, &Data{t, e})
	}
}

func (c *Container) TypeOfUnparse() int {
	u, ok := mUnparse[c.Name]
	if ok {
		return u
	}
	if c.Type == T_UVARIABLE {
		return U_OMULTI
	}
	return U_LINES
}

func (c *Container) Unparse() []string {
	out, p := comments(c.Data)
	if p < 0 {
		return out
	}
	d := c.Data[p:]
	switch c.TypeOfUnparse() {
	case U_SINGLE:
		out = append(out, fmt.Sprintf("%s=%s", c.Name, concat(d)))
	case U_SINGLE_WITH_QUOTES:
		out = append(out, fmt.Sprintf("%s=%s", c.Name, quotify(concat(d))))
	case U_OMULTI:
		s := ""
		if len(d) < 2 {
			s = concat(d)
		} else {
			s = concatMulti(d)
			if s != "" {
				s = "(" + s + ")"
			}
		}
		out = append(out, fmt.Sprintf("%s=%s", c.Name, s))
	case U_MULTI1:
		out = append(out, fmt.Sprintf("%s=(%s)", c.Name, concatMulti(d)))
	case U_MULTI2:
		s := c.Name + "=" + "("
		t := strings.Repeat(" ", len(s))
		switch len(d) {
		case 0:
			out = append(out, s+")")
		case 1:
			out = append(out, fmt.Sprintf("%s%s)", s, quotify(d[0].Value)))
		default:
			s += quotify(d[0].Value)
			for _, e := range d[1:] {
				out = append(out, s)
				s = t + quotify(e.Value)
			}
			out = append(out, s+")")
		}
	case U_MULTI1_WITHOUT_QUOTES:
		s := ""
		for _, e := range d {
			if s != "" && e.Value != "" {
				s += " "
			}
			s += e.Value
		}
		out = append(out, fmt.Sprintf("%s=(%s)", c.Name, s))
	default:
		for _, e := range d {
			out = append(out, e.Value)
		}
		if c.Name != HEADER && out[0] != "" {
			l := out
			out = make([]string, len(l)+1)
			for i, e := range l {
				out[i+1] = e
			}
		}
	}
	return out
}

func (c1 *Container) Merge(c2 *Container) {
	c1.Data = append(c1.Data, c2.Data...)
}

func newContainer(l string) (*Container, int) {
	c, end := new(Container), 0
	c.Data = make([]*Data, 0)
	lt := strings.TrimSpace(l)
	switch {
	case lt == "":
		c.Name = BLANKCOMMENT
		c.Type = T_BLANK
		c.Append(c.Type, lt)
	case strings.HasPrefix(lt, "#"):
		c.Name = BLANKCOMMENT
		c.Type = T_COMMENT
		c.Append(c.Type, lt)
	case regVar.MatchString(lt):
		d := regVar.FindStringSubmatch(lt)
		c.Name = d[0]
		c.Type = T_UVARIABLE
		if strings.HasPrefix(d[1], "(") && !strings.HasSuffix(d[1], ")") {
			end++
		}
		for _, v := range lstVar {
			if v == c.Name {
				c.Type = T_VARIABLE
				break
			}
		}
		for _, v := range s2a(d[1]) {
			c.Append(c.Type, v)
		}
	case regFunc.MatchString(lt):
		d := regFunc.FindStringSubmatch(lt)
		c.Name = d[0]
		c.Type = T_UFUNCTION
		end = strings.Count(lt, "{") - strings.Count(lt, "}")
		for _, v := range lstFunc {
			if v == c.Name {
				c.Type = T_FUNCTION
				break
			}
		}
		c.Append(c.Type, lt)
	default:
		c.Name = UNKNOWN
		c.Type = T_UNKNOWN
		c.Append(c.Type, l)
	}
	return c, end
}

// PKGBUILD description
type Pkgbuild struct {
	Fields map[string][]*Container
}

func (p *Pkgbuild) Insert(c ...*Container) {
	for _, e := range c {
		d, ok := p.Fields[e.Name]
		if !ok {
			d = make([]*Container, 0, 1)
		}
		d = append(d, e)
		p.Fields[e.Name] = d
	}
}

func Parse(lines []string) *Pkgbuild {
	pkg := new(Pkgbuild)
	pkg.Fields = make(map[string][]*Container)
	end := 0
	var c *Container
	for i, l := range lines {
		switch {
		case c == nil:
			c, end = newContainer(l)
			c.Begin = i
			if c.Name == BLANKCOMMENT && len(pkg.Fields) == 0 {
				c.Name = HEADER
			}
		case c.Type == T_VARIABLE || c.Type == T_UVARIABLE:
			lt := strings.TrimSpace(l)
			c.Append(c.Type, s2a(lt)...)
			if strings.HasSuffix(lt, ")") {
				end = 0
			}
		case c.Type == T_FUNCTION || c.Type == T_UFUNCTION:
			c.Append(c.Type, l)
			end += strings.Count(l, "{")
			end -= strings.Count(l, "}")
		default:
			c2, e := newContainer(l)
			end = e
			c2.Begin = i
			switch {
			case c2.Name == BLANKCOMMENT:
				c.Merge(c2)
			case c2.Name == UNKNOWN:
				c.Name, c.Type = c2.Name, c2.Type
				c.Merge(c2)
			case c.Name == HEADER || c.Name == UNKNOWN:
				pkg.Insert(c)
				c = c2
			default:
				c.Name, c.Type = c2.Name, c2.Type
				c.Merge(c2)
			}
		}
		c.End = i
		if end == 0 {
			if c.Type == T_VARIABLE || c.Type == T_UVARIABLE || c.Type == T_FUNCTION || c.Type == T_UFUNCTION {
				pkg.Insert(c)
				c = nil
			}
		}
	}
	if c != nil {
		if end != 0 {
			c.Name, c.Type = UNKNOWN, T_UNKNOWN
		}
		pkg.Insert(c)
	}
	return pkg
}

func unparse1(cont map[string][]*Container, key string, lines []string) []string {
	if c, ok := cont[key]; ok {
		for _, e := range c {
			lines = append(lines, e.Unparse()...)
		}
		delete(cont, key)
	}
	return lines
}

func unparse2(cont map[string][]*Container, tpe int, lines []string) []string {
	for k, c := range cont {
		keep := make([]*Container, 0, len(c))
		for _, e := range c {
			if e.Type == tpe {
				lines = append(lines, e.Unparse()...)
			} else {
				keep = append(keep, e)
			}
		}
		if len(keep) == 0 {
			delete(cont, k)
		} else {
			cont[k] = keep
		}
	}
	return lines
}

func Unparse(pkg *Pkgbuild) []string {
	out := make([]string, len(pkg.Fields))
	cont := make(map[string][]*Container)
	for s, c := range pkg.Fields {
		cont[s] = c
	}
	out = unparse1(cont, HEADER, out)
	out = unparse1(cont, PKGBASE, out)
	out = unparse1(cont, PKGNAME, out)
	out = unparse1(cont, PKGVER, out)
	out = unparse1(cont, PKGREL, out)
	out = unparse1(cont, EPOCH, out)
	out = unparse1(cont, PKGDESC, out)
	out = unparse1(cont, ARCH, out)
	out = unparse1(cont, URL, out)
	out = unparse1(cont, LICENSE, out)
	out = unparse1(cont, GROUPS, out)
	out = unparse1(cont, DEPENDS, out)
	out = unparse1(cont, MAKEDEPENDS, out)
	out = unparse1(cont, CHECKDEPENDS, out)
	out = unparse1(cont, OPTDEPENDS, out)
	out = unparse1(cont, PROVIDES, out)
	out = unparse1(cont, CONFLICTS, out)
	out = unparse1(cont, REPLACES, out)
	out = unparse1(cont, BACKUP, out)
	out = unparse1(cont, OPTIONS, out)
	out = unparse1(cont, INSTALL, out)
	out = unparse1(cont, CHANGELOG, out)
	out = unparse1(cont, SOURCE, out)
	out = unparse1(cont, NOEXTRACT, out)
	out = unparse1(cont, MD5SUMS, out)
	out = unparse1(cont, SHA1SUMS, out)
	out = unparse1(cont, SHA256SUMS, out)
	out = unparse2(cont, T_UVARIABLE, out)
	out = unparse1(cont, PREPARE, out)
	out = unparse1(cont, BUILD, out)
	out = unparse1(cont, CHECK, out)
	out = unparse1(cont, PACKAGE, out)
	out = unparse2(cont, T_UFUNCTION, out)
	for _, c := range cont {
		for _, e := range c {
			out = append(out, e.Unparse()...)
		}
	}
	return out
}
