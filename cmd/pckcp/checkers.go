package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	. "github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/pkgbuild"
	"github.com/bvaudour/kcp/pkgbuild/atom"
	"github.com/bvaudour/kcp/pkgbuild/standard"
)

type color int

const (
	cNone color = iota
	cRed
	cGreen
	cYellow

	urlHook = "https://jlk.fjfi.cvut.cz/arch/manpages/man/alpm-hooks.5"
)

var (
	colorType = map[string]color{
		typeError:   cRed,
		typeWarning: cYellow,
		typeInfo:    cGreen,
	}
)

func (c color) String() string {
	if c == 0 {
		return "\033[m"
	}
	return fmt.Sprintf("\033[1;3%dm", c)
}

func message(t string, l1, l2 int, msg string) {
	var position string
	if l1 > 0 {
		if l1 == l2 {
			position = fmt.Sprintf(" (L.%d)", l1)
		} else {
			position = fmt.Sprintf(" (L.%d-%d)", l1, l2)
		}
	}
	fmt.Printf(
		`%s%s%s%s:
  %s
`,
		colorType[t],
		t,
		position,
		cNone,
		msg,
	)
}

func loadExceptions() (exceptions map[string]bool) {
	exceptions = make(map[string]bool)
	fp := Config.Get("pckcp.exceptionsFile")
	if fp == "" {
		return
	}
	fp = JoinIfRelative(ConfigBaseDir, fp)
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		for _, e := range strings.Fields(sc.Text()) {
			exceptions[e] = true
		}
	}
	return
}

func formatPackage(v string) string {
	for _, sep := range []string{">", "<", "=", ":"} {
		i := strings.Index(v, sep)
		if i >= 0 {
			v = v[:i]
		}
	}
	return v
}

func isPackageInRepo(v string) bool {
	result, _ := GetOutputCommand("pacman", "-Si", v)
	if len(result) > 0 {
		return true
	}
	result, _ = GetOutputCommand("kcp", "-Ns", v)
	for _, e := range strings.Fields(string(result)) {
		if e == v {
			return true
		}
	}
	return false
}

func checkHeader(p *pkgbuild.PKGBUILD, edit bool) {
	l := p.Len()
	bh, eh := -1, -1
	for i := 0; i < l; i++ {
		info, isBlank, _ := p.GetIndex(i)
		if info == nil {
			break
		}
		if i == 0 && isBlank {
			continue
		}
		eh = i
		if bh < 0 {
			bh = i
		}
	}
	if bh < 0 {
		message(typeInfo, 0, 0, Tr(infoHeader))
		return
	}

	message(typeWarning, bh+1, eh+1, Tr(warnHeader))
	if edit && QuestionYN(Tr(questionHeader), true) {
		for i := bh; i <= eh; i++ {
			p.RemoveIndex(bh)
		}
	}
}

func checkDuplicates(p *pkgbuild.PKGBUILD, edit bool) {
	names := make(map[string]int)
	var infoNames [][]*atom.Info
	infos := p.GetInfos()
	for _, info := range infos {
		name := info.Name()
		if i, exists := names[name]; exists {
			infoNames[i] = append(infoNames[i], info)
		} else {
			names[name] = len(infoNames)
			infoNames = append(infoNames, []*atom.Info{info})
		}
	}

	var duplicates [][]*atom.Info
	for _, l := range infoNames {
		if len(l) > 1 {
			duplicates = append(duplicates, l)
		}
	}

	if len(duplicates) == 0 {
		message(typeInfo, 0, 0, Tr(infoDuplicate))
		return
	}

	message(typeWarning, 0, 0, Tr(warnDuplicate))
	for _, d := range duplicates {
		name := d[0].Name()
		lines := make([]string, len(d))
		for i, info := range d {
			b, e := info.GetPositions()
			t := "V"
			if info.IsFunc() {
				t = "F"
			}
			if b.Line == e.Line {
				lines[i] = fmt.Sprintf("L.%d (%s)", b.Line, t)
			} else {
				lines[i] = fmt.Sprintf("L.%d-%d (%s)", b.Line, e.Line, t)
			}
		}
		fmt.Printf("  - '%s': %s\n", name, strings.Join(lines, ", "))
	}
	if edit && QuestionYN(Tr(questionDuplicate), true) {
		for _, d := range duplicates {
			for _, i := range d[1:] {
				p.RemoveInfo(i)
			}
		}
	}
}

