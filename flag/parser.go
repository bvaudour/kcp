package flag

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"unicode/utf8"
)

// Parser is a structure containing definitions to parse arguments of a CLI.
type Parser struct {
	p        Properties
	flags    []*Flag
	args     map[string]*Flag
	groups   map[*Group][]*Flag
	fgroups  map[*Flag][]*Group
	requires map[*Flag][]*Flag
	pre      []string
	post     []string
}

// Set modifies the given property with the given value.
func (p *Parser) Set(k PropertyType, v any) error {
	return p.p.Set(k, v)
}

// GetString returns the string representation of the needed property
func (p *Parser) GetString(k PropertyType) string {
	return p.p.ValueString(k)
}

// GetBool returns the boolean representation of the needed property.
func (p *Parser) GetBool(k PropertyType) bool {
	return p.p.ValueBool(k)
}

// Name returns the name of the application.
func (p *Parser) Name() string {
	return p.GetString(Name)
}

// Description returns the description of the application.
func (p *Parser) Description() string {
	return p.GetString(Description)
}

// LongDescription returns the long description of the application.
func (p *Parser) LongDescription() string {
	return p.GetString(LongDescription)
}

// Synopsis returns the synopsis of the application.
func (p *Parser) Synopsis() string {
	return p.GetString(Synopsis)
}

// Author returns the author's name of the application.
func (p *Parser) Author() string {
	return p.GetString(Author)
}

// Version returns the version of the application.
func (p *Parser) Version() string {
	return p.GetString(Version)
}

// AllowPreArgs returns true if the parser accepts anonymous args before flags.
func (p *Parser) AllowPreArgs() bool {
	return p.GetBool(AllowPreArs)
}

// AllowPostArgs returns true if the parser accepts anonymous args after flags.
func (p *Parser) AllowPostArgs() bool {
	return p.GetBool(AllowPostArgs)
}

// GetFlag returns the flag where the given name is defined.
func (p *Parser) GetFlag(name string) *Flag {
	f, _ := p.args[name]
	return f
}

// ContainsFlag returns true if a flag with the given name is defined.
func (p *Parser) ContainsFlag(name string) bool {
	_, ok := p.args[name]
	return ok
}

// Add appends a flags to the parser.
// If a flag with same name(s) is (are) defined, return a non nil error.
func (p *Parser) Add(f *Flag) error {
	long, short := f.Long(), f.Short()
	if p.ContainsFlag(long) {
		return NewError(errUnexpectedFlag, long)
	}
	if p.ContainsFlag(short) {
		return NewError(errUnexpectedFlag, short)
	}
	p.flags = append(p.flags, f)
	p.fgroups[f] = make([]*Group, 0)
	if long != "" {
		p.args[long] = f
	}
	if short != "" {
		p.args[short] = f
	}
	return nil
}

// AddAll appends given flags to the parser.
// If a flag cannot be added, return an error.
func (p *Parser) AddAll(flags ...*Flag) (err []error) {
	for _, f := range flags {
		if e := p.Add(f); e != nil {
			err = append(err, e)
		}
	}
	return
}

// NewParser returns an new parser initialized with the description and the version of the application.
func NewParser(description, version string) *Parser {
	p := &Parser{
		p:        ParserProps(),
		args:     make(map[string]*Flag),
		groups:   make(map[*Group][]*Flag),
		fgroups:  make(map[*Flag][]*Group),
		requires: make(map[*Flag][]*Flag),
	}
	p.Set(Description, description)
	p.Set(Version, version)
	return p
}

// Bool appends a new boolean flag and returns a pointer to its value.
func (p *Parser) Bool(s, l, desc string) (*bool, error) {
	f, v, e := NewBoolFlag(s, l, desc)
	if e == nil {
		e = p.Add(f)
	}
	return v, e
}

// String appends a new string flag and returns a pointer to its value.
func (p *Parser) String(s, l, desc, vn, df string) (*string, error) {
	f, v, e := NewStringFlag(s, l, desc, vn, df)
	if e == nil {
		e = p.Add(f)
	}
	return v, e
}

// Choice appends a new choice flag and returns a pointer to its value.
func (p *Parser) Choice(s, l, desc, df string, choices []string) (*string, error) {
	f, v, e := NewChoiceFlag(s, l, desc, df, choices)
	if e == nil {
		e = p.Add(f)
	}
	return v, e
}

// Array appends a new multiple string flag and returns a pointer to its value.
func (p *Parser) Array(s, l, desc, vn string) (*[]string, error) {
	f, v, e := NewArrayFlag(s, l, desc, vn)
	if e == nil {
		e = p.Add(f)
	}
	return v, e
}

// Int appends a new multiple int flag and returns a pointer to its value.
func (p *Parser) Int(s, l, desc, vn string, df int) (*int, error) {
	f, v, e := NewIntFlag(s, l, desc, vn, df)
	if e == nil {
		e = p.Add(f)
	}
	return v, e
}

// Group groups all flags with given names and returns the number of flags in the group.
// Only one flag of a group can be used at same time.
func (p *Parser) Group(names ...string) int {
	g := new(Group)
	var flags []*Flag
	for _, n := range names {
		f := p.GetFlag(n)
		if f != nil {
			flags = append(flags, f)
			p.fgroups[f] = append(p.fgroups[f], g)
		}
	}
	p.groups[g] = flags
	return len(flags)
}

