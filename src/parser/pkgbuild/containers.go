package pkgbuild

import (
	"fmt"
	"strings"
)

type Data struct {
	Type  DataType
	Line  int
	Value string
}

func (d *Data) Append(s string) {
	d.Value += s
}

func (d *Data) AppendRunes(r ...rune) {
	d.Append(string(r))
}

func (d *Data) String() string {
	return d.Value
}

type Block struct {
	Name       string
	Type       BlockType
	Begin, End int
	Header     []*Data
	Values     []*Data
}

func NewBlock(n string, b int, t BlockType) *Block {
	out := new(Block)
	out.Name = n
	out.Begin = b
	out.Type = t
	out.Header = []*Data{}
	out.Values = []*Data{}
	return out
}

func emptyBlock(begin int) *Block {
	return NewBlock("", begin, BT_NONE)
}

func initb(begin int, line string) *Block {
	b, d := emptyBlock(begin), initd()
	t, v := parse_line(line)
	switch t {
	case DT_UNKNOWN:
		b.Name, b.Type = UNKNOWN, BT_UNKNOWN
		b.AppendDataString(line, t, begin)
	case DT_VARIABLE:
		b.Name, b.Type = v[0], BT_UVARIABLE
		b.update_type()
		b.AppendData(split_var(v[1], begin, d)...)
	case DT_FUNCTION:
		b.Name, b.Type = v[0], BT_UFUNCTION
		b.update_type()
	default:
		if begin == 1 {
			b.Name, b.Type = HEADER, BT_HEADER
		}
		b.AppendHeaderString(line, t, begin)
	}
	return b
}

func (b *Block) AppendHeader(d ...*Data) {
	b.Header = append(b.Header, d...)
}

func (b *Block) AppendHeaderString(s string, t DataType, l int) {
	b.AppendHeader(&Data{t, l, s})
}

func (b *Block) AppendData(d ...*Data) {
	b.Values = append(b.Values, d...)
}

func (b *Block) AppendDataString(s string, t DataType, l int) {
	b.AppendData(&Data{t, l, s})
}

func (b *Block) FusionData(d ...*Data) {
	if len(d) > 0 {
		if l := len(b.Values) - 1; l >= 0 {
			b.Values[l].Append("\n" + d[0].String())
			b.AppendData(d[1:]...)
			return
		}
	}
	b.AppendData(d...)
}

func (b *Block) update_type() {
	a, t := L_VARIABLES, BT_VARIABLE
	if b.Type == BT_UFUNCTION {
		a, t = L_FUNCTIONS, BT_UFUNCTION
	}
	for _, e := range a {
		if b.Name == e {
			b.Type = t
			return
		}
	}
}

func (b *Block) s_name(unknown bool) string {
	n := b.Name
	if b.Type == BT_FUNCTION || b.Type == BT_UFUNCTION {
		n += "()"
	}
	out := fmt.Sprintf("\033[1;31m%s (%d-%d)\033[m", n, b.Begin, b.End)
	if unknown && b.Type != BT_UNKNOWN {
		out += " [unknown]"
	}
	return out
}

func (b *Block) s_header() string {
	out := ""
	for _, d := range b.Header {
		out = fmt.Sprintf("%s\n%s", out, d.String())
	}
	return out
}

func (b *Block) s_values() string {
	out := ""
	for _, d := range b.Values {
		out = fmt.Sprintf("%s\n  - '%s'", out, d.String())
		if d.Type == DT_COMMENT {
			out += " [comment]"
		}
	}
	return out
}

func (b *Block) stringer(unknown bool) string {
	out := b.s_name(unknown)
	out += "\n------------------------------------"
	out += b.s_header()
	out += "\n------------------------------------"
	out += b.s_values()
	out += "\n------------------------------------"
	return out
}

func (b *Block) String() string {
	return b.stringer(false)
}

func (b *Block) l_header() []string {
	out := make([]string, 0, len(b.Header))
	if b.Type == BT_HEADER {
		if len(b.Header) > 0 && b.Header[0].Type == DT_BLANK {
			out = append(out, "")
		}
	} else if b.Type != BT_VARIABLE && b.Type != BT_UVARIABLE {
		out = append(out, "")
	}
	for _, d := range b.Header {
		if d.Type != DT_BLANK {
			out = append(out, d.String())
		}
	}
	return out
}

