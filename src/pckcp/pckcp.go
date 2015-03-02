package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	E_SAVEPKGBUILD = "PKGBUILD cannot be saved!"
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
	W_CONFLICTS2   = "%s isn't in repo neither in kcp. Variable '%s' doesn't need to contain it."
	Q_CONFLICTS    = "Remove %s in variable '%s'?"
	W_SPLITTED     = "PKGBUILD is a split PKGBUILD. Make as many PKGBUILDs as this contains different packages!"
	I_PACKAGE      = "package() function is present."
	W_PACKAGE      = "package() function missing. You should have to add it."
	//I_URL          = "url is clean."
	//W_URL          = "No url specified."
	//Q_URL          = "Add url?"
)

func ReadFile(path string) (lines []string, err error) {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		err = e
		lines = []string{}
	} else {
		lines = strings.Split(string(b), "\n")
	}
	return
}

func WriteFile(path string, lines []string) error {
	b := append([]byte(strings.Join(lines, "\n")), '\n')
	return ioutil.WriteFile(path, b, 0644)
}

func LaunchCommandWithResult(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if out, err := cmd.Output(); err == nil {
		return string(out), nil
	} else {
		return "", err
	}
}

//TODO
func t(s string) string {
	return s
}

func message(tpe int, s string, a ...interface{}) {
	b := ""
	switch tpe {
	case E:
		b = "\033[1;31mError:   \033[m"
	case W:
		b = "\033[1;33mWarning: \033[m"
	case I:
		b = "\033[1;32mInfo:    \033[m"
	}
	s = fmt.Sprintf(s, a...)
	fmt.Println(b + s)
}

func s2a(s string) []string {
	if strings.HasPrefix(s, "(") || strings.HasSuffix(s, ")") {
		s = strings.Trim(s, "()")
	}
	out := strings.Fields(s)
	for i, e := range out {
		if strings.HasPrefix(e, "\"") || strings.HasPrefix(e, "'") {
			out[i] = strings.Trim(e, "'\"")
		}
	}
	return out
}

func a2s(a []string) string {
	out := ""
	for _, e := range a {
		if out != "" {
			out += " "
		}
		out += "'" + e + "'"
	}
	return out
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

func open_pkgbuild() []string {
	var out []string
	var e error
	if out, e = ReadFile("PKGBUILD"); e != nil {
		message(E, t(E_NOPKGBUILD))
		os.Exit(1)
	}
	return out
}

func save_pkgbuild(lines []string) {
	if e := WriteFile("PKGBUILD", lines); e != nil {
		message(E, t(E_SAVEPKGBUILD))
	}
}

func read_package(lines []string) string {
	for _, l := range lines {
		l := strings.TrimSpace(l)
		if strings.HasPrefix(l, "pkgname=") {
			return strings.TrimPrefix(l, "pkgname=")
		}
	}
	return ""
}

func exists_package(p string) bool {
	s, e := LaunchCommandWithResult("pacman", "-Q", p)
	if e != nil {
		return false
	}
	if s != "" {
		return true
	}
	s, e = LaunchCommandWithResult("kcp", "-Ns", p)
	return e == nil && s != ""
}

func check_header(lines []string, edit bool) []string {
	p, h := -1, false
	var out []string
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			h = true
			p = i + 1
		} else {
			break
		}
	}
	if h {
		message(W, t(W_HEADER))
		if edit && question(t(Q_HEADER), true) {
			out = lines[p:]
		} else {
			out = lines
		}
	} else {
		message(I, t(I_HEADER))
		out = lines
	}
	return out
}

