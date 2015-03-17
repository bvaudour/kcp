package main

import (
	"bufio"
	"fmt"
	"gettext"
	"os"
	"os/exec"
	"parser/pkgbuild"
	"regexp"
	"strings"
)

const (
	E int = iota
	W
	I
	N
)

const (
	E_NOPKGBUILD   = "Current folder doesn't contain PKGBUILD!"
	I_HEADER       = "Header is clean."
	W_HEADER       = "Header was found. Do not use names of maintainers or contributors in PKGBUILD, anyone can contribute, keep the header clean from this."
	Q_HEADER       = "Remove header?"
	I_PKGREL       = "pkgrel is clean."
	W_PKGREL       = "pkgrel is different from 1. It should be the case only if build instructions are edited but not pkgver."
	Q_PKGREL       = "Reset pkgrel to 1?"
	I_ARCH         = "arch is clean."
	W_ARCH         = "arch is different from 'x86_64'. Since KaOS only supports this architecture, no other arch would be added here."
	Q_ARCH         = "Reset arch to x86_64?"
	W_EMPTYVAR     = "Variable '%s' is empty."
	Q_EMPTYVAR     = "Remove variable '%s'?"
	I_CONFLICTS    = "Variable '%s' is clean."
	W_CONFLICTS    = "Variable '%s' contains name of the package. It is useless."
	W_CONFLICTS2   = "%s isn't in repo neither in kcp. Variable '%s' shouldn't contain it."
	Q_CONFLICTS    = "Modify %s in variable '%s'?"
	Q_CONFLICTS2   = "Replace %s by... (leave blank to remove it):"
	W_SPLITTED     = "PKGBUILD is a split PKGBUILD. Make as many PKGBUILDs as this contains different packages!"
	I_PACKAGE      = "package() function is present."
	W_PACKAGE      = "package() function missing. You need to add it."
	W_EMPTYDEPENDS = "Variables 'depends' and 'makedepends' are empty. You should manually check if it is not a missing."
	I_DEPENDS      = "'%s' is clean."
	W_DEPENDS      = "%s isn't in repo neither in kcp. Variable '%s' shouldn't contain it."
	Q_DEPENDS      = "Modify %s as %s?"
	Q_DEPENDS2     = "Replace %s by... (leave blank to remove it):"
	I_URL          = "url is clean."
	W_URL          = "No url specified."
	Q_URL          = "Add url?"
	T_URL          = "Please, type the URL to include:"
	SYNOPSIS       = "%s is a simple PKGBUILD Checker for the KaOS Community Packages."
	LOCALE_DIR     = "/usr/share/locale"
	NEWPKGBUILD    = "PKGBUILD.new"
	I_SAVED        = "Modifications saved in %s!"
	E_SAVED        = "Error on save file %s!"
	I_INSTALL      = "install is clean."
	W_INSTALL      = "%s doesn't exist!"
	Q_INSTALL      = "Modify name of %s file?"
	T_INSTALL      = "Replace %s by... (leave blank to remove it):"
)

// List of exceptions for depends
var exceptions = []string{
	"java-runtime",
	"java-environment",
	"libreoffice-en-US",
	"libreoffice-langpack",
	"phonon-backend",
	"phonon-backend-qt5",
}

func launch(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if out, err := cmd.Output(); err == nil {
		return string(out), nil
	} else {
		return "", err
	}
}

func pkgname(p *pkgbuild.Pkgbuild) string {
	if bl, ok := p.Variables[pkgbuild.PKGNAME]; ok {
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_VARIABLE {
				return d.String()
			}
		}
	}
	return ""
}

func exists_package(p string) bool {
	switch {
	case strings.Contains(p, "<"):
		p = p[:strings.Index(p, "<")]
	case strings.Contains(p, ">"):
		p = p[:strings.Index(p, ">")]
	case strings.Contains(p, "="):
		p = p[:strings.Index(p, "=")]
	}
	for _, e := range exceptions {
		if e == p {
			return true
		}
	}
	s, _ := launch("pacman", "-Si", p)
	if s != "" {
		return true
	}
	s, _ = launch("kcp", "-Ns", p)
	for _, e := range strings.Split(s, "\n") {
		if e == p {
			return true
		}
	}
	return false
}

