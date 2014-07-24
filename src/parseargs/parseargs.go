// parseargs project parseargs.go
package parseargs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	NAME = iota
	SHORTNAME
	DESCRIPTION
	LONGDESCRIPTION
	SYNOPSIS
	AUTHOR
	VERSION
	ALLOWPREARGS
	ALLOWPOSTARGS
	VALUENAME
	DEFAULTVALUE
	MULTIPLEVALUES
)

var ParserProperties = map[uint][]string{
	NAME:            []string{},
	DESCRIPTION:     []string{},
	LONGDESCRIPTION: []string{},
	SYNOPSIS:        []string{},
	AUTHOR:          []string{},
	VERSION:         []string{},
	ALLOWPREARGS:    []string{"0", "1"},
	ALLOWPOSTARGS:   []string{"0", "1"},
}

var FlagProperties = map[uint][]string{
	NAME:           []string{},
	SHORTNAME:      []string{},
	DESCRIPTION:    []string{},
	VALUENAME:      []string{},
	DEFAULTVALUE:   []string{},
	MULTIPLEVALUES: []string{"0", "1"},
}

func initProperties(l map[uint][]string) map[uint]string {
	out := make(map[uint]string)
	for k, a := range l {
		if len(a) == 0 {
			out[k] = ""
		} else {
			out[k] = a[0]
		}
	}
	return out
}

func contains(slice []string, value string) bool {
	for _, s := range slice {
		if value == s {
			return true
		}
	}
	return false
}

func checkShortFlag(n string) error {
	switch {
	case len(n) == 0:
		return nil
	case len(n) == 1:
		return errors.New("Invalid short flag " + n + ": too short!")
	case len(n) > 2:
		return errors.New("Invalid short flag " + n + ": too long!")
	case n[0] != '-' || n[1] == '-':
		return errors.New("Invalid short flag " + n + ": must begin with '-'!")
	default:
		return nil
	}
}

func checkLongFlag(n string) error {
	switch {
	case len(n) == 0:
		return nil
	case len(n) < 3:
		return errors.New("Invalid long flag " + n + ": too short!")
	case n[0:2] != "--":
		return errors.New("Invalid long flag " + n + ": must begin with '--'!")
	default:
		return nil
	}
}

type Parser struct {
	properties map[uint]string
	flags      map[string]*flag
	oflags     []*flag
	preargs    []string
	postargs   []string
}

type Group struct {
	in       *Parser
	selected *flag
}

type flag struct {
	properties map[uint]string
	process    func(string) error
	in         *Group
	linked     *flag
}

func (f *flag) Get(property uint) string {
	v, _ := f.properties[property]
	return v
}

func (f *flag) Set(property uint, value string) error {
	if v, ok := FlagProperties[property]; !ok {
		panic("Invalid property: " + string(property))
	} else if len(v) != 0 && !contains(v, value) {
		panic("Invalid property's value: " + value)
	}
	f.properties[property] = value
	switch property {
	case SHORTNAME:
		return checkShortFlag(value)
	case NAME:
		return checkLongFlag(value)
	default:
		return nil
	}
}

func newFlag() *flag {
	f := flag{initProperties(FlagProperties), nil, nil, nil}
	return &f
}

func initFlag(shortname, name, description string, process func(string) error) *flag {
	f := newFlag()
	if err := f.Set(SHORTNAME, shortname); err != nil {
		panic(err.Error())
	}
	if err := f.Set(NAME, name); err != nil {
		panic(err.Error())
	}
	f.Set(DESCRIPTION, description)
	f.process = process
	return f
}

func (p *Parser) boolFlag(shortname, name, description string) (*flag, *bool) {
	out := new(bool)
	*out = false
	process := func(s string) error {
		if s != "" {
			return errors.New("Unexpected arg: " + s)
		}
		*out = true
		return nil
	}
	f := initFlag(shortname, name, description, process)
	return f, out
}

