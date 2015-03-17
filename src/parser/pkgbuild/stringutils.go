package pkgbuild

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Line matchers
var r_blank = regexp.MustCompile(`^$`)
var r_comment = regexp.MustCompile(`^(#.*)$`)
var r_variable = regexp.MustCompile(`^(\w+)=(.*)$`)
var r_function = regexp.MustCompile(`^([\w\-]+)\s*\(\s*\)\s*\{?$`)

// Parse a line and get its type
func parse_line(line string) (DataType, []string) {
	lc := strings.TrimSpace(line)
	lr := strings.TrimRight(line, " \t\r\n")
	switch {
	case r_blank.MatchString(lc):
		return DT_BLANK, []string{}
	case r_comment.MatchString(lc):
		return DT_COMMENT, []string{lr}
	case r_variable.MatchString(lc):
		return DT_VARIABLE, r_variable.FindStringSubmatch(lc)[1:]
	case r_function.MatchString(lc):
		return DT_FUNCTION, r_function.FindStringSubmatch(lc)[1:]
	default:
		return DT_UNKNOWN, []string{lr}
	}
}

// Split a line into variables
func split_var(line string, pos int, d *delimiter) []*Data {
	dc := &Data{DT_VARIABLE, pos, ""}
	out := make([]*Data, 0)
	ig := false
	for i, c := range line {
		switch {
		case ig:
			if d.quote_closed() {
				if !is_quote(c) {
					dc.AppendRunes('\\')
				}
			} else if !d.is_quote(c) {
				dc.AppendRunes('\\')
			}
			dc.AppendRunes(c)
			ig = false
		case c == '\\':
			ig = true
		case d.quote_opened():
			if !d.set(c) {
				dc.AppendRunes(c)
			}
		case c == '#':
			if len(dc.Value) == 0 {
				dc.Type = DT_COMMENT
				dc.Append(strings.TrimRight(line[i:], " \t"))
				out = append(out, dc)
				return out
			}
			dc.AppendRunes(c)
		case is_blank(c):
			if len(dc.Value) > 0 {
				out = append(out, dc)
				dc = &Data{DT_VARIABLE, pos, ""}
			}
		case is_quote(c):
			d.set(c)
		default:
			if !d.set(c) {
				dc.AppendRunes(c)
			}
		}
	}
	if len(dc.Value) > 0 {
		out = append(out, dc)
	}
	return out
}

// Quotify if needed
func quote(s string) string {
	if s == "" {
		return s
	}
	q := "\""
	if !strings.Contains(s, "$") && !strings.Contains(s, "'") {
		q = "'"
	}
	s = strings.Replace(s, q, "\\"+q, -1)
	return q + s + q
}

func join_single(data []*Data, q bool) string {
	out := ""
	for _, d := range data {
		if q && d.Type == DT_COMMENT {
			if out != "" {
				out = quote(out)
			}
			q = false
		}
		if out != "" || d.Type == DT_COMMENT {
			out += " "
		}
		out += d.String()
	}
	if q {
		return quote(out)
	}
	return out
}

func join_list(data []*Data, q bool, multi bool, indent string) string {
	out := ""
	nl := false
	for _, d := range data {
		if nl {
			out += indent
		} else if out != "" || d.Type == DT_COMMENT {
			if multi {
				out += indent
			} else {
				out += " "
			}
		}
		nl = d.Type == DT_COMMENT
		if q && !nl {
			out += quote(d.String())
		} else {
			out += d.String()
		}
	}
	if nl {
		out += indent
	}
	return "(" + out + ")"
}

// Join data into a string
func join_data(b *Block) string {
	indent, v := "", 0
	t := UT_LINES
	if b.Type == BT_VARIABLE || b.Type == BT_UVARIABLE {
		tt, ok := U_VARIABLES[b.Name]
		if !ok {
			t = UT_OPTIONALQ
		} else {
			t = tt
		}
		indent = "\n" + strings.Repeat(" ", utf8.RuneCountInString(b.Name)+2)
		for _, d := range b.Values {
			if d.Type == DT_VARIABLE {
				v++
			}
		}
	}
	switch t {
	case UT_SINGLEVAR:
		return join_single(b.Values, false)
	case UT_SINGLEVARQ:
		return join_single(b.Values, true)
	case UT_MULTIPLEVAR:
		return join_list(b.Values, false, false, indent)
	case UT_MULTIPLEVARQ:
		return join_list(b.Values, true, false, indent)
	case UT_MULTIPLELINES:
		return join_list(b.Values, true, true, indent)
	case UT_OPTIONAL:
		if v < 2 {
			return join_single(b.Values, false)
		} else {
			return join_list(b.Values, true, false, indent)
		}
	case UT_OPTIONALQ:
		if v < 2 {
			return join_single(b.Values, true)
		} else {
			return join_list(b.Values, true, false, indent)
		}
	default:
		out := ""
		for _, d := range b.Values {
			if out != "" {
				out += "\n"
			}
			out += d.String()
		}
		return out
	}
}