func t(s string) string {
	return gettext.Gettext(s)
}

func message(s string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(s, a...))
}

func message_check(t, l1, l2 int, s string, a ...interface{}) {
	b := ":\033[m"
	if l1 > 0 {
		if l1 == l2 {
			b = fmt.Sprintf(" (L.%d)%s", l1, b)
		} else {
			b = fmt.Sprintf(" (L.%d-%d)%s", l1, l2, b)
		}
	}
	b = "\033[1;%sm%s" + b
	var cl, tp string
	switch t {
	case E:
		cl, tp = "31", "Error"
	case W:
		cl, tp = "33", "Warning"
	case I:
		cl, tp = "32", "Info"
	}
	message(b, cl, tp)
	message("  "+s, a...)
}

func question(msg string, defaultValue bool) bool {
	var defstr string = "[Y/n]"
	if !defaultValue {
		defstr = "[y/N]"
	}
	fmt.Printf("\033[1;33m%s %s \033[m", msg, defstr)
	var response string
	if _, err := fmt.Scanf("%v", &response); err != nil || len(response) == 0 {
		return defaultValue
	}
	response = strings.ToLower(response)
	switch {
	case strings.HasPrefix(response, "y"):
		return true
	case strings.HasPrefix(response, "n"):
		return false
	default:
		return defaultValue
	}
}

func response(msg string) string {
	fmt.Print(msg + " ")
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return sc.Text()
}

// Checkers
func check_header(p *pkgbuild.Pkgbuild, edit bool) {
	b, e, ok := 0, 0, true
	bl := p.Header
	if bl != nil {
		b, e = bl.Begin, bl.End
		for _, d := range bl.Header {
			if d.Type == pkgbuild.DT_COMMENT {
				ok = false
				break
			}
		}
	}
	if ok {
		message_check(I, b, e, t(I_HEADER))
	} else {
		message_check(W, b, e, t(W_HEADER))
		if edit && question(t(Q_HEADER), true) {
			if bl.Header[0].Type == pkgbuild.DT_BLANK {
				bl.Header = bl.Header[:1]
			} else {
				bl.Header = make([]*pkgbuild.Data, 0)
			}
		}
	}
}

func check_arch(p *pkgbuild.Pkgbuild, edit bool) {
	b, e, ok := 0, 0, false
	bl, _ := p.Variables[pkgbuild.ARCH]
	if bl != nil {
		b, e, ok = bl.Begin, bl.End, true
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_VARIABLE && d.String() != "x86_64" {
				ok = false
				break
			}
		}
	}
	if ok {
		message_check(I, b, e, t(I_ARCH))
	} else {
		message_check(E, b, e, t(W_ARCH))
		if edit && question(t(Q_ARCH), true) {
			if bl == nil {
				bl = pkgbuild.NewBlock(pkgbuild.ARCH, -1, pkgbuild.BT_VARIABLE)
				bl.End = -1
				p.Append(bl)
			} else {
				bl.Values = make([]*pkgbuild.Data, 0)
			}
			bl.AppendDataString("x86_64", pkgbuild.DT_VARIABLE, bl.Begin)
		}
	}
}

func check_pkgrel(p *pkgbuild.Pkgbuild, edit bool) {
	b, e, ok := 0, 0, false
	bl, _ := p.Variables[pkgbuild.PKGREL]
	if bl != nil {
		b, e, ok = bl.Begin, bl.End, true
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_VARIABLE && d.String() != "1" {
				ok = false
				break
			}
		}
	}
	if ok {
		message_check(I, b, e, t(I_PKGREL))
	} else {
		tp, d := W, false
		if bl == nil {
			tp, d = E, true
		}
		message_check(tp, b, e, t(W_PKGREL))
		if edit && question(t(Q_PKGREL), d) {
			if bl == nil {
				bl = pkgbuild.NewBlock(pkgbuild.PKGREL, -1, pkgbuild.BT_VARIABLE)
				bl.End = -1
				p.Append(bl)
			} else {
				bl.Values = make([]*pkgbuild.Data, 0)
			}
			bl.AppendDataString("1", pkgbuild.DT_VARIABLE, bl.Begin)
		}
	}
}

