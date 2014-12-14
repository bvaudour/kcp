package pargs

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type Parser struct {
	p        properties
	flags    []*Flag
	args     map[string]*Flag
	groups   map[*Group][]*Flag
	fgroups  map[*Flag][]*Group
	requires map[*Flag][]*Flag
	pre      []string
	post     []string
}

// Properties management
func (p *Parser) Set(k int, v interface{}) error { return p.p.set(k, v) }
func (p *Parser) GetString(k int) string         { return p.p.vstring(k) }
func (p *Parser) GetBool(k int) bool             { return p.p.vbool(k) }
func (p *Parser) Name() string                   { return p.GetString(NAME) }
func (p *Parser) Description() string            { return p.GetString(DESCRIPTION) }
func (p *Parser) LongDescription() string        { return p.GetString(LONGDESCRIPTION) }
func (p *Parser) Synopsis() string               { return p.GetString(SYNOPSIS) }
func (p *Parser) Author() string                 { return p.GetString(AUTHOR) }
func (p *Parser) Version() string                { return p.GetString(VERSION) }
func (p *Parser) AllowPreArgs() bool             { return p.GetBool(ALLOWPREARGS) }
func (p *Parser) AllowPostArgs() bool            { return p.GetBool(ALLOWPOSTARGS) }

// Flag management
func (p *Parser) GetFlag(arg string) *Flag {
	f, _ := p.args[arg]
	return f
}
func (p *Parser) ContainsFlag(arg string) bool {
	_, ok := p.args[arg]
	return ok
}
func (p *Parser) Add(f *Flag) error {
	lf, sf := f.Long(), f.Short()
	if p.ContainsFlag(lf) {
		return unexpectedFlag(lf)
	} else if p.ContainsFlag(sf) {
		return unexpectedFlag(sf)
	}
	p.flags = append(p.flags, f)
	p.fgroups[f] = make([]*Group, 0)
	if lf != "" {
		p.args[lf] = f
	}
	if sf != "" {
		p.args[sf] = f
	}
	return nil
}
func (p *Parser) AddAll(flags ...*Flag) []error {
	out := make([]error, 0)
	for _, f := range flags {
		if e := p.Add(f); e != nil {
			out = append(out, e)
		}
	}
	return out
}

// Parser Builder
func New(d, v string) *Parser {
	p := new(Parser)
	p.p = parserProps()
	p.flags = make([]*Flag, 0)
	p.args = make(map[string]*Flag)
	p.groups = make(map[*Group][]*Flag)
	p.fgroups = make(map[*Flag][]*Group)
	p.requires = make(map[*Flag][]*Flag)
	p.Set(DESCRIPTION, d)
	p.Set(VERSION, v)
	return p
}
func (p *Parser) Bool(sf, lf, d string) (*bool, error) {
	f, o, e := NewBoolFlag(sf, lf, d)
	if e == nil {
		e = p.Add(f)
	}
	return o, e
}
func (p *Parser) String(sf, lf, d, vn, dv string) (*string, error) {
	f, o, e := NewStringFlag(sf, lf, d, vn, dv)
	if e == nil {
		e = p.Add(f)
	}
	return o, e
}
func (p *Parser) Choice(sf, lf, d, dv string, c []string) (*string, error) {
	f, o, e := NewChoiceFlag(sf, lf, d, dv, c)
	if e == nil {
		e = p.Add(f)
	}
	return o, e
}
func (p *Parser) Array(sf, lf, d, vn string) (*[]string, error) {
	f, o, e := NewArrayFlag(sf, lf, d, vn)
	if e == nil {
		e = p.Add(f)
	}
	return o, e
}
func (p *Parser) Int(sf, lf, d, vn string, dv int) (*int, error) {
	f, o, e := NewIntFlag(sf, lf, d, vn, dv)
	if e == nil {
		e = p.Add(f)
	}
	return o, e
}