func (p *Parser) stringFlag(shortname, name, description, valuename, value string) (*flag, *string) {
	out := new(string)
	*out = ""
	process := func(s string) error {
		if s == "" {
			return errors.New("Arg needed!")
		}
		*out = s
		return nil
	}
	f := initFlag(shortname, name, description, process)
	if valuename == "" {
		valuename = "ARG"
	}
	f.Set(VALUENAME, valuename)
	f.Set(DEFAULTVALUE, value)
	return f, out
}

func (p *Parser) choiceFlag(shortname, name, description, value string, choices []string) (*flag, *string) {
	if value != "" && !contains(choices, value) {
		panic("Default value must be in proposed choices!")
	}
	out := new(string)
	*out = ""
	process := func(s string) error {
		switch {
		case s == "":
			return errors.New("Arg needed!")
		case !contains(choices, s):
			return errors.New("Invalid arg: " + s)
		default:
			*out = s
		}
		return nil
	}
	f := initFlag(shortname, name, description, process)
	valuename := "["
	for i, v := range choices {
		if i != 0 {
			valuename += "|"
		}
		valuename += v
	}
	valuename += "]"
	f.Set(VALUENAME, valuename)
	f.Set(DEFAULTVALUE, value)
	return f, out
}

func (p *Parser) stringsFlag(shortname, name, description, valuename string) (*flag, *[]string) {
	var out *[]string
	*out = make([]string, 0, 1)
	process := func(s string) error {
		if s != "" {
			*out = append(*out, s)
		}
		return nil
	}
	f := initFlag(shortname, name, description, process)
	if valuename == "" {
		valuename = "ARGS..."
	}
	f.Set(VALUENAME, valuename)
	return f, out
}

func (p *Parser) intFlag(shortname, name, description, valuename string, novalue int) (*flag, *int) {
	out := new(int)
	*out = novalue
	process := func(s string) error {
		var err error
		*out, err = strconv.Atoi(s)
		return err
	}
	f := initFlag(shortname, name, description, process)
	if valuename == "" {
		valuename = "ARG"
	}
	f.Set(VALUENAME, valuename)
	return f, out
}

func (p *Parser) containsFlag(n string) bool {
	_, ok := p.flags[n]
	return ok
}

func (p *Parser) Get(property uint) string {
	v, _ := p.properties[property]
	return v
}

func (p *Parser) Set(property uint, value string) error {
	if v, ok := ParserProperties[property]; !ok {
		return errors.New("Invalid property: " + string(property))
	} else if len(v) != 0 && !contains(v, value) {
		return errors.New("Invalid property's value: " + value)
	}
	p.properties[property] = value
	return nil
}

func New(description, version string) *Parser {
	p := &Parser{
		initProperties(ParserProperties),
		make(map[string]*flag),
		[]*flag{},
		[]string{},
		[]string{},
	}
	p.Set(DESCRIPTION, description)
	p.Set(VERSION, version)
	return p
}

func (p *Parser) add(f *flag) {
	sn := f.Get(SHORTNAME)
	n := f.Get(NAME)
	for _, v := range []string{sn, n} {
		switch {
		case v == "":
			continue
		case p.containsFlag(v):
			panic("Error: flag " + v + " already set!")
		default:
			p.flags[v] = f
		}
	}
	p.oflags = append(p.oflags, f)
}

func (p *Parser) Bool(shortname, name, description string) *bool {
	f, out := p.boolFlag(shortname, name, description)
	p.add(f)
	return out
}

func (p *Parser) String(shortname, name, description, valuename, value string) *string {
	f, out := p.stringFlag(shortname, name, description, valuename, value)
	p.add(f)
	return out
}

func (p *Parser) Choice(shortname, name, description, value string, choices []string) *string {
	f, out := p.choiceFlag(shortname, name, description, value, choices)
	p.add(f)
	return out
}