func checkMissingVars(p *pkgbuild.PKGBUILD, edit bool) {
	var missings []string
	for _, v := range standard.GetVariables() {
		if !standard.IsRequiredVariable(v) {
			continue
		}
		if !p.ContainsVariable(v) {
			missings = append(missings, v)
		}
	}
	if len(missings) == 0 {
		message(typeInfo, 0, 0, Tr(infoMissingVar))
		return
	}
	for _, v := range missings {
		message(typeError, 0, 0, Tr(errMissingVar, v))
		if !(edit && QuestionYN(Tr(questionMissingVar, v), true)) {
			continue
		}
		r := strings.TrimSpace(Question(Tr(questionAddValue, v)))
		if len(r) == 0 {
			continue
		}
		isArray := standard.IsArrayVariable(v)
		var values []string
		if isArray {
			values = strings.Split(r, " ")
		} else {
			values = append(values, r)
		}
		p.AddVariable(v, isArray, values...)
	}
}

func checkMissingFuncs(p *pkgbuild.PKGBUILD, edit bool) {
	var missings []string
	for _, f := range standard.GetFunctions() {
		if !standard.IsRequiredFunction(f) {
			continue
		}
		if !p.ContainsFunction(f) {
			missings = append(missings, f)
		}
	}
	if len(missings) == 0 {
		message(typeInfo, 0, 0, Tr(infoMissingFunc))
		return
	}
	for _, f := range missings {
		message(typeError, 0, 0, Tr(errMissingFunc, f))
		fmt.Printf("  %s\n", commentAddManually)
	}
}

func checkInfoTypes(p *pkgbuild.PKGBUILD, edit bool) {
	infos := p.GetInfos()
	clean := true
	for _, info := range infos {
		name := info.Name()
		p0, p1 := info.GetPositions()
		var actualType, neededType string
		var isBad bool
		if standard.IsStandardFunction(name) {
			if isBad = !info.IsFunc(); isBad {
				actualType, neededType = commentVariable, commentFunction
			}
		} else if standard.IsStandardVariable(name) {
			if isBad = !info.IsVar(); isBad {
				actualType, neededType = commentFunction, commentVariable
			} else {
				actualType, neededType = commentStringVar, commentStringVar
				if info.IsArrayVar() {
					actualType = commentArrayVar
				}
				if standard.IsArrayVariable(name) {
					neededType = commentArrayVar
				}
				isBad = actualType != neededType
			}
		}
		if isBad {
			clean = false
			message(typeWarning, p0.Line, p1.Line, Tr(warnBadType, name, Tr(actualType), Tr(neededType)))
		}
	}
	if clean {
		message(typeInfo, 0, 0, Tr(infoBadType))
	}
}

func checkEmpty(p *pkgbuild.PKGBUILD, edit bool) {
	infos := p.GetInfos(atom.VarArray, atom.VarString)
	clean := true
	for _, info := range infos {
		values := info.ArrayValue()
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			clean = false
			name := info.Name()
			p0, p1 := info.GetPositions()
			message(typeWarning, p0.Line, p1.Line, Tr(warnEmpty, name))
			if edit && QuestionYN(Tr(questionRemoveEmpty, name), true) {
				p.RemoveIndex(info.Index())
			}
		}
	}
	if clean {
		message(typeInfo, 0, 0, Tr(infoEmpty))
	}
}

func checkPkgrel(p *pkgbuild.PKGBUILD, edit bool) {
	info, ok := p.GetInfo(standard.PKGREL, atom.VarString)
	if !ok {
		return
	}
	p0, p1 := info.GetPositions()
	ok, t, m := true, typeInfo, Tr(infoVarClean, standard.PKGREL)
	if info.StringValue() != "1" {
		ok, t, m = false, typeWarning, Tr(warnPkgrel)
	}
	message(t, p0.Line, p1.Line, m)
	if !ok && edit && QuestionYN(Tr(questionPkgrel), false) {
		p.SetValue(info, "1")
	}
}