func check_conflicts(p *pkgbuild.Pkgbuild, edit bool) {
	pn := pkgname(p)
	lconf := []string{pkgbuild.CONFLICTS, pkgbuild.PROVIDES, pkgbuild.REPLACES}
	for _, n := range lconf {
		bl, ok := p.Variables[n]
		if !ok {
			continue
		}
		b, e := bl.Begin, bl.End
		c, keep := 0, make([]*pkgbuild.Data, 0)
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_COMMENT {
				if edit {
					keep = append(keep, d)
				}
				continue
			}
			c++
			okt := true
			if pn != "" && d.String() == pn {
				okt = false
				message_check(W, b, e, t(W_CONFLICTS), n)
			} else if !exists_package(d.String()) {
				okt = false
				message_check(W, b, e, t(W_CONFLICTS2), d.String(), n)
			}
			if edit {
				if okt {
					keep = append(keep, d)
				} else if question(fmt.Sprintf(t(Q_CONFLICTS), d.String(), n), true) {
					d.Value = strings.TrimSpace(response(fmt.Sprintf(t(Q_CONFLICTS2), d.String())))
					if d.String() != "" {
						keep = append(keep, d)
					}
				}
			}
			if !okt {
				ok = false
			}
		}
		if edit {
			bl.Values = keep
		}
		if ok && c > 0 {
			message_check(I, b, e, t(I_CONFLICTS), n)
		}
	}
}

func check_depends(p *pkgbuild.Pkgbuild, edit bool) {
	lconf := []string{pkgbuild.DEPENDS, pkgbuild.MAKEDEPENDS, pkgbuild.OPTDEPENDS, pkgbuild.CHECKDEPENDS}
	for _, n := range lconf {
		bl, ok := p.Variables[n]
		if !ok {
			continue
		}
		b, e := bl.Begin, bl.End
		c, keep := 0, make([]*pkgbuild.Data, 0)
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_COMMENT {
				if edit {
					keep = append(keep, d)
				}
				continue
			}
			c++
			okt := true
			v := d.String()
			if n == pkgbuild.OPTDEPENDS {
				lst := strings.Split(v, ":")
				v = lst[0]
			}
			if !exists_package(v) {
				okt = false
				message_check(W, b, e, t(W_DEPENDS), v, n)
			}
			if edit {
				if okt {
					keep = append(keep, d)
				} else if question(fmt.Sprintf(t(Q_DEPENDS), v, n), true) {
					d.Value = strings.TrimSpace(response(fmt.Sprintf(t(Q_DEPENDS2), v)))
					if d.String() != "" {
						keep = append(keep, d)
					}
				}
			}
			if !okt {
				ok = false
			}
		}
		if edit {
			bl.Values = keep
		}
		if ok && c > 0 {
			message_check(I, b, e, t(I_DEPENDS), n)
		}
	}
}

func check_emptyvar(p *pkgbuild.Pkgbuild, edit bool) {
	for _, bl := range p.Variables {
		n, b, e := bl.Name, bl.Begin, bl.End
		ok := false
		for _, d := range bl.Values {
			if d.Type == pkgbuild.DT_VARIABLE {
				ok = true
				break
			}
		}
		if !ok {
			message_check(W, b, e, t(W_EMPTYVAR), n)
			if edit && !question(fmt.Sprintf(t(Q_EMPTYVAR), n), true) {
				delete(p.Variables, n)
			}
		}
	}
}

func check_emptydepends(p *pkgbuild.Pkgbuild, edit bool) {
	hasdepend := false
	for _, n := range []string{pkgbuild.DEPENDS, pkgbuild.MAKEDEPENDS} {
		if _, ok := p.Variables[n]; ok {
			hasdepend = true
			break
		}
	}
	if !hasdepend {
		message_check(W, 0, 0, t(W_EMPTYDEPENDS))
	}
}

