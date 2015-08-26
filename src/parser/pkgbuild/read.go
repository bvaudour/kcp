package pkgbuild

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

func read(rd io.RuneReader) (r rune, e error) {
	r, _, e = rd.ReadRune()
	return
}

func readAll(rd io.RuneReader, wr *bytes.Buffer, contains bool, rg string) (r rune, e error, dl int) {
	r, e = read(rd)
	for e == nil && strings.ContainsRune(rg, r) == contains {
		if r == '\n' {
			dl++
		}
		wr.WriteRune(r)
		r, e = read(rd)
	}
	return
}

func readLine(rd io.RuneReader, wr *bytes.Buffer) (r rune, e error) {
	r, e, _ = readAll(rd, wr, false, "\n")
	return
}

func readBlank(rd io.RuneReader, wr *bytes.Buffer) (r rune, e error) {
	r, e, _ = readAll(rd, wr, true, " \t")
	return
}

func readId(rd io.RuneReader, wr *bytes.Buffer) (r rune, e error) {
	r, e, _ = readAll(rd, wr, false, " \t\n=(")
	return
}

func readFunction(rd io.RuneReader, wr *bytes.Buffer) (r rune, e error, dl int) {
	d, q, ig := 1, rune(0), false
	r, e = read(rd)
loop:
	for e == nil {
		if r == '\n' {
			dl++
		}
		if ig {
			ig = false
		} else {
			switch r {
			case '\\':
				ig = true
			case '\'', '"':
				if q == 0 {
					q = r
				} else if q == r {
					q = 0
				}
			case '{':
				if q == 0 {
					d++
				}
			case '}':
				if q == 0 {
					d--
					break loop
				}
			}
		}
		wr.WriteRune(r)
		r, e = read(rd)
	}
	return
}

func readVariable(rd io.RuneReader, wr *bytes.Buffer, q rune, ig bool) (r rune, e error, dl int) {
	r, e = read(rd)
loop:
	for e == nil {
		add := true
		if ig {
			ig = false
			switch r {
			case '\'':
				if q == '"' {
					wr.WriteRune('\\')
				}
			case '"':
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case '$':
				wr.WriteRune('\\')
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case ' ', '\t', '\n', '#', '(', ')', '[', ']', '{', '}':
				if q != 0 {
					wr.WriteRune('\\')
				}
			}
		} else {
			switch r {
			case '\\':
				ig, add = true, false
			case '"', '\'':
				if q == 0 {
					q, add = r, false
				} else if q == r {
					q, add = 0, false
				}
			case '$':
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case ' ', '\t', '\n', ')':
				if q == 0 {
					break loop
				}
			}
			if add {
				wr.WriteRune(r)
			}
			if r == '\n' {
				dl++
			}
			r, e = read(rd)
		}
	}
	return
}

func dataVariable(rd io.RuneReader, wr *bytes.Buffer, line int, q rune, ig bool) (r rune, e error, data *Data, nl int) {
	r, e = read(rd)
	nl = line
loop:
	for e == nil {
		add := true
		if ig {
			ig = false
			switch r {
			case '\'':
				if q == '"' {
					wr.WriteRune('\\')
				}
			case '"':
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case '$':
				wr.WriteRune('\\')
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case '\n', ' ', '\t', '#', '[', ']', '{', '}', '(', ')':
				if q != 0 {
					wr.WriteRune('\\')
				}
			}
		} else {
			switch r {
			case '\'', '"':
				if q == 0 {
					q, add = r, false
				} else if q == r {
					q, add = 0, false
				}
			case '\\':
				ig, add = true, false
			case '$':
				if q == '\'' {
					wr.WriteRune('\\')
				}
			case '\n', ' ', '\t', ')':
				if q == 0 {
					break loop
				}
			}
		}
		if add {
			wr.WriteRune(r)
		}
		if r == '\n' {
			nl++
		}
		r, e = read(rd)
	}
	data = &Data{
		Type:  DT_COMMENT,
		Line:  line,
		Value: wr.String(),
	}
	return
}

