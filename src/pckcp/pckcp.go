package main

import (
	"bytes"
	"fmt"
	"gettext"
	"io/ioutil"
	"os"
	"parser/flag"
	"parser/pkgbuild"
	"repo"
	"strconv"
	"strings"
	"sysutil"
)

//Flags informations
const (
	LONGDESCRIPTION = `Provides a tool to check the validity of a PKGBUILD according to the KCP standards.

If flag -e is used, the common errors can be checked and a (potentially) valid PKGBUILD.new is created.`
	APP_DESCRIPTION = "Tool in command-line to manage common PKGBUILD errors"
	SYNOPSIS        = "[-h|-e|-v|-g[-c]]"
	D_HELP          = "Print this help"
	D_VERSION       = "Print version"
	D_EDIT          = "Interactive edition"
	D_GENERATE      = "Generate a prototype of PKGBUILD"
	D_CLEAN         = "Removes the useless comments and blanks of the prototype"
)

var (
	flags                                             *flag.Parser
	fHelp, fVersion, fEdit, fDebug, fGenerate, fClean *bool
)

//PKGBUILD declarations
var p *pkgbuild.Pkgbuild
var dependExceptions []string

//Messages types
const (
	E = "Error"
	W = "Warning"
	I = "Info"
	N = ""
)

//Messages checkers
const (
	I_ARCH         = "arch is clean."
	I_CONFLICTS    = "Variable '%s' is clean."
	I_DEPENDS      = "'%s' is clean."
	I_HEADER       = "Header is clean."
	I_INSTALL      = "install is clean."
	I_PACKAGE      = "package() function is present."
	I_PKGREL       = "pkgrel is clean."
	I_SAVED        = "Modifications saved in %s!"
	E_ARCH         = "arch is different from 'x86_64'. Since KaOS only supports this architecture, no other arch would be added here."
	E_DUPLF        = "Function '%s()' is present %d times (lines %s). Only one is expected."
	E_DUPLV        = "Variable '%s' is present %d times (lines %s). Only one is expected."
	E_MISS         = "Variable '%s' is missing."
	E_NOPKGBUILD   = "Current folder doesn't contain PKGBUILD!"
	E_NOTAFILE     = "%s is not a file!" /*#*/
	E_SAVED        = "Error on save file %s!"
	E_UNKNOWN      = "'%s' is not valid. You should suppress it"
	Q_ARCH         = "Reset arch to x86_64?"
	Q_CLEANUP      = "Remove all unneeded comments before save?"
	Q_CONFLICTS    = "Modify %s in variable '%s'?"
	Q_DEPENDS      = "Modify %s as %s?"
	Q_DUPLF        = "Remove unneeded duplicate functions?"
	Q_DUPLV        = "Remove unneeded duplicate variables?"
	Q_EMPTY        = "Remove it?"
	Q_ERASE        = "PKGBUILD file already exists. Do you want to erase it?"
	Q_HEADER       = "Remove header?"
	Q_INSTALL      = "Modify name of %s file?"
	Q_MISS         = "Add variable '%s'?"
	Q_PKGREL       = "Reset pkgrel to 1?"
	Q_UNKNOWN      = "Remove it?"
	R_CONFLICTS    = "Replace %s by... (leave blank to remove it):"
	R_DEPENDS      = "Replace %s by... (leave blank to remove it):"
	R_DUPL         = "Choose the line to keep among [%s] (last line by default)."
	R_INSTALL      = "Replace %s by... (leave blank to remove it):"
	R_MISS         = "Set variable '%s' with (leave blank to ignore):"
	W_CONFLICTS    = "Variable '%s' contains name of the package. It is useless."
	W_CONFLICTS2   = "%s isn't in repo neither in kcp. Variable '%s' shouldn't contain it."
	W_DEPENDS      = "%s isn't in repo neither in kcp. Variable '%s' shouldn't contain it."
	W_EMPTYDEPENDS = "Variables 'depends' and 'makedepends' are empty. You should manually check if it is not a missing."
	W_EMPTYF       = "Function '%s()' is empty."
	W_EMPTYV       = "Variable '%s' is empty."
	W_HEADER       = "Header was found. Do not use names of maintainers or contributors in PKGBUILD, anyone can contribute, keep the header clean from this."
	W_INSTALL      = "%s doesn't exist!"
	W_PACKAGE      = "package() function missing. You need to add it."
	W_PKGREL       = "pkgrel is different from 1. It should be the case only if build instructions are edited but not pkgver."
	W_SPLITTED     = "PKGBUILD is a split PKGBUILD. Make as many PKGBUILDs as this contains different packages!"
	W_CANCEL       = "Action cancelled"
)