// Grouping/Requiring flags
func (p *Parser) Group(args ...string) int {
	g := new(Group)
	flags := make([]*Flag, 0)
	for _, a := range args {
		f := p.GetFlag(a)
		if f == nil {
			continue
		}
		flags = append(flags, f)
		p.fgroups[f] = append(p.fgroups[f], g)
	}
	p.groups[g] = flags
	return len(flags)
}
func (p *Parser) Require(arg0 string, args ...string) int {
	f0 := p.GetFlag(arg0)
	if f0 == nil {
		return 0
	}
	flags := make([]*Flag, 0)
	for _, a := range args {
		f := p.GetFlag(a)
		if f != nil {
			flags = append(flags, f)
		}
	}
	p.requires[f0] = flags
	return len(flags)
}

// Parsing
func format(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		a = strings.TrimSpace(a)
		a = strings.Trim(a, "=")
		l := len(a)
		switch {
		case l == 0:
			continue
		case l > 2 && a[0] == '-':
			if a[1] == '-' {
				i := strings.Index(a, "=")
				if i < 0 {
					out = append(out, a)
				} else {
					out = append(out, a[:i], a[i+1:])
				}
			} else {
				for _, c := range a[1:] {
					out = append(out, fmt.Sprintf("-%c", c))
				}
			}
		default:
			out = append(out, a)
		}
	}
	return out
}
func (p *Parser) Parse(args []string) error {
	if len(args) == 0 {
		return notEnoughArg()
	}
	_, app := path.Split(args[0])
	p.Set(NAME, app)
	args = format(args[1:])
	if contains(args, "--create-manpage") {
		p.PrintMan()
		os.Exit(0)
	}
	if len(args) == 0 {
		return nil
	}

	// Get Flags
	l := len(args)
	idxF, flags := make([]int, 0, l), make([]*Flag, 0, l)
	for i, a := range args {
		switch {
		case a[0] != '-':
			continue
		case !p.ContainsFlag(a):
			return unsupportedFlag(a)
		default:
			f := p.GetFlag(a)
			f.used = a
			flags = append(flags, f)
			idxF = append(idxF, i)
		}
	}

	// Extract Pre-args
	switch {
	case len(idxF) == 0:
		switch {
		case p.AllowPreArgs():
			p.pre = args
		case p.AllowPostArgs():
			p.post = args
		default:
			return notAllowed("Pre-args")
		}
	case p.AllowPreArgs():
		p.pre = args[:idxF[0]]
	case idxF[0] == 0:
	default:
		return notAllowed("Pre-args")
	}

	// Parse args
	l = len(idxF) - 1
	for i, f := range flags {
		a := args[idxF[i]]
		if grps, ok := p.fgroups[f]; ok {
			for _, g := range grps {
				if g.selected != "" && g.selected != f.Short() && g.selected != f.Long() {
					return mustBeAlternative(a, g.selected)
				}
				g.selected = a
			}
		}
		i1 := idxF[i] + 1
		i2 := len(args)
		if i != l {
			i2 = idxF[i+1]
		}
		switch {
		case i1 == i2:
			if e := f.f(f.DefaultValue()); e != nil {
				return e
			}
		case i1+1 == i2:
			if e := f.f(args[i1]); e != nil {
				return e
			}
		case f.AllowMultipleValues():
			for _, an := range args[i1:i2] {
				if e := f.f(an); e != nil {
					return e
				}
			}
		case i != l:
			return noMultipleAllowed(a)
		default:
			if e := f.f(args[i1]); e != nil {
				return e
			}
			if p.AllowPostArgs() {
				p.post = args[i1+1:]
			} else {
				return notAllowed("Post-args")
			}
		}
	}

	// Check Requirements
	for f0, req := range p.requires {
		if f0.used == "" {
			continue
		}
		ok := false
		str := make([]string, 0, len(req)*2)
		for _, r := range req {
			if r.used != "" {
				ok = true
				break
			}
			str = append(str, r.Short(), r.Long())
		}
		if !ok {
			return needRequirment(f0.used, str)
		}
	}
	return nil
}
func (p *Parser) GetPreArgs() []string {
	if p.pre != nil {
		return p.pre
	}
	return make([]string, 0)
}
func (p *Parser) GetPostArgs() []string {
	if p.post != nil {
		return p.post
	}
	return make([]string, 0)
}