func check_arch(lines []string, edit bool) []string {
	out := make([]string, 0, len(lines))
	checked := false
	for _, l := range lines {
		lt := strings.TrimSpace(l)
		if strings.HasPrefix(lt, "arch=") {
			checked = true
			flds := s2a(strings.TrimPrefix(lt, "arch="))
			ok := false
			switch {
			case len(flds) != 1:
				message(E, t(W_ARCH))
			case flds[0] != "x86_64":
				message(E, t(W_ARCH))
			default:
				message(I, t(I_ARCH))
				ok = true
			}
			if !ok && edit && question(t(Q_ARCH), true) {
				l = "arch=('x86_64')"
			}
		}
		out = append(out, l)
	}
	if !checked {
		message(E, t(W_ARCH))
		if edit && question(t(Q_ARCH), true) {
			out = append(out, "arch=('x86_64')")
		}
	}
	return out
}

func check_pkgrel(lines []string, edit bool) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		lt := strings.TrimSpace(l)
		if strings.HasPrefix(lt, "pkgrel=") {
			if strings.TrimPrefix(lt, "pkgrel=") != "1" {
				message(W, t(W_PKGREL))
				if edit && question(t(Q_PKGREL), false) {
					l = "pkgrel=1"
				}
			} else {
				message(I, t(I_PKGREL))
			}
		}
		out = append(out, l)
	}
	return out
}

func check_emptyvar(lines []string, edit bool) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		lt := strings.TrimSpace(l)
		if strings.HasSuffix(lt, "=()") {
			lt = strings.TrimSuffix(lt, "=()")
			message(W, t(W_EMPTYVAR), lt)
			if edit && question(fmt.Sprintf(t(Q_EMPTYVAR), lt), true) {
				continue
			}
		}
		out = append(out, l)
	}
	return out
}

func check_conflicts(lines []string, edit bool) []string {
	out := make([]string, 0, len(lines))
	pkgname := read_package(lines)
	for _, l := range lines {
		lt := strings.TrimSpace(l)
		v := ""
		switch {
		case strings.HasPrefix(lt, "provides="):
			v = "provides"
		case strings.HasPrefix(lt, "conflicts="):
			v = "conflicts"
		case strings.HasPrefix(lt, "replaces="):
			v = "replaces"
		}
		if v != "" {
			lst := s2a(strings.TrimPrefix(lt, v+"="))
			keep := make([]string, 0, len(lst))
			ok := true
			for _, e := range lst {
				okt := true
				if e == pkgname {
					okt = false
					message(W, W_CONFLICTS, v)
				} else if !exists_package(e) {
					okt = false
					message(W, W_CONFLICTS2, e, v)
				}
				if !okt {
					ok = false
					if !edit || !question(fmt.Sprintf(Q_CONFLICTS, e, v), true) {
						keep = append(keep, e)
					}
				} else {
					keep = append(keep, e)
				}
			}
			if ok {
				message(I, I_CONFLICTS, v)
			}
			if len(keep) == 0 {
				continue
			}
			l = fmt.Sprintf("%s=(%s)", v, a2s(keep))
		}
		out = append(out, l)
	}
	return out
}

func check_package_func(lines []string, edit bool) []string {
	has_p, splitted := false, false
	for _, l := range lines {
		l := strings.TrimSpace(l)
		if strings.HasPrefix(l, "package()") {
			has_p = true
			break
		}
		if strings.HasPrefix(l, "package_") {
			splitted = true
			break
		}
	}
	if splitted {
		message(W, W_SPLITTED)
	} else if has_p {
		message(I, I_PACKAGE)
	} else {
		message(W, W_PACKAGE)
	}
	return lines
}

func main() {
	edit := false
	if len(os.Args) > 1 {
		a := os.Args[1]
		if a == "-e" || a == "--edit" {
			edit = true
		}
	}
	lines := open_pkgbuild()
	lines = check_header(lines, edit)
	lines = check_arch(lines, edit)
	lines = check_pkgrel(lines, edit)
	lines = check_emptyvar(lines, edit)
	lines = check_conflicts(lines, edit)
	lines = check_package_func(lines, edit)
	if edit {
		save_pkgbuild(lines)
	}
}