//Files paths
const (
	NEWPKGBUILD = "PKGBUILD.new"
	EXCEPTIONS  = "/etc/pckcp/exceptions"
)

//Helpers
var tr = gettext.Gettext

func trf(form string, e ...interface{}) string { return fmt.Sprintf(tr(form), e...) }
func isPackageInRepo(p string) bool {
	switch {
	case strings.Contains(p, "<"):
		p = p[:strings.Index(p, "<")]
	case strings.Contains(p, ">"):
		p = p[:strings.Index(p, ">")]
	case strings.Contains(p, "="):
		p = p[:strings.Index(p, "=")]
	}
	for _, e := range dependExceptions {
		if e == p {
			return true
		}
	}
	b, _ := sysutil.GetOutputCommand("pacman", "-Si", p)
	if len(b) > 0 {
		return true
	}
	b, _ = sysutil.GetOutputCommand("kcp", "-Ns", p)
	for _, e := range strings.Split(string(b), "\n") {
		if e == p {
			return true
		}
	}
	return false
}
func message(tpe string, l1, l2 int, msg string) {
	form := ":\033[m"
	if l1 > 0 {
		if l1 == l2 {
			form = fmt.Sprintf(" (L.%d)%s", l1, form)
		} else {
			form = fmt.Sprintf(" (L.%d-%d)%s", l1, l2, form)
		}
	}
	form = "\033[1;%sm%s" + form + "\n  %s\n"
	var col string
	switch tpe {
	case E:
		col = "31"
	case W:
		col = "33"
	case I:
		col = "32"
	}
	fmt.Printf(form, col, tpe, msg)
}