func (p *Parser) Strings(shortname, name, description, valuename string) *[]string {
	f, out := p.stringsFlag(shortname, name, description, valuename)
	p.add(f)
	return out
}

func (p *Parser) Int(shortname, name, description, valuename string, novalue int) *int {
	f, out := p.intFlag(shortname, name, description, valuename, novalue)
	p.add(f)
	return out
}

func (p *Parser) InsertGroup() *Group {
	return &Group{p, nil}
}

func (g *Group) add(f *flag) {
	f.in = g
	g.in.add(f)
}

func (g *Group) Bool(shortname, name, description string) *bool {
	f, out := g.in.boolFlag(shortname, name, description)
	g.add(f)
	return out
}

func (g *Group) String(shortname, name, description, valuename, value string) *string {
	f, out := g.in.stringFlag(shortname, name, description, valuename, value)
	g.add(f)
	return out
}

func (g *Group) Choice(shortname, name, description, value string, choices []string) *string {
	f, out := g.in.choiceFlag(shortname, name, description, value, choices)
	g.add(f)
	return out
}

func (g *Group) Strings(shortname, name, description, valuename string) *[]string {
	f, out := g.in.stringsFlag(shortname, name, description, valuename)
	g.add(f)
	return out
}

func (g *Group) Int(shortname, name, description, valuename string, novalue int) *int {
	f, out := g.in.intFlag(shortname, name, description, valuename, novalue)
	g.add(f)
	return out
}

func (p *Parser) Link(flag1, flag2 string) {
	switch {
	case !p.containsFlag(flag1):
		panic("Flag " + flag1 + " doesn't exist!")
	case !p.containsFlag(flag2):
		panic("Flag " + flag2 + " doesn't exist!")
	case flag1 == flag2:
		panic("Flag " + flag1 + " cannot be linked to itself!")
	default:
		p.flags[flag1].linked = p.flags[flag2]
	}
}

func (p *Parser) PrintVersion() {
	v := p.Get(VERSION)
	fmt.Println(v)
}

func (p *Parser) PrintUsage() {
	a := p.Get(NAME)
	s := p.Get(SYNOPSIS)
	if s == "" {
		if e := p.Get(ALLOWPREARGS); e == "1" {
			s += "[BARGS...] "
		}
		s += "[OPTIONS] [ARGS]"
		if e := p.Get(ALLOWPOSTARGS); e == "1" {
			s += " [EARGS...]"
		}
	}
	fmt.Println("Usage:", a, s)
}

func (p *Parser) PrintDescription() {
	fmt.Println(p.Get(DESCRIPTION))
}