// Man Print
func (f *Flag) man() string {
	h := new(bytes.Buffer)
	sf, lf := f.Short(), f.Long()
	switch {
	case sf == "":
		fmt.Fprintf(h, "\\-\\-%s", lf[2:])
	case lf == "":
		fmt.Fprintf(h, "\\-%s", sf[1:])
	default:
		fmt.Fprintf(h, "\\-%s|\\-\\-%s", sf[1:], lf[2:])
	}
	if vn := f.ValueName(); vn != "" {
		if dv := f.DefaultValue(); dv == "" {
			fmt.Fprintf(h, " [%s]", vn)
		} else {
			fmt.Fprintf(h, " %s", vn)
		}
	}
	return h.String()
}
func (p *Parser) synopsis() string {
	h := new(bytes.Buffer)
	if p.AllowPreArgs() {
		fmt.Fprint(h, " [BARGS...]")
	}
	for _, f := range p.flags {
		fmt.Fprint(h, " [%s]", f.man())
	}
	if p.AllowPostArgs() {
		fmt.Fprint(h, " [EARGS...]")
	}
	return h.String()
}
func (p *Parser) description() string {
	h := new(bytes.Buffer)
	for _, l := range strings.Split(p.LongDescription(), "\n") {
		if l == "" {
			fmt.Fprintln(h, ".PP")
		} else {
			fmt.Fprintln(h, l)
		}
	}
	return h.String()
}
func (p *Parser) PrintMan() {
	app, version, summ, author := p.Name(), p.Version(), p.Description(), p.Author()
	fmt.Printf(".TH \"%s\" 1 \"%s\" \"%s\" \"\"\n", app, time.Now().Format("January 2, 2006"), version)
	fmt.Println(".SH NAME")
	fmt.Println(app, "\\-", summ)
	fmt.Println(".SH SYNOPSIS")
	fmt.Println(app, p.synopsis())
	fmt.Println(".SH DESCRIPTION")
	fmt.Println(p.description())
	fmt.Println(".SH OPTIONS")
	for _, f := range p.flags {
		if f.Hidden() {
			continue
		}
		fmt.Println(".TP")
		fmt.Println(f.man())
		fmt.Printf("\t%s\n", f.Description())
	}
	if author != "" {
		fmt.Printf(".SH AUTHOR\n%s\n", author)
	}
}

// Version Print
func (p *Parser) PrintVersion() {
	fmt.Println(p.Version())
}

// Help Print
func (p *Parser) PrintHelp() {
	// Usage
	a, s := p.Name(), p.Synopsis()
	if s == "" {
		if p.AllowPreArgs() {
			s += "[BARGS...] "
		}
		s += "[OPTIONS] [ARGS]"
		if p.AllowPostArgs() {
			s += " [EARGS...]"
		}
	}
	fmt.Println("Usage:", a, s)
	fmt.Println(p.Description())
	fmt.Println("\n\033[1;31mOptions\033[m")
	for _, f := range p.flags {
		if f.Hidden() {
			continue
		}
		sf, lf, vn, dv, d := f.Short(), f.Long(), f.ValueName(), f.DefaultValue(), f.Description()
		var flags, args string
		switch {
		case lf == "":
			flags = sf + " "
		case sf == "":
			flags = "    " + lf
		default:
			flags = sf + ", " + lf
		}
		if vn != "" {
			args = vn
			if dv != "" {
				args = "[" + args + "]"
			}
			if lf != "" {
				args = " " + args
			}
			args += " "
		}
		i := 2 + len(flags) + len(args)
		if i > 28 {
			d = "\n" + d
		} else {
			for c := 0; c < 30-i; c++ {
				d = " " + d
			}
		}
		fmt.Printf("  \033[1m%s\033[1;33m%s\033[1;34m%s\033[m\n", flags, args, d)
	}
}