//Atomic checkers
func checkCommentHeader(bl *pkgbuild.Block) bool {
	ok := true
	for _, d := range bl.Values {
		if d.Type == pkgbuild.DT_COMMENT {
			ok = false
			break
		}
	}
	if !ok {
		message(W, bl.From, bl.To, tr(W_HEADER))
		if *fEdit && sysutil.QuestionYN(tr(Q_HEADER), true) {
			if bl.Values[0].Type == pkgbuild.DT_BLANK {
				bl.Values = bl.Values[:1]
			} else {
				bl.Values = bl.Values[:0]
			}
		}
	}
	return ok
}
func checkEmpty(tpe pkgbuild.DataType, bl *pkgbuild.Block) bool {
	ok := false
	for _, d := range bl.Values {
		if d.Type == tpe {
			ok = true
			break
		}
	}
	if !ok {
		f := W_EMPTYV
		if tpe == pkgbuild.DT_FUNCTION {
			f = W_EMPTYF
		}
		message(W, bl.From, bl.To, trf(f, bl.Name))
		ok = !*fEdit || !sysutil.QuestionYN(tr(Q_EMPTY), true)
	}
	return ok
}
func checkDuplicate(tpe pkgbuild.DataType, l []*pkgbuild.Block) int {
	nb := len(l)
	if nb <= 1 {
		return -1
	}
	s := make([]string, nb)
	for i, bl := range l {
		s[i] = fmt.Sprintf("%d", bl.From)
	}
	f1, f2 := E_DUPLV, Q_DUPLV
	if tpe == pkgbuild.DT_FUNCTION {
		f1, f2 = E_DUPLF, Q_DUPLF
	}
	message(E, 0, 0, trf(f1, l[0].Name, nb, strings.Join(s, ", ")))
	if *fEdit && sysutil.QuestionYN(tr(f2), true) {
		stridx := sysutil.Question(trf(R_DUPL, strings.Join(s, ", ")))
		if i, e := strconv.Atoi(stridx); e == nil && i >= 0 && i < nb {
			return i
		}
		return nb - 1
	}
	return -1
}
func checkArch(bl *pkgbuild.Block) {
	ok := true
	for _, d := range bl.Values {
		if d.Type == pkgbuild.DT_VARIABLE && d.Value != "x86_64" {
			ok = false
			break
		}
	}
	if ok {
		message(I, bl.From, bl.To, tr(I_ARCH))
		return
	}
	message(E, bl.From, bl.To, tr(E_ARCH))
	if *fEdit && sysutil.QuestionYN(tr(Q_ARCH), true) {
		d := &pkgbuild.Data{Type: pkgbuild.DT_VARIABLE, Value: "x86_64"}
		bl.Values = []*pkgbuild.Data{d}
	}
}
func checkPkgrel(bl *pkgbuild.Block) {
	ok := true
	for _, d := range bl.Values {
		if d.Type == pkgbuild.DT_VARIABLE && d.Value != "1" {
			ok = false
			break
		}
	}
	if ok {
		message(I, bl.From, bl.To, tr(I_PKGREL))
		return
	}
	message(W, bl.From, bl.To, tr(W_PKGREL))
	if *fEdit && sysutil.QuestionYN(tr(Q_PKGREL), false) {
		d := &pkgbuild.Data{Type: pkgbuild.DT_VARIABLE, Value: "1"}
		bl.Values = []*pkgbuild.Data{d}
	}
}
func checkInstall(bl *pkgbuild.Block, install string) bool {
	if _, err := os.Stat(install); err == nil {
		message(I, bl.From, bl.To, tr(I_INSTALL))
		return true
	}
	message(W, bl.From, bl.To, trf(W_INSTALL, install))
	if *fEdit && sysutil.QuestionYN(trf(Q_INSTALL, install), true) {
		install = sysutil.Question(trf(R_INSTALL, install))
		if install == "" {
			return false
		} else {
			d := &pkgbuild.Data{
				Type:  pkgbuild.DT_VARIABLE,
				Value: install,
			}
			bl.Values = []*pkgbuild.Data{d}
		}
	}
	return true
}
func checkConflict(bl *pkgbuild.Block, pkgname string) bool {
	var keep []*pkgbuild.Data
	ok := true
	for _, d := range bl.Values {
		if d.Type != pkgbuild.DT_VARIABLE {
			if *fEdit {
				keep = append(keep, d)
			}
			continue
		}
		exists := true
		if pkgname == d.Value {
			exists = false
			message(W, bl.From, bl.To, trf(W_CONFLICTS, bl.Name))
		} else if !isPackageInRepo(d.Value) {
			exists = false
			message(W, bl.From, bl.To, trf(W_CONFLICTS2, d.Value, bl.Name))
		}
		if !exists {
			ok = false
			if *fEdit {
				if sysutil.QuestionYN(trf(Q_CONFLICTS, d.Value, bl.Name), true) {
					d.Value = sysutil.Question(trf(R_CONFLICTS, d.Value))
					if d.Value != "" {
						keep = append(keep, d)
					}
				} else {
					keep = append(keep, d)
				}
			}
		}
	}
	if ok {
		message(I, bl.From, bl.To, trf(I_CONFLICTS, bl.Name))
	}
	if *fEdit {
		bl.Values = keep
		return len(keep) > 0
	}
	return true
}
func checkMissing(p *pkgbuild.Pkgbuild) {
	for _, n := range pkgbuild.L_NEEDED {
		if _, ok := p.Variables[n]; !ok {
			message(E, 0, 0, trf(E_MISS, n))
			if *fEdit && sysutil.QuestionYN(trf(Q_MISS, n), true) {
				if s := sysutil.Question(trf(R_MISS, n)); s != "" {
					d := &pkgbuild.Data{Type: pkgbuild.DT_VARIABLE, Value: s}
					b := &pkgbuild.Block{Type: pkgbuild.BT_VARIABLE, Name: n}
					b.Values = []*pkgbuild.Data{d}
					p.Add(b)
				}
			}
		}
	}
}
func checkUnknown(bl *pkgbuild.Block) bool {
	if bl.Type == pkgbuild.BT_UNKNOWN {
		message(E, bl.From, bl.To, trf(E_UNKNOWN, bl.Name))
		return *fEdit && !sysutil.QuestionYN(tr(Q_UNKNOWN), true)
	}
	var keep []*pkgbuild.Data
	for _, d := range bl.Values {
		if d.Type != pkgbuild.DT_UNKNOWN {
			if *fEdit {
				keep = append(keep, d)
			}
			continue
		}
		message(E, d.Line, d.Line, trf(E_UNKNOWN, d.Value))
		if *fEdit && !sysutil.QuestionYN(tr(Q_UNKNOWN), true) {
			keep = append(keep, d)
		}
	}
	if *fEdit {
		bl.Values = keep
	}
	return true
}

