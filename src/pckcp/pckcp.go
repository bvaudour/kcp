package main

import (
	"bufio"
	"fmt"
	"gettext"
	"os"
	"os/exec"
	"parser/pkgbuild"
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

func pkgname(p pkgbuild.Pkgbuild) string {
	if cc, ok := p[pkgbuild.PKGNAME]; ok {
		if len(cc[0].Values) > 0 {
			return cc[0].Values[0].String()
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
func check_header(p pkgbuild.Pkgbuild, edit bool) {
	b, e := 0, 0
	h := true
	var c *pkgbuild.Container
	if cc, ok := p[pkgbuild.HEADER]; ok {
		c = cc[0]
		b, e = c.Begin+1, c.End+1
		for _, d := range c.Values {
			if d.Type == pkgbuild.TD_COMMENT {
				h = false
				break
			}
		}
	}
	if h {
		message_check(I, b, e, t(I_HEADER))
	} else {
		message_check(W, b, e, t(W_HEADER))
		if edit && question(t(Q_HEADER), true) {
			c.Values = make([]*pkgbuild.Data, 0)
			c.Append(pkgbuild.TD_BLANK, b, "")
		}
	}
}

func check_arch(p pkgbuild.Pkgbuild, edit bool) {
	b, e := 0, 0
	h := false
	var c *pkgbuild.Container
	if cc, ok := p[pkgbuild.ARCH]; ok {
		c = cc[0]
		h = true
		b, e = c.Begin+1, c.End+1
		if len(c.Values) != 1 {
			h = false
		} else if c.Values[0].String() != "x86_64" {
			h = false
		}
	}
	if h {
		message_check(I, b, e, t(I_ARCH))
	} else {
		message_check(E, b, e, t(W_ARCH))
		if edit && question(t(Q_ARCH), true) {
			if c == nil {
				c, _ = pkgbuild.NewContainer("arch=('x86_64')", -1)
				c.End = -1
				p.Insert(c)
			} else {
				c.Values = make([]*pkgbuild.Data, 0)
				c.Append(pkgbuild.TD_VARIABLE, b, "x86_64")
			}
		}
	}
}

func check_pkgrel(p pkgbuild.Pkgbuild, edit bool) {
	b, e := 0, 0
	h := false
	var c *pkgbuild.Container
	if cc, ok := p[pkgbuild.PKGREL]; ok {
		c = cc[0]
		h = true
		b, e = c.Begin+1, c.End+1
		if len(c.Values) != 1 || c.Values[0].String() != "1" {
			h = false
		}
	}
	if h {
		message_check(I, b, e, t(I_PKGREL))
	} else {
		tp, d := W, false
		if c == nil {
			tp, d = E, true
		}
		message_check(tp, b, e, t(W_PKGREL))
		if edit && question(t(Q_PKGREL), d) {
			if c == nil {
				c, _ = pkgbuild.NewContainer("pkgrel=1", -1)
				c.End = -1
				p.Insert(c)
			} else {
				c.Values = make([]*pkgbuild.Data, 0)
				c.Append(pkgbuild.TD_VARIABLE, b, "1")
			}
		}
	}
}

func check_conflicts(p pkgbuild.Pkgbuild, edit bool) {
	pn := pkgname(p)
	lconf := []string{pkgbuild.CONFLICTS, pkgbuild.PROVIDES, pkgbuild.REPLACES}
	for _, n := range lconf {
		cc, ok := p[n]
		if !ok {
			continue
		}
		for _, c := range cc {
			b, e := c.Begin+1, c.End+1
			if len(c.Values) == 0 {
				continue
			}
			h := true
			keep := make([]*pkgbuild.Data, 0, len(c.Values))
			for _, d := range c.Values {
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
			}
			if edit {
				c.Values = keep
			}
			if h {
				message_check(I, b, e, t(I_CONFLICTS), n)
			}
		}
	}
}

func check_depends(p pkgbuild.Pkgbuild, edit bool) {
	lconf := []string{pkgbuild.DEPENDS, pkgbuild.MAKEDEPENDS, pkgbuild.OPTDEPENDS, pkgbuild.CHECKDEPENDS}
	for _, n := range lconf {
		cc, ok := p[n]
		if !ok {
			continue
		}
		for _, c := range cc {
			b, e := c.Begin+1, c.End+1
			if len(c.Values) == 0 {
				continue
			}
			h := true
			keep := make([]*pkgbuild.Data, 0, len(c.Values))
			for _, d := range c.Values {
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
			}
			if edit {
				c.Values = keep
			}
			if h {
				message_check(I, b, e, t(I_DEPENDS), n)
			}
		}
	}
}

func check_emptyvar(p pkgbuild.Pkgbuild, edit bool) {
	for k, cc := range p {
		if len(cc) == 0 {
			continue
		}
		if cc[0].Type != pkgbuild.TC_VARIABLE && cc[0].Type != pkgbuild.TC_UVARIABLE {
			continue
		}
		if t, ok := pkgbuild.U_VARIABLES[k]; ok && (t == pkgbuild.TU_SINGLEVAR || t == pkgbuild.TU_SINGLEVARQ) {
			continue
		}
		keep := make([]*pkgbuild.Container, 0, len(cc))
		for _, c := range cc {
			if c.Empty() {
				message_check(W, c.Begin+1, c.End+1, t(W_EMPTYVAR), c.Name)
				if edit && !question(fmt.Sprintf(t(Q_EMPTYVAR), c.Name), true) {
					keep = append(keep, c)
				}
			} else {
				keep = append(keep, c)
			}
		}
		if edit {
			p[k] = keep
		}
	}
}

func check_emptydepends(p pkgbuild.Pkgbuild, edit bool) {
	hasdepend := false
	for _, k := range []string{pkgbuild.DEPENDS, pkgbuild.MAKEDEPENDS} {
		if _, ok := p[k]; ok {
			hasdepend = true
			break
		}
	}
	if !hasdepend {
		message_check(W, 0, 0, t(W_EMPTYDEPENDS))
	}
}

func check_package(p pkgbuild.Pkgbuild, edit bool) {
	has_p, splitted := false, false
	for k, cc := range p {
		if k == pkgbuild.PACKAGE {
			has_p = true
		} else if strings.HasPrefix(k, pkgbuild.PACKAGE) {
			splitted = true
			c := cc[0]
			b, e := c.Begin+1, c.End+1
			message_check(W, b, e, t(W_SPLITTED))
		}
	}
	if !splitted {
		if has_p {
			message_check(I, 0, 0, t(I_PACKAGE))
		} else {
			message_check(E, 0, 0, t(W_PACKAGE))
		}
	}
}

func check_url(p pkgbuild.Pkgbuild, edit bool) {
	if cc, ok := p[pkgbuild.URL]; ok {
		c := cc[0]
		b, e := c.Begin+1, c.End+1
		message_check(I, b, e, t(I_URL))
	} else {
		message_check(W, 0, 0, t(W_URL))
		if edit && question(t(Q_URL), true) {
			s := strings.TrimSpace(response(T_URL))
			if s == "" {
				return
			}
			c, _ := pkgbuild.NewContainer(fmt.Sprintf("url='%s'", s), -1)
			c.End = -1
			p.Insert(c)
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
	edit := false
	if len(os.Args) > 1 {
		a := os.Args[1]
		if a == "-e" || a == "--edit" {
			edit = true
		} else if a == "-v" || a == "--version" {
			v, _ := launch("kcp", "-v")
			message(v)
			return
		} else {
			message(t(SYNOPSIS), os.Args[0])
			return
		}
	}
	p := pkgbuild.ParseFile("PKGBUILD")
	check_header(p, edit)
	check_arch(p, edit)
	check_pkgrel(p, edit)
	check_conflicts(p, edit)
	check_depends(p, edit)
	check_emptyvar(p, edit)
	check_emptydepends(p, edit)
	check_package(p, edit)
	check_url(p, edit)
	if edit {
		e := pkgbuild.UnparseInFile(p, NEWPKGBUILD)
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