func check_package(p *pkgbuild.Pkgbuild, edit bool) {
	ok := false
	for k, bl := range p.Functions {
		if k == pkgbuild.PACKAGE {
			ok = true
			message_check(I, bl.Begin, bl.End, t(I_PACKAGE))
		} else if strings.HasPrefix(k, pkgbuild.PACKAGE) {
			message_check(W, bl.Begin, bl.End, t(W_SPLITTED))
		}
	}
	if !ok {
		message_check(E, 0, 0, t(W_PACKAGE))
	}
}

func check_url(p *pkgbuild.Pkgbuild, edit bool) {
	if bl, ok := p.Variables[pkgbuild.URL]; ok {
		message_check(I, bl.Begin, bl.End, t(I_URL))
	} else {
		message_check(W, 0, 0, t(W_URL))
		if edit && question(t(Q_URL), true) {
			s := strings.TrimSpace(response(T_URL))
			if s == "" {
				return
			}
			bl := pkgbuild.NewBlock(pkgbuild.URL, -1, pkgbuild.BT_VARIABLE)
			bl.End = -1
			bl.AppendDataString(s, pkgbuild.DT_VARIABLE, bl.Begin)
			p.Append(bl)
		}
	}
}

func check_install(p *pkgbuild.Pkgbuild, edit bool) {
	bl, ok := p.Variables[pkgbuild.INSTALL]
	if !ok {
		return
	}
	b, e := bl.Begin, bl.End
	install, ok := "", false
	for _, d := range bl.Values {
		if d.Type == pkgbuild.DT_VARIABLE {
			install = d.String()
			ok = true
			break
		}
	}
	if !ok {
		return
	}
	r := regexp.MustCompile(`\$\w+|\$\{.+\}`)
	if r.MatchString(install) {
		v := r.FindStringSubmatch(install)[0]
		k := strings.Trim(v, "${}")
		repl := ""
		if bl2, ok := p.Variables[k]; ok {
			for _, d := range bl2.Values {
				if d.Type == pkgbuild.DT_VARIABLE {
					repl = d.String()
					break
				}
			}
		}
		install = r.ReplaceAllString(v, repl)
	}
	if _, err := os.Stat(install); err == nil {
		message_check(I, b, e, t(I_INSTALL))
	} else {
		message_check(W, b, e, t(W_INSTALL), install)
		if edit && question(fmt.Sprintf(t(Q_INSTALL), install), true) {
			install = response(fmt.Sprintf(t(T_INSTALL), install))
			install = strings.TrimSpace(install)
			if install == "" {
				delete(p.Variables, pkgbuild.INSTALL)
			} else {
				bl.Values = make([]*pkgbuild.Data, 0, 1)
				bl.AppendDataString(install, pkgbuild.DT_VARIABLE, bl.Begin)
			}
		}
	}
}

func init() {
	// Init the locales
	os.Setenv("LANGUAGE", os.Getenv("LC_MESSAGES"))
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.BindTextdomain("pckcp", LOCALE_DIR)
	gettext.Textdomain("pckcp")
}

func main() {
	edit, debug := false, false
	if len(os.Args) > 1 {
		a := os.Args[1]
		switch {
		case a == "-e" || a == "--edit":
			edit = true
		case a == "-v" || a == "--version":
			v, _ := launch("kcp", "-v")
			message(v)
			return
		case a == "-d" || a == "--debug":
			debug = true
		default:
			message(t(SYNOPSIS), os.Args[0])
			return
		}
	}
	p, e := pkgbuild.Parse("PKGBUILD")
	if e != nil {
		message(t(E_NOPKGBUILD))
		os.Exit(1)
	}
	if debug {
		fmt.Println(p)
		return
	}
	check_header(p, edit)
	check_arch(p, edit)
	check_pkgrel(p, edit)
	check_conflicts(p, edit)
	check_depends(p, edit)
	check_emptyvar(p, edit)
	check_emptydepends(p, edit)
	check_package(p, edit)
	check_url(p, edit)
	check_install(p, edit)
	if edit {
		e := pkgbuild.Unparse(p, NEWPKGBUILD)
		m, r := "\n\033[1;1m%s\033[m", 0
		if e == nil {
			m = fmt.Sprintf(m, I_SAVED)
		} else {
			m, r = fmt.Sprintf(m, E_SAVED), 1
		}
		message(m, NEWPKGBUILD)
		os.Exit(r)
	}
}