//Global checkers
func checkHeaders(p *pkgbuild.Pkgbuild) {
	for i, bl := range p.Headers {
		if bl.From == 1 {
			if !checkCommentHeader(bl) {
				if len(bl.Values) == 0 {
					delete(p.Headers, i)
				}
				return
			}
			break
		}
	}
	message(I, 0, 0, tr(I_HEADER))
}
func checkVariables(p *pkgbuild.Pkgbuild) {
	pkgname := p.Name()
loop:
	for n, l := range p.Variables {
		if i := checkDuplicate(pkgbuild.DT_VARIABLE, l); i >= 0 {
			p.Variables[n] = l[i : i+1]
		}
		if len(p.Variables[n]) > 1 {
			continue
		}
		bl := p.Variables[n][0]
		checkUnknown(bl)
		if !checkEmpty(pkgbuild.DT_VARIABLE, bl) {
			delete(p.Variables, n)
			continue
		}
		switch n {
		case pkgbuild.ARCH:
			checkArch(bl)
		case pkgbuild.PKGREL:
			checkPkgrel(bl)
		case pkgbuild.INSTALL:
			if !checkInstall(bl, p.Variable(pkgbuild.INSTALL)) {
				delete(p.Variables, n)
			}
		case pkgbuild.CONFLICTS, pkgbuild.PROVIDES, pkgbuild.REPLACES:
			fallthrough
		case pkgbuild.DEPENDS, pkgbuild.MAKEDEPENDS, pkgbuild.OPTDEPENDS, pkgbuild.CHECKDEPENDS:
			if !checkConflict(bl, pkgname) {
				delete(p.Variables, n)
				continue loop
			}
		}
	}
	checkMissing(p)
	if _, ok := p.Variables[pkgbuild.DEPENDS]; !ok {
		if _, ok := p.Variables[pkgbuild.MAKEDEPENDS]; !ok {
			message(W, 0, 0, tr(W_EMPTYDEPENDS))
		}
	}
}
func checkFunctions(p *pkgbuild.Pkgbuild) {
	ok := false
	for n, l := range p.Functions {
		if i := checkDuplicate(pkgbuild.DT_FUNCTION, l); i >= 0 {
			p.Functions[n] = l[i : i+1]
		}
		if len(p.Functions[n]) > 1 {
			continue
		}
		bl := p.Functions[n][0]
		checkUnknown(bl)
		if !checkEmpty(pkgbuild.DT_FUNCTION, bl) {
			delete(p.Functions, n)
			continue
		}
		if n == pkgbuild.PACKAGE {
			ok = true
			message(I, bl.From, bl.To, tr(I_PACKAGE))
		} else if strings.HasPrefix(n, pkgbuild.PACKAGE) {
			message(W, bl.From, bl.To, tr(W_SPLITTED))
		}
	}
	if !ok {
		message(E, 0, 0, tr(W_PACKAGE))
	}
}
func checkUnknowns(p *pkgbuild.Pkgbuild) {
	var keep []*pkgbuild.Block
	for _, bl := range p.Unknown {
		if checkUnknown(bl) {
			keep = append(keep, bl)
		}
	}
	p.Unknown = keep
}