// Require defines all flags required by the flag with name name0 and returns the number of required flags.
func (p *Parser) Require(name0 string, names ...string) int {
	f0 := p.GetFlag(name0)
	if f0 == nil {
		return 0
	}
	flags := make([]*Flag, 0)
	for _, n := range names {
		f := p.GetFlag(n)
		if f != nil {
			flags = append(flags, f)
		}
	}
	p.requires[f0] = flags
	return len(flags)
}

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

// Parse parses the givens arguments according to the definition of the parser.
func (p *Parser) Parse(args []string) error {
	if len(args) == 0 {
		return NewError(errNotEnoughArg)
	}
	_, app := path.Split(args[0])
	p.Set(Name, app)
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
			return NewError(errUnsupportedFlag, a)
		default:
			f := p.GetFlag(a)
			f.used = a
			flags, idxF = append(flags, f), append(idxF, i)
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
			return NewError(errNotAllowed, "Pre-args")
		}
	case p.AllowPreArgs():
		p.pre = args[:idxF[0]]
	case idxF[0] == 0:
	default:
		return NewError(errNotAllowed, "Pre-args")
	}

	// Parse args
	l = len(idxF) - 1
	for i, f := range flags {
		a := args[idxF[i]]
		if grps, ok := p.fgroups[f]; ok {
			for _, g := range grps {
				if g.selected != "" && g.selected != f.Short() && g.selected != f.Long() {
					return NewError(errMustBeAlternative, a, g.selected)
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
			return NewError(errNoMultipleAllowed, a)
		default:
			if e := f.f(args[i1]); e != nil {
				return e
			}
			if p.AllowPostArgs() {
				p.post = args[i1+1:]
			} else {
				return NewError(errNotAllowed, "Post-args")
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
			return NewError(errNeedRequirment, f0.used, str)
		}
	}
	return nil
}

// GetPreArgs returns anonymous args before flags.
func (p *Parser) GetPreArgs() []string {
	return p.pre
}

// GetPostArgs returns anonymous args after flags.
func (p *Parser) GetPostArgs() []string {
	return p.post
}

// Manpage creation
func (f *Flag) man() string {
	b := new(strings.Builder)
	short, long := f.Short(), f.Long()
	switch {
	case short == "":
		fmt.Fprintf(b, `\-\-%s`, long[2:])
	case long == "":
		fmt.Fprintf(b, `\-%s`, short[1:])
	default:
		fmt.Fprintf(b, `\-%s|\-\-%s`, short[1:], long[2:])
	}
	if vn := f.ValueName(); vn != "" {
		if dv := f.DefaultValue(); dv == "" {
			fmt.Fprintf(b, " [%s]", vn)
		} else {
			fmt.Fprintf(b, " %s", vn)
		}
	}
	return b.String()
}
func (p *Parser) synopsis() string {
	b := new(strings.Builder)
	if p.AllowPreArgs() {
		fmt.Fprint(b, " [BARGS...]")
	}
	for _, f := range p.flags {
		fmt.Fprintf(b, " [%s]", f.man())
	}
	if p.AllowPostArgs() {
		fmt.Fprint(b, " [EARGS...]")
	}
	return b.String()
}
func (p *Parser) description() string {
	b := new(strings.Builder)
	for _, l := range strings.Split(p.LongDescription(), "\n") {
		if l == "" {
			fmt.Fprintln(b, ".PP")
		} else {
			fmt.Fprintln(b, l)
		}
	}
	return b.String()
}

// PrinMan prints the Manpage to the standard output according to the parser's definition.
func (p *Parser) PrintMan() {
	app, version, summ, author := p.Name(), p.Version(), p.Description(), p.Author()
	fmt.Printf(`.TH "%s" 1 "%s" "%s" ""`, app, time.Now().Format("January 2, 2006"), version)
	fmt.Println()
	fmt.Println(".SH NAME")
	fmt.Println(app, `\-`, summ)
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

// PrintVersion displays the version of the application.
func (p *Parser) PrintVersion() { fmt.Println(p.Name(), p.Version()) }

// PrintHelp displays the help of the application.
func (p *Parser) PrintHelp() {
	// Usage
	a, s := p.Name(), p.Synopsis()
	if s == "" {
		s = "[OPTIONS] [ARGS]"
		if p.AllowPreArgs() {
			s = fmt.Sprintf("[BARGS...] %s", s)
		}
		if p.AllowPostArgs() {
			s = fmt.Sprintf("%s [EARGS...]", s)
		}
	}
	fmt.Println(usage, a, s)
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
			flags = fmt.Sprintf("%s ", sf)
		case sf == "":
			flags = fmt.Sprintf("    %s", lf)
		default:
			flags = fmt.Sprintf("%s, %s", sf, lf)
		}
		if vn != "" {
			args = vn
			if dv != "" {
				args = fmt.Sprintf("[%s]", args)
			}
			if lf != "" {
				args = fmt.Sprintf(" %s", args)
			}
			args = fmt.Sprintf("%s ", args)
		}
		i := 2 + utf8.RuneCountInString(flags) + utf8.RuneCountInString(args)
		if i > 28 {
			d = "\n" + d
		} else {
			d = strings.Repeat(" ", 30-i) + d
		}
		fmt.Printf("  \033[1m%s\033[1;33m%s\033[1;34m%s\033[m\n", flags, args, d)
	}
}