func (f *flag) Print() {
	sn := f.Get(SHORTNAME)
	n := f.Get(NAME)
	v := f.Get(VALUENAME)
	d := f.Get(DESCRIPTION)
	var flags, args string
	switch {
	case n == "":
		flags = sn + " "
	case sn == "":
		flags = "    " + n
	default:
		flags = sn + ", " + n
	}
	if v != "" {
		args = v
		if n != "" {
			args = "=" + args
		}
		if opt := f.Get(DEFAULTVALUE); opt == "" {
			args = "[" + args + "]"
		}
		if n == "" {
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

func (p *Parser) PrintHelp() {
	p.PrintUsage()
	p.PrintDescription()
	fmt.Println("\n\033[1;31mOptions:\033[m")
	for _, f := range p.oflags {
		f.Print()
	}
}

func (p *Parser) synopsis() string {
	h := new(bytes.Buffer)
	if b := p.Get(ALLOWPREARGS); b == "1" {
		fmt.Fprint(h, " [BARGS...]")
	}
	for _, f := range p.oflags {
		fmt.Fprint(h, " [")
		sn := f.Get(SHORTNAME)
		n := f.Get(NAME)
		switch {
		case sn == "":
			fmt.Fprint(h, "\\-\\-%s", n[2:])
		case n == "":
			fmt.Fprint(h, "\\-%c", sn[1])
		default:
			fmt.Fprint(h, "\\-%c", sn[1])
			fmt.Fprint(h, "|\\-\\-%s", n[2:])
		}
		if vn := f.Get(VALUENAME); vn != "" {
			if dv := f.Get(DEFAULTVALUE); dv == "" {
				fmt.Fprint(h, " [%s]", vn)
			} else {
				fmt.Fprint(h, " %s", vn)
			}
		}
		fmt.Fprint(h, "]")
	}
	if b := p.Get(ALLOWPOSTARGS); b == "1" {
		fmt.Fprint(h, " [EARGS...]")
	}
	return h.String()
}

func (f *flag) Man() string {
	h := new(bytes.Buffer)
	sn := f.Get(SHORTNAME)
	n := f.Get(NAME)
	switch {
	case sn == "":
		fmt.Fprint(h, "\\-\\-%s", n[2:])
	case n == "":
		fmt.Fprint(h, "\\-%c", sn[1])
	default:
		fmt.Fprint(h, "\\-%c", sn[1])
		fmt.Fprint(h, ", \\-\\-%s", n[2:])
	}
	if vn := f.Get(VALUENAME); vn != "" {
		if dv := f.Get(DEFAULTVALUE); dv == "" {
			if n == "" {
				fmt.Fprint(h, " [%s]", vn)
			} else {
				fmt.Fprint(h, "[=%s]", vn)
			}
		} else {
			if n == "" {
				fmt.Fprint(h, " %s", vn)
			} else {
				fmt.Fprint(h, "=%s", vn)
			}
		}
	}
	fmt.Fprint(h, "\n\t%s", f.Get(DESCRIPTION))
	return h.String()
}

func (p *Parser) description() string {
	descr := p.Get(LONGDESCRIPTION)
	h := new(bytes.Buffer)
	lines := strings.Split(descr, "\n")
	for _, l := range lines {
		if l == "" {
			fmt.Fprintln(h, ".PP")
		} else {
			fmt.Fprintln(h, l)
		}
	}
	return h.String()
}

func (p *Parser) PrintMan() {
	app := p.Get(NAME)
	vers := p.Get(VERSION)
	summ := p.Get(DESCRIPTION)
	author := p.Get(AUTHOR)
	fmt.Printf(".TH \"%s\" 1 \"%s\" \"%s\" \"\"\n", app, time.Now().Format("January 2, 2006"), vers)
	fmt.Println(".SH NAME")
	fmt.Println(app, "\\-", summ)
	fmt.Println(".SH SYNOPSIS")
	fmt.Println(app, p.synopsis())
	fmt.Println(".SH DESCRIPTION")
	fmt.Println(p.description())
	fmt.Println(".SH OPTIONS")
	for _, f := range p.oflags {
		fmt.Println(".TP")
		fmt.Println(f.Man())
	}
	if author != "" {
		fmt.Printf(".SH AUTHOR\n%s\n", author)
	}
}

func formatArgs(args []string) []string {
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

func (p *Parser) getFlags(args []string) ([]int, []*flag, error) {
	var err error = nil
	l := len(args)
	idxs, flags := make([]int, 0, l), make([]*flag, 0, l)
	for i, a := range args {
		switch {
		case a[0] != '-':
			continue
		case !p.containsFlag(a):
			err = errors.New("Flag " + a + " invalid!")
			break
		default:
			flags = append(flags, p.flags[a])
			idxs = append(idxs, i)
		}
	}
	return idxs, flags, err
}

func (p *Parser) checkPreArgs(args []string, idxs []int) error {
	prearg := p.Get(ALLOWPREARGS)
	err := errors.New("Pre-args not allowed!")
	switch {
	case len(idxs) == 0:
		if prearg == "1" {
			p.preargs = append(p.preargs, args...)
		} else if postarg := p.Get(ALLOWPOSTARGS); postarg != "1" {
			p.postargs = append(p.postargs, args...)
		} else {
			return err
		}
	case idxs[0] == 0:
		return nil
	default:
		if prearg == "1" {
			p.preargs = append(p.preargs, args...)
		} else {
			return err
		}
	}
	return nil
}

func (p *Parser) parseArgs(idxs []int, flags []*flag, args []string) error {
	l := len(idxs) - 1
	for i, f := range flags[:l] {
		if f.in != nil {
			if f.in.selected != nil {
				for j, sel := range flags[:i] {
					if sel == f.in.selected {
						return errors.New("Flag " + args[idxs[i]] + " cannot be used with " + args[idxs[j]])
					}
				}
			} else {
				f.in.selected = f
			}
		}
		i1, i2 := idxs[i]+1, idxs[i+1]
		switch {
		case i1 == i2:
			if err := f.process(f.Get(DEFAULTVALUE)); err != nil {
				return err
			}
		case i1+1 == i2:
			if err := f.process(args[i1]); err != nil {
				return err
			}
		case f.Get(MULTIPLEVALUES) == "0":
			return errors.New("Flag " + args[i1-1] + " doesn't allow multiple arguments!")
		default:
			for _, a := range args[i1:i2] {
				if err := f.process(a); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *Parser) parseLastArg(idxs []int, flags []*flag, args []string) error {
	l := len(idxs) - 1
	idx, f := idxs[l], flags[l]
	if f.in != nil {
		if f.in.selected != nil {
			for j, sel := range flags[:l] {
				if sel == f.in.selected {
					return errors.New("Flag " + args[idxs[l]] + " cannot be used with " + args[idxs[j]])
				}
			}
		} else {
			f.in.selected = f
		}
	}
	switch {
	case idx+1 == len(args):
		return f.process(f.Get(DEFAULTVALUE))
	case idx+2 == len(args):
		return f.process(args[idx+1])
	case f.Get(MULTIPLEVALUES) == "1":
		for _, a := range args[idx+1:] {
			if err := f.process(a); err != nil {
				return err
			}
		}
	default:
		if err := f.process(args[idx+1]); err != nil {
			return err
		}
	}
	if b := p.Get(ALLOWPOSTARGS); b == "1" {
		p.postargs = append(p.postargs, args[idx+2:]...)
		return nil
	}
	return errors.New("Post-args not allowed!")
}

func (f *flag) inSlice(flags []*flag) bool {
	for _, fl := range flags {
		if fl == f {
			return true
		}
	}
	return false
}

func (p *Parser) Parse(args []string) error {
	if len(args) == 0 {
		return errors.New("Parser needs at least one argument!")
	}
	_, app := path.Split(args[0])
	p.Set(NAME, app)
	args = formatArgs(args[1:])
	if contains(args, "--create-manpage") {
		p.PrintMan()
		os.Exit(0)
	}
	if len(args) == 0 {
		return nil
	}
	idxs, flags, err := p.getFlags(args)
	if err != nil {
		return err
	}
	if err := p.checkPreArgs(args, idxs); err != nil {
		return err
	} else if len(idxs) == 0 {
		return nil
	}
	if err := p.parseArgs(idxs, flags, args); err != nil {
		return err
	}
	if err := p.parseLastArg(idxs, flags, args); err != nil {
		return err
	}
	for i, f := range flags {
		if lnk := f.linked; lnk != nil && !lnk.inSlice(flags) {
			n := lnk.Get(SHORTNAME)
			if n == "" {
				n = lnk.Get(NAME)
			}
			return errors.New("Flag " + args[idxs[i]] + " needs flag " + n + ".")
		}
	}
	return nil
}

func (p *Parser) getPreArgs() []string {
	return p.preargs
}

func (p *Parser) getPostArgs() []string {
	return p.postargs
}