//Generate proto
func generate(clean bool) {
	b, err := repo.PkgbuildProto()
	if err != nil {
		sysutil.PrintError(err)
		os.Exit(1)
	}
	if info, err := os.Stat("PKGBUILD"); err == nil {
		mode := info.Mode()
		if !mode.IsRegular() {
			sysutil.PrintError(trf(E_NOTAFILE, "PKGBUILD"))
			os.Exit(1)
		}
		if !sysutil.QuestionYN(tr(Q_ERASE), false) {
			sysutil.PrintWarning(tr(W_CANCEL))
			return
		}
	}
	if clean {
		prefixes := []string{"prepare()", "build()", "check()", "package()"}
		buf := new(bytes.Buffer)
		for _, l := range strings.Split(string(b), "\n") {
			if lt := strings.TrimSpace(l); lt != "" && lt[0] != '#' {
				for _, pref := range prefixes {
					if strings.HasPrefix(lt, pref) {
						buf.WriteByte('\n')
						break
					}
				}
				buf.WriteString(l + "\n")
			}
		}
		b = buf.Bytes()
	}
	if err := ioutil.WriteFile("PKGBUILD", b, 0644); err != nil {
		sysutil.PrintError(trf(E_SAVED, "PKGBUILD"))
		os.Exit(1)
	}
	sysutil.PrintWarning(trf(I_SAVED, "PKGBUILD"))
}

func init() {
	//Init the locales
	os.Setenv("LANGUAGE", os.Getenv("LC_MESSAGES"))
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.BindTextdomain("pckcp", sysutil.LOCALE_DIR)
	gettext.Textdomain("pckcp")

	//Init the flags
	flags = flag.NewParser(tr(APP_DESCRIPTION), sysutil.VERSION)
	flags.Set(flag.SYNOPSIS, tr(SYNOPSIS))
	flags.Set(flag.LONGDESCRIPTION, tr(LONGDESCRIPTION))

	fHelp, _ = flags.Bool("-h", "--help", tr(D_HELP))
	fVersion, _ = flags.Bool("-v", "--version", tr(D_VERSION))
	fEdit, _ = flags.Bool("-e", "--edit", tr(D_EDIT))
	fGenerate, _ = flags.Bool("-g", "--generate", tr(D_GENERATE))

	fClean, _ = flags.Bool("-c", "--clean", tr(D_CLEAN))
	flags.Require("-c", "-g")

	fDebug, _ = flags.Bool("-d", "--debug", "")
	flags.GetFlag("--debug").Set(flag.HIDDEN, true)
	flags.Group("-h", "-e", "-v", "-g")

	//Init the depends exceptions
	b, _ := ioutil.ReadFile(EXCEPTIONS)
	for _, e := range strings.Split(string(b), "\n") {
		if e = strings.TrimSpace(e); e != "" {
			dependExceptions = append(dependExceptions, e)
		}
	}
}

func main() {
	if e := flags.Parse(os.Args); e != nil {
		sysutil.PrintError(e)
		fmt.Println()
		flags.PrintHelp()
		os.Exit(1)
	}
	if *fHelp {
		flags.PrintHelp()
		return
	}
	if *fVersion {
		flags.PrintVersion()
		return
	}
	if *fGenerate {
		generate(*fClean)
		return
	}
	p, e := pkgbuild.Parse("PKGBUILD")
	if e != nil {
		sysutil.PrintError(tr(E_NOPKGBUILD))
		os.Exit(1)
	}
	if *fDebug {
		fmt.Println(p)
		return
	}
	checkHeaders(p)
	checkVariables(p)
	checkFunctions(p)
	checkUnknowns(p)
	if !*fEdit {
		return
	}
	fmt.Println()
	clean := sysutil.QuestionYN(tr(Q_CLEANUP), true)
	e = ioutil.WriteFile(NEWPKGBUILD, p.Unparse(clean), 0644)
	if e != nil {
		sysutil.PrintError(trf(E_SAVED, NEWPKGBUILD))
		os.Exit(1)
	}
	sysutil.PrintWarning(trf(I_SAVED, NEWPKGBUILD))
}