func parse(rd io.RuneReader) (p *Pkgbuild, e error) {
	p = &Pkgbuild{
		Headers:   make(map[int]*Block),
		Variables: make(map[string][]*Block),
		Functions: make(map[string][]*Block),
	}
	var block *Block
	var last rune
	l := 1
loop:
	for e == nil {
		blank := new(bytes.Buffer)
		last, e = readBlank(rd, blank)
		comment := false
		if last == '#' {
			comment = true
			blank.Reset()
			blank.WriteRune(last)
			last, e = readLine(rd, blank)
		}
		if last == '\n' {
			if block == nil {
				block = newBlock(BT_HEADER, l)
			}
			data := new(Data)
			data.Line, data.Value = l, blank.String()
			if comment {
				data.Type = DT_COMMENT
			} else {
				data.Type = DT_BLANK
			}
			block.add(data)
			l++
			continue
		}
		if e != nil {
			break
		}
		if block != nil {
			p.Add(block)
		}
		block = newBlock(BT_UNKNOWN, l)
		id := new(bytes.Buffer)
		id.WriteRune(last)
		last, e = readId(rd, id)
		block.Name = id.String()
		if e == nil {
			switch last {
			case '(':
				id.WriteRune(last)
				last, e = read(rd)
				if last == ')' {
					id.WriteRune(last)
					dl := 0
					last, e, dl = readAll(rd, id, true, " \t\n")
					l += dl
					if last == '{' {
						id.WriteRune(last)
						v := new(bytes.Buffer)
						last, e, dl = readFunction(rd, v)
						ol := l
						l += dl
						if last == '}' {
							data := &Data{
								Type:  DT_FUNCTION,
								Line:  ol,
								Value: v.String(),
							}
							data.Value = strings.Trim(data.Value, "\n")
							block.add(data)
							block.Type, block.To = BT_FUNCTION, l
							blank.Reset()
							last, e = readBlank(rd, blank)
							if e == nil && last != '\n' {
								data := new(Data)
								if last == '#' {
									blank.Reset()
									data.Type = DT_COMMENT
								} else {
									data.Type = DT_UNKNOWN
								}
								last, e = readLine(rd, blank)
								data.Line, data.Value = l, blank.String()
								block.add(data)
							}
							p.Add(block)
							block = nil
							l++
							continue loop
						}
						v.WriteTo(id)
					}
				}
			case '=':
				block.Type = BT_VARIABLE
				blank.Reset()
				last, e = readBlank(rd, blank)
				if blank.Len() > 0 && e == nil && last != '\n' {
					data := new(Data)
					if last == '#' {
						data.Type = DT_COMMENT
						blank.Reset()
						last, e = readLine(rd, blank)
					} else {
						data.Type = DT_UNKNOWN
					}
					data.Line, data.Value = l, blank.String()
					block.add(data)
				}
				if last != '\n' && e == nil {
					mult := false
					if last == '(' {
						mult = true
						last, e = read(rd)
					}
					if e == nil {
						q, ig, v, dl := rune(0), false, new(bytes.Buffer), 0
						switch last {
						case '\'', '"':
							q = last
						case '\\':
							ig = true
						default:
							v.WriteRune(last)
						}
						last, e, dl = readVariable(rd, v, q, ig)
						data := &Data{
							Type:  DT_VARIABLE,
							Line:  l,
							Value: v.String(),
						}
						block.add(data)
						l += dl
						block.To = l
						if mult {
							for e == nil && last != ')' {
								last, e, dl = readAll(rd, new(bytes.Buffer), true, " \t\n")
								l += dl
								if e != nil || last == ')' {
									break
								}
								v, data = new(bytes.Buffer), new(Data)
								if last == '#' {
									last, e = readLine(rd, v)
									data.Type = DT_COMMENT
								} else {
									data.Type = DT_VARIABLE
									q, ig = 0, false
									switch last {
									case '\'', '"':
										q = last
									case '\\':
										ig = true
									default:
										v.WriteRune(last)
									}
									last, e, dl = readVariable(rd, v, q, ig)
									l += dl
								}
								data.Line, data.Value = l, v.String()
								block.add(data)
								block.To = l
								if last == '\n' {
									l++
								}
							}
							last, e = read(rd)
						}
					}
				}
				if last != '\n' && e == nil {
					data := new(Data)
					if last == '#' {
						blank.Reset()
						data.Type = DT_COMMENT
					} else {
						data.Type = DT_UNKNOWN
					}
					last, e = readLine(rd, blank)
					data.Line, data.Value = l, blank.String()
					block.add(data)
				}
				p.Add(block)
				block = nil
				l++
				continue loop
			}
		}
		if e == nil && last != '\n' {
			last, e = readLine(rd, id)
			block.Name, block.To = id.String(), l
		}
		if len(block.Name) > 0 {
			p.Add(block)
		}
		block = nil
		l++
	}
	if block != nil {
		p.Add(block)
	}
	return
}

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

func joinSingle(data []*Data, q bool) string {
	b := new(bytes.Buffer)
	for _, d := range data {
		if d.Type == DT_UNKNOWN || d.Type == DT_BLANK {
			continue
		}
		if q && d.Type == DT_COMMENT {
			if b.Len() > 0 {
				b = bytes.NewBufferString(quote(b.String()))
			}
			q = false
		}
		if b.Len() > 0 || d.Type == DT_COMMENT {
			b.WriteByte(' ')
		}
		b.WriteString(d.Value)
	}
	if q {
		return quote(b.String())
	}
	return b.String()
}

func joinList(data []*Data, q bool, multi bool, indent string) string {
	b := new(bytes.Buffer)
	nl := false
	for _, d := range data {
		if d.Type == DT_UNKNOWN || d.Type == DT_BLANK {
			continue
		}
		if nl {
			b.WriteString(indent)
		} else if b.Len() > 0 || d.Type == DT_COMMENT {
			if multi {
				b.WriteString(indent)
			} else {
				b.WriteByte(' ')
			}
		}
		nl = d.Type == DT_COMMENT
		if q && !nl {
			b.WriteString(quote(d.Value))
		} else {
			b.WriteString(d.Value)
		}
	}
	if nl {
		b.WriteString(indent)
	}
	return "(" + b.String() + ")"
}

func joinData(b *Block) string {
	indent, v, t := "", 0, 0

	if tt, ok := uVariables[b.Name]; ok {
		t = tt
	} else {
		t = uOptionalQ
	}
	indent = "\n" + strings.Repeat(" ", utf8.RuneCountInString(b.Name)+2)
	for _, d := range b.Values {
		if d.Type == DT_VARIABLE {
			v++
		}
	}
	switch t {
	case uSingleVar:
		return joinSingle(b.Values, false)
	case uSingleVarQ:
		return joinSingle(b.Values, true)
	case uMultipleVar:
		return joinList(b.Values, false, false, indent)
	case uMultipleVarQ:
		return joinList(b.Values, true, false, indent)
	case uMultipleLines:
		return joinList(b.Values, true, true, indent)
	case uOptional:
		if v < 2 {
			return joinSingle(b.Values, false)
		} else {
			return joinList(b.Values, true, false, indent)
		}
	case uOptionalQ:
		if v < 2 {
			return joinSingle(b.Values, true)
		} else {
			return joinList(b.Values, true, false, indent)
		}
	}
	buf := new(bytes.Buffer)
	for i, d := range b.Values {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(d.Value)
	}
	return buf.String()
}
