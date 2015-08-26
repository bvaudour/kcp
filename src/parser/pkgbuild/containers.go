package pkgbuild

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
)

//DataType is the type of the data.
type DataType int

const (
	DT_UNKNOWN DataType = iota
	DT_BLANK
	DT_COMMENT
	DT_VARIABLE
	DT_FUNCTION
)

//Data is an atomic element of the PKGBUILD.
type Data struct {
	Type  DataType
	Line  int
	Value string
}

//BlocType is the type of the block.
type BlockType int

const (
	BT_UNKNOWN BlockType = iota
	BT_HEADER
	BT_VARIABLE
	BT_FUNCTION
)

//Block is a group of data in the PKGBUILD.
type Block struct {
	Name     string
	Type     BlockType
	From, To int
	Values   []*Data
}

func newBlock(t BlockType, l int) *Block {
	return &Block{
		Type: t,
		From: l,
		To:   l,
	}
}

func (b *Block) add(d *Data) {
	b.Values = append(b.Values, d)
	b.To = d.Line
}

func (b *Block) str() string {
	buf := new(bytes.Buffer)
	switch b.Type {
	case BT_VARIABLE:
		buf.WriteString(b.Name)
		buf.WriteRune('=')
		buf.WriteString(joinData(b))
	case BT_FUNCTION:
		buf.WriteString(b.Name)
		buf.WriteString("() {")
		i := 0
		if len(b.Values) > 0 && b.Values[i].Type == DT_FUNCTION {
			buf.WriteByte('\n')
			buf.WriteString(b.Values[i].Value)
			buf.WriteByte('\n')
			i++
		}
		buf.WriteString("}")
		for _, v := range b.Values[i:] {
			if v.Type == DT_COMMENT {
				buf.WriteByte(' ')
			}
			buf.WriteString(v.Value)
		}
		buf.WriteByte('\n')
	default:
		for _, d := range b.Values {
			buf.WriteString(d.Value)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

//Sort engine
type blockSorter struct {
	l []*Block
	f func(*Block, *Block) bool
}

func (s *blockSorter) Len() int           { return len(s.l) }
func (s *blockSorter) Less(i, j int) bool { return s.f(s.l[i], s.l[j]) }
func (s *blockSorter) Swap(i, j int)      { s.l[i], s.l[j] = s.l[j], s.l[i] }

func sortBlocks(l []*Block, f func(*Block, *Block) bool) {
	s := &blockSorter{l, f}
	sort.Sort(s)
}

//Pkgbuild represents the parsed PKGBUILD.
type Pkgbuild struct {
	Headers   map[int]*Block
	Variables map[string][]*Block
	Functions map[string][]*Block
	Unknown   []*Block
}

func strDt(t DataType) string {
	switch t {
	case DT_BLANK:
		return "<blank>"
	case DT_COMMENT:
		return "<comment>"
	case DT_FUNCTION:
		return "<function>"
	case DT_VARIABLE:
		return "<variable>"
	}
	return "<unknown>"
}

//String is the string representation of the parsed PKGBUILD (for debug).
func (p *Pkgbuild) String() string {
	b := new(bytes.Buffer)
	sep0 := "\n#######################"
	sep1 := "\n+++++++++++++++++++++++"
	sep2 := "\n-----------------------"
	sep3 := "\n#######################"
	b.WriteString(fmt.Sprintf("Header (%d)", len(p.Headers)))
	b.WriteString(sep0)
	for _, bl := range p.Headers {
		b.WriteString(fmt.Sprintf("\n(%d-%d)", bl.From, bl.To))
		b.WriteString(sep1)
		for _, d := range bl.Values {
			b.WriteString(fmt.Sprintf("\n%s (%d)\nø%sø", strDt(d.Type), d.Line, d.Value))
			b.WriteString(sep2)
		}
	}
	b.WriteString(fmt.Sprintf("\nVariables (%d)", len(p.Variables)))
	b.WriteString(sep0)
	for n, l := range p.Variables {
		b.WriteString(fmt.Sprintf("\n%s (%d)", n, len(l)))
		b.WriteString(sep3)
		for _, bl := range l {
			b.WriteString(fmt.Sprintf("\n(%d-%d)", bl.From, bl.To))
			b.WriteString(sep1)
			for _, d := range bl.Values {
				b.WriteString(fmt.Sprintf("\n%s (%d)\nø%sø", strDt(d.Type), d.Line, d.Value))
				b.WriteString(sep2)
			}
		}
	}
	b.WriteString(fmt.Sprintf("\nFunctions (%d)", len(p.Functions)))
	b.WriteString(sep0)
	for n, l := range p.Functions {
		b.WriteString(fmt.Sprintf("\n%s (%d)", n, len(l)))
		b.WriteString(sep3)
		for _, bl := range l {
			b.WriteString(fmt.Sprintf("\n(%d-%d)", bl.From, bl.To))
			b.WriteString(sep1)
			for _, d := range bl.Values {
				b.WriteString(fmt.Sprintf("\n%s (%d)\nø%sø", strDt(d.Type), d.Line, d.Value))
				b.WriteString(sep2)
			}
		}
	}
	b.WriteString(fmt.Sprintf("\nUnkowns (%d)", len(p.Unknown)))
	b.WriteString(sep0)
	for _, bl := range p.Unknown {
		b.WriteString(fmt.Sprintf("\n(%d-%d)", bl.From, bl.To))
		b.WriteString(sep1)
		if bl.Name != "" {
			b.WriteString(fmt.Sprintf("\nø%sø", bl.Name))
			b.WriteString(sep2)
		} else {
			for _, d := range bl.Values {
				b.WriteString(fmt.Sprintf("\n%s (%d)\nø%sø", strDt(d.Type), d.Line, d.Value))
				b.WriteString(sep2)
			}
		}
	}
	return b.String()
}

//Add append a new block into the PKGBUILD.
func (p *Pkgbuild) Add(b *Block) {
	switch b.Type {
	case BT_UNKNOWN:
		p.Unknown = append(p.Unknown, b)
	case BT_HEADER:
		p.Headers[b.To] = b
	case BT_FUNCTION:
		p.Functions[b.Name] = append(p.Functions[b.Name], b)
	case BT_VARIABLE:
		p.Variables[b.Name] = append(p.Variables[b.Name], b)
	}
}

//Clean removes all uneeded comments and invalid data.
func (p *Pkgbuild) Clean() {
	p.Unknown = []*Block{}
	headers := make(map[int]*Block)
	for _, h := range p.Headers {
		if h.From == 1 {
			if len(h.Values) > 0 && h.Values[0].Type == DT_BLANK {
				h.Values = []*Data{&Data{Type: DT_BLANK}}
				headers[h.To] = h
			}
			break
		}
	}
	p.Headers = headers
	variables := make(map[string][]*Block)
	for k, l := range p.Variables {
		bl := l[len(l)-1]
		var data []*Data
		for _, v := range bl.Values {
			if v.Type == DT_VARIABLE {
				data = append(data, v)
			}
		}
		if len(data) > 0 {
			bl.Values = data
			variables[k] = []*Block{bl}
		}
	}
	p.Variables = variables
	functions := make(map[string][]*Block)
	for k, l := range p.Functions {
		bl := l[len(l)-1]
		var data []*Data
		for _, v := range bl.Values {
			if v.Type == DT_FUNCTION {
				data = append(data, v)
			}
		}
		if len(data) > 0 {
			bl.Values = data
			functions[k] = []*Block{bl}
		}
	}
	p.Functions = functions
}

//Variable returns the string represention of the elements of the needed variable.
func (p *Pkgbuild) Variable(name string) string {
	if l, ok := p.Variables[name]; ok && len(l) > 0 {
		var a []string
		for _, d := range l[0].Values {
			if d.Type == DT_VARIABLE {
				v := d.Value
				r := regexp.MustCompile(`(\$\w+|\$\{.+?\})`)
				if r.MatchString(v) {
					for _, l := range r.FindAllStringSubmatch(v, -1) {
						e := strings.Trim(l[0], "${}")
						v = strings.Replace(v, l[0], p.Variable(e), -1)
					}
				}
				a = append(a, v)
			}
		}
		return strings.Join(a, " ")
	}
	return ""
}

//Version returns the complete version of the PKGBUILD (pkgver+pkgrel).
func (p *Pkgbuild) Version() string { return p.Variable(PKGVER) + "-" + p.Variable(PKGREL) }

//Name returns the packages' name of the PKGBUILD.
func (p *Pkgbuild) Name() string {
	if l, ok := p.Variables[PKGNAME]; ok {
		for _, bl := range l {
			for _, v := range bl.Values {
				if v.Type == DT_VARIABLE {
					return v.Value
				}
			}
		}
	}
	return ""
}

//Parse reads a PKGBUILD and parses it.
func Parse(file string) (*Pkgbuild, error) {
	b, e := ioutil.ReadFile(file)
	if e != nil {
		return nil, e
	}
	rd := bytes.NewBuffer(b)
	if p, e := parse(rd); e == nil || e == io.EOF {
		return p, nil
	} else {
		return nil, e
	}
}

//ParseBytes reads PKGBUILD and parses it.
func ParseBytes(b []byte) (*Pkgbuild, error) {
	rd := bytes.NewBuffer(b)
	if p, e := parse(rd); e == nil || e == io.EOF {
		return p, nil
	} else {
		return nil, e
	}
}

//Unparse returns a formatted PKGBUILD.
func (p *Pkgbuild) Unparse(clean bool) []byte {
	if clean {
		p.Clean()
	}
	b := new(bytes.Buffer)
	mheader := make(map[int]bool)
	mvf := make(map[string]bool)
	if clean {
		for i, h := range p.Headers {
			mheader[i] = true
			b.WriteString(h.str())
		}
	}
	for _, k := range L_VARIABLES {
		l, ok := p.Variables[k]
		if !ok {
			continue
		}
		for _, v := range l {
			if h, ok := p.Headers[v.From-1]; ok && !mheader[v.From-1] {
				b.WriteString(h.str())
				mheader[h.To] = true
			}
			b.WriteString(v.str())
			b.WriteByte('\n')
		}
		mvf[k] = true
	}
	var bl []*Block
	for k, l := range p.Variables {
		if !mvf[k] {
			bl = append(bl, l...)
		}
	}
	sortBlocks(bl, func(b1, b2 *Block) bool { return b1.Name < b2.Name || (b1.Name == b2.Name && b1.From < b2.From) })
	for _, v := range bl {
		if h, ok := p.Headers[v.From-1]; ok && !mheader[v.From-1] {
			b.WriteString(h.str())
			mheader[h.To] = true
		}
		b.WriteString(v.str())
		b.WriteByte('\n')
	}
	for _, k := range L_FUNCTIONS {
		l, ok := p.Functions[k]
		if !ok {
			continue
		}
		for _, f := range l {
			if clean {
				b.WriteByte('\n')
			}
			if h, ok := p.Headers[f.From-1]; ok && !mheader[f.From-1] {
				b.WriteString(h.str())
				mheader[h.To] = true
			}
			b.WriteString(f.str())
			b.WriteByte('\n')
		}
		mvf[k] = true
	}
	bl = make([]*Block, 0, len(p.Functions))
	for k, l := range p.Functions {
		if !mvf[k] {
			bl = append(bl, l...)
		}
	}
	sortBlocks(bl, func(b1, b2 *Block) bool { return b1.From < b2.From })
	for _, f := range bl {
		if clean {
			b.WriteByte('\n')
		}
		if h, ok := p.Headers[f.From-1]; ok {
			b.WriteString(h.str())
			mheader[h.To] = true
		}
		b.WriteString(f.str())
		b.WriteByte('\n')
	}
	if len(p.Unknown) > 0 {
		b.WriteByte('\n')
	}
	for _, u := range p.Unknown {
		if h, ok := p.Headers[u.From-1]; ok {
			b.WriteString(h.str())
			mheader[h.To] = true
		}
		b.WriteString(u.str())
	}
	bl = []*Block{}
	for i, h := range p.Headers {
		if !mheader[i] {
			bl = append(bl, h)
		}
	}
	if len(bl) > 0 {
		sortBlocks(bl, func(b1, b2 *Block) bool { return b1.From < b2.From })
		bh := new(bytes.Buffer)
		for _, h := range bl {
			bh.WriteString(h.str())
		}
		b.WriteTo(bh)
		b = bh
	}
	return b.Bytes()
}

//Version extracts the complete version (pkgver+pkgrel) of a PKGBUILD without parsing completely the file.
func Version(b []byte) (string, bool) {
	pkgver, pkgrel := "", ""
	const v, r = "pkgver=", "pkgrel="
loop:
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, v):
			pkgver = strings.TrimSpace(l[len(v):])
		case strings.HasPrefix(l, r):
			pkgrel = strings.TrimSpace(l[len(r):])
		default:
			continue loop
		}
		if pkgver != "" && pkgrel != "" {
			return pkgver + "-" + pkgrel, true
		}
	}
	return "", false
}