func (b *Block) l_values() []string {
	out := make([]string, 0, len(b.Values))
	switch b.Type {
	case BT_VARIABLE:
		fallthrough
	case BT_UVARIABLE:
		out = append(out, fmt.Sprintf("%s=%s", b.Name, join_data(b)))
	case BT_FUNCTION:
		fallthrough
	case BT_UFUNCTION:
		out = append(out, fmt.Sprintf("%s() {", b.Name))
		d := join_data(b)
		if d != "" {
			out = append(out, d)
		}
		out = append(out, "}")
	default:
		d := join_data(b)
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}

func (b *Block) Lines() []string {
	out := b.l_header()
	out = append(out, b.l_values()...)
	return out
}

type Pkgbuild struct {
	Header    *Block
	Variables map[string]*Block
	Functions map[string]*Block
	Unknowns  []*Block
}

func NewPkgbuild() *Pkgbuild {
	p := new(Pkgbuild)
	p.Variables = make(map[string]*Block)
	p.Functions = make(map[string]*Block)
	p.Unknowns = make([]*Block, 0)
	return p
}

func (p *Pkgbuild) Append(b *Block) {
	switch b.Type {
	case BT_HEADER:
		p.Header = b
	case BT_VARIABLE:
		fallthrough
	case BT_UVARIABLE:
		if _, ok := p.Variables[b.Name]; ok {
			p.Unknowns = append(p.Unknowns, b)
		} else {
			p.Variables[b.Name] = b
		}
	case BT_FUNCTION:
		fallthrough
	case BT_UFUNCTION:
		if _, ok := p.Functions[b.Name]; ok {
			p.Unknowns = append(p.Unknowns, b)
		} else {
			p.Functions[b.Name] = b
		}
	default:
		p.Unknowns = append(p.Unknowns, b)
	}
}

func (p *Pkgbuild) String() string {
	out := ""
	if p.Header != nil {
		out = p.Header.String()
	}
	for _, b := range p.Variables {
		if out != "" {
			out += "\n"
		}
		out += b.String()
	}
	for _, b := range p.Functions {
		if out != "" {
			out += "\n"
		}
		out += b.String()
	}
	for _, b := range p.Unknowns {
		if out != "" {
			out += "\n"
		}
		out += b.stringer(true)
	}
	return out
}

func add_lines(mb map[string]*Block, key string, lines *[]string) {
	if b, ok := mb[key]; ok {
		*lines = append(*lines, b.Lines()...)
	}
}
func add_ulines(mb map[string]*Block, tpe BlockType, lines *[]string) {
	for _, b := range mb {
		if b.Type == tpe {
			*lines = append(*lines, b.Lines()...)
		}
	}
}

func (p *Pkgbuild) Lines() []string {
	out := make([]string, 0)
	if p.Header != nil {
		out = p.Header.Lines()
	}
	for _, k := range L_VARIABLES {
		add_lines(p.Variables, k, &out)
	}
	add_ulines(p.Variables, BT_UVARIABLE, &out)
	for _, k := range L_FUNCTIONS {
		add_lines(p.Functions, k, &out)
	}
	add_ulines(p.Functions, BT_UFUNCTION, &out)
	for _, b := range p.Unknowns {
		out = append(out, b.Lines()...)
	}
	return out
}

func append_data(d *[]*Data, l string, pos int) {
	*d = append(*d, &Data{DT_UNKNOWN, pos, l})
}

func init_data(d *[]*Data, l string, pos int) {
	*d = make([]*Data, 0, 1)
	append_data(d, l, pos)
}

func (p *Pkgbuild) Parse(lines []string) {
	var bc *Block
	dc := make([]*Data, 0)
	d := initd()
	begin := false
	for i, l := range lines {
		lc := strings.TrimRight(l, " \t\r\n")
		switch {
		case bc == nil:
			bc = initb(i+1, lc)
			d.update(lc)
			if bc.Type != BT_HEADER && bc.Type != BT_NONE {
				init_data(&dc, lc, bc.Begin)
				if bc.Type == BT_FUNCTION || bc.Type == BT_UFUNCTION {
					if !strings.HasSuffix(lc, "{") {
						begin = true
					}
				}
			}
		case bc.Type == BT_VARIABLE || bc.Type == BT_UVARIABLE:
			append_data(&dc, lc, i+1)
			if d.quote_opened() {
				bc.FusionData(split_var(lc, i+1, d)...)
			} else {
				bc.AppendData(split_var(lc, i+1, d)...)
			}
		case bc.Type == BT_FUNCTION || bc.Type == BT_UFUNCTION:
			append_data(&dc, lc, i+1)
			d.update(lc)
			switch {
			case begin:
				switch strings.TrimSpace(lc) {
				case "":
				case "{":
					begin = false
				default:
					bc.Type = BT_UNKNOWN
					bc.Values = dc
					dc = make([]*Data, 0)
					begin = false
				}
			case d.closed():
				if strings.TrimSpace(lc) != "}" {
					bc.Type = BT_UNKNOWN
					bc.Values = dc
					dc = make([]*Data, 0)
				}
			default:
				bc.AppendDataString(lc, DT_FUNCTION, i+1)
			}
		case bc.Type == BT_UNKNOWN:
			append_data(&dc, lc, i+1)
			d.update(lc)
			bc.AppendDataString(lc, DT_UNKNOWN, i+1)
		default:
			bcn := initb(i+1, lc)
			d.update(lc)
			switch {
			case bcn.Type == BT_NONE:
				bc.AppendHeader(bcn.Header...)
			case bc.Type == BT_HEADER:
				p.Append(bc)
				bc = bcn
			default:
				init_data(&dc, lc, bcn.Begin)
				bcn.AppendHeader(bc.Header...)
				bc = bcn
				if bc.Type == BT_FUNCTION || bc.Type == BT_UFUNCTION {
					if !strings.HasSuffix(lc, "{") {
						begin = true
					}
				}
			}
		}
		bc.End = i + 1
		if !begin && d.closed() && bc.Type != BT_HEADER && bc.Type != BT_NONE {
			p.Append(bc)
			bc = nil
		}
	}
	/*
		if bc != nil {
			if bc.Type != BT_HEADER {
				bc.Type = BT_UNKNOWN
			}
			bc.Values = dc
			p.Append(bc)
		}
	*/
}