func checkArch(p *pkgbuild.PKGBUILD, edit bool) {
	info, ok := p.GetInfo(standard.ARCH, atom.VarString, atom.VarArray)
	if !ok {
		return
	}
	p0, p1 := info.GetPositions()
	ok, t, m := true, typeInfo, Tr(infoVarClean, standard.ARCH)
	values := info.ArrayValue()
	if len(values) != 1 || values[0] != "x86_64" {
		ok, t, m = false, typeWarning, Tr(warnArch)
	}
	message(t, p0.Line, p1.Line, m)
	if !ok && edit && QuestionYN(Tr(questionArch), true) {
		if !info.IsArrayVar() {
			p.SetArrayVar(info)
		}
		p.SetValue(info, "x86_64")
	}
}

func checkDep(p *pkgbuild.PKGBUILD, edit bool, info *atom.Info, pkgname string, exceptions map[string]bool) {
	errors := make(map[string]int)
	values := info.ArrayValue()
	for i, v := range values {
		v = formatPackage(v)
		values[i] = v
		if v == pkgname {
			errors[v] = 1
		} else if !exceptions[v] && !isPackageInRepo(v) {
			errors[v] = 2
		}
	}
	name := info.Name()
	p0, p1 := info.GetPositions()
	if len(errors) == 0 {
		message(typeInfo, p0.Line, p1.Line, Tr(infoVarClean, name))
		return
	}
	message(typeWarning, p0.Line, p1.Line, Tr(warnDepends, name))
	var newValues []string
	for _, v := range values {
		switch errors[v] {
		case 1:
			fmt.Printf("  %s\n", Tr(warnPackageIsName, v))
			if edit && QuestionYN("  "+Tr(questionDepend, v), true) {
				v = Question("  " + Tr(questionTypeDepend))
			}
		case 2:
			fmt.Printf("  %s\n", Tr(warnPackageNotInRepo, v))
			if edit && QuestionYN("  "+Tr(questionDepend, v), true) {
				v = Question("  " + Tr(questionTypeDepend))
			}
		}
		if v != "" {
			newValues = append(newValues, v)
		}
	}
	if edit {
		p.SetValue(info, newValues...)
	}
}

func checkDepends(p *pkgbuild.PKGBUILD, edit bool) {
	var hasDepend bool
	pkgname := p.GetValue(standard.PKGNAME)
	names := map[string]bool{
		standard.CONFLICTS:    true,
		standard.PROVIDES:     true,
		standard.REPLACES:     true,
		standard.DEPENDS:      true,
		standard.MAKEDEPENDS:  true,
		standard.OPTDEPENDS:   true,
		standard.CHECKDEPENDS: true,
	}
	nameDepends := map[string]bool{
		standard.DEPENDS:     true,
		standard.MAKEDEPENDS: true,
	}
	exceptions := loadExceptions()
	infos := p.GetInfos(atom.VarArray)
	for _, info := range infos {
		name := info.Name()
		if !names[name] || len(info.ArrayValue()) == 0 {
			continue
		}
		hasDepend = hasDepend || nameDepends[name]
		checkDep(p, edit, info, pkgname, exceptions)
	}
	if !hasDepend {
		message(typeWarning, 0, 0, Tr(warnMissingDepends))
	}
}

func checkInstalls(p *pkgbuild.PKGBUILD, edit bool) {
	info, ok := p.GetInfo(standard.INSTALL, atom.VarString)
	if !ok {
		return
	}
	p0, p1 := info.GetPositions()
	file := info.StringValue()
	if FileExists(file) {
		message(typeInfo, p0.Line, p1.Line, Tr(infoVarClean, standard.INSTALL))
		fmt.Printf("  %s\n", Tr(commentInstall, urlHook))
		return
	}
	message(typeWarning, p0.Line, p1.Line, Tr(warnInstall, file))
	if !(edit && QuestionYN(Tr(questionInstall, file), true)) {
		return
	}
	newFile := Question(Tr(questionInstall2))
	if newFile == "" {
		p.RemoveInfo(info)
	} else {
		p.SetValue(info, newFile)
	}
}
