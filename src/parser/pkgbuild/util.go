package pkgbuild

import (
	"io/ioutil"
	"strings"
)

func lineType(l string) (int, int, []string) {
	switch {
	case R_BLANK.MatchString(l):
		return TD_BLANK, 0, []string{strings.TrimSpace(l)}
	case R_COMMENT.MatchString(l):
		return TD_COMMENT, 0, []string{strings.TrimSpace(l)}
	case R_FUNCTION.MatchString(l):
		return TD_FUNC, strings.Count(l, "{") - strings.Count(l, "}"), []string{R_FUNCTION.FindStringSubmatch(l)[0], strings.TrimSpace(l)}
	case R_MVAR1.MatchString(l):
		return TD_VARIABLE, 0, R_MVAR1.FindStringSubmatch(l)
	case R_MVAR2.MatchString(l):
		return TD_VARIABLE, 1, R_MVAR2.FindStringSubmatch(l)
	case R_MVAR3.MatchString(l):
		return TD_VARIABLE, 0, R_MVAR3.FindStringSubmatch(l)
	default:
		return TD_UNKNOWN, 0, []string{strings.TrimRight(l, "\t ")}
	}
}

func splitString(s string) []string {
	s = strings.Trim(s, "() ")
	out := make([]string, 0)
	q, sc, ign := "", make([]rune, 0), false
	for _, c := range s {
		switch {
		case c == '\\':
			if ign {
				sc = append(sc, c, c)
			} else {
				ign = false
			}
		case c == ' ':
			switch {
			case q != "":
				if ign {
					sc = append(sc, '\\')
				}
				sc = append(sc, c)
			case ign:
				sc = append(sc, c)
			case len(sc) > 0:
				out = append(out, string(sc))
				sc = make([]rune, 0)
			}
		case c == '\'' || c == '"':
			switch {
			case ign:
				sc = append(sc, c)
			case q == "":
				if len(sc) > 0 {
					out = append(out, string(sc))
					sc = make([]rune, 0)
				}
				q = string(c)
			case q != string(c):
				sc = append(sc, c)
			case len(sc) > 0:
				out = append(out, string(sc))
				sc = make([]rune, 0)
			}
		default:
			if ign {
				sc = append(sc, '\\')
			}
			sc = append(sc, c)
		}
		if ign && c != '\\' {
			ign = false
		}
	}
	return out
}

func joinData(data []*Data, spl, q bool) string {
	s := ""
	for _, d := range data {
		sc := ""
		if q {
			sc = d.Quote()
		} else {
			sc = d.String()
		}
		if spl && s != "" && sc != "" {
			s += " "
		}
		s += sc
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
	} else {
		q = "'"
	}
	s = strings.Replace(s, q, "\\"+q, -1)
	return q + s + q
}

func getLinesByKey(p Pkgbuild, order map[*Container]*Container, key string, lines *[]string) {
	if c, ok := p[key]; ok {
		for len(c) > 0 {
			if c[0].Type == TC_FUNCTION {
				if prev, ok := order[c[0]]; ok && prev.Type == TC_BLANKCOMMENT {
					*lines = append(*lines, prev.Lines()...)
				}
				if l := len(*lines); l > 0 && (*lines)[l-1] != "" {
					*lines = append(*lines, "")
				}
			}
			*lines = append(*lines, c[0].Lines()...)
			c = c[1:]
		}
		delete(p, key)
	}
}

func getLinesByType(p Pkgbuild, order map[*Container]*Container, key int, lines *[]string) {
	c, idx := make(LContainer, 0), make(map[*Container]int)
	for _, cont := range p {
		for i, cc := range cont {
			if cc.Type == key {
				c = append(c, cc)
				idx[cc] = i
			}
		}
	}
	c.Sort()
	checkPrev := key == TC_SFUNCTION || key == TC_UFUNCTION || key == TC_UNKNOWN
	for _, e := range c {
		if checkPrev {
			if prev, ok := order[c[0]]; ok && prev.Type == TC_BLANKCOMMENT {
				*lines = append(*lines, prev.Lines()...)
			}
			if l := len(*lines); l > 0 && (*lines)[l-1] != "" {
				*lines = append(*lines, "")
			}
		}
		*lines = append(*lines, e.Lines()...)
		p.Remove(e.Name, idx[e])
		if len(p[e.Name]) == 0 {
			delete(p, e.Name)
		}
	}
}

func readFile(path string) (lines []string, err error) {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		err = e
		lines = []string{}
	} else {
		lines = strings.Split(string(b), "\n")
	}
	return
}

func writeFile(path string, lines []string) error {
	b := append([]byte(strings.Join(lines, "\n")), '\n')
	return ioutil.WriteFile(path, b, 0644)
}
