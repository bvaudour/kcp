package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/pkgbuild"
	pformat "codeberg.org/bvaudour/kcp/pkgbuild/format"
	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/standard"
	"git.kaosx.ovh/benjamin/collection"
	fformat "git.kaosx.ovh/benjamin/format"
	"mvdan.cc/sh/v3/syntax"
)

const (
	urlHook = "https://archlinux.org/pacman/alpm-hooks.5.html"
)

var (
	colorType = map[string]fformat.Format{
		typeError:   fformat.FormatOf("l_red"),
		typeWarning: fformat.FormatOf("l_yellow"), typeInfo: fformat.FormatOf("l_green"),
	}
)

func message(t string, l1, l2 int, msg string) {
	var position string
	if l1 > 0 {
		if l1 == l2 {
			position = fmt.Sprintf("(L.%d)", l1)
		} else {
			position = fmt.Sprintf("(L.%d-%d)", l1, l2)
		}
	}
	fmt.Printf(
		"%s:\n  %s\n",
		colorType[t].Sprintf("%s %s", t, position),
		msg,
	)
}

func loadExceptions() (exceptions collection.Set[string]) {
	exceptions = collection.NewSet[string]()
	fp := common.Config.Get("pckcp.exceptionsFile")
	if fp == "" {
		return
	}
	fp = common.JoinIfRelative(common.ConfigBaseDir, fp)
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		for e := range strings.FieldsSeq(sc.Text()) {
			exceptions.Add(e)
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
	result, _ := common.GetOutputCommand("pacman", "-Si", v)
	if len(result) > 0 {
		return true
	}
	result, _ = common.GetOutputCommand("kcp", "-Ns", v)
	for e := range strings.FieldsSeq(string(result)) {
		if e == v {
			return true
		}
	}
	return false
}

func checkHeader(p *pkgbuild.PKGBUILD, edit bool) (formatOptions []pformat.FormatOption) {
	if !p.HasHeader() {
		message(typeInfo, 0, 0, common.Tr(infoHeader))
		return
	}
	begin, _ := p.NodeInfoList[0].Position()
	end, _ := p.NodeInfoList[0].InnerPosition()
	message(typeWarning, int(begin.Line()), int(end.Line()-1), common.Tr(warnHeader))
	if edit && common.QuestionYN(common.Tr(questionHeader), true) {
		formatOptions = append(formatOptions, pformat.OptionRemoveHeader)
	}
	return
}

func checkDuplicates(p *pkgbuild.PKGBUILD, edit bool) (formatOptions []pformat.FormatOption) {
	duplicates := p.GetDuplicates()
	if len(duplicates) == 0 {
		message(typeInfo, 0, 0, common.Tr(infoDuplicate))
		return
	}

	message(typeWarning, 0, 0, common.Tr(warnDuplicate))
	for name, nodes := range duplicates {
		lines := make([]string, len(nodes))
		for i, node := range nodes {
			begin, end := node.InnerPosition()
			t := "V"
			if node.Type == info.Function {
				t = "F"
			}
			if begin.Line() == end.Line() {
				lines[i] = fmt.Sprintf("L.%d (%s)", begin.Line(), t)
			} else {
				lines[i] = fmt.Sprintf("L.%d-%d (%s)", begin.Line(), end.Line(), t)
			}
			fmt.Printf("  - '%s': %s\n", name, strings.Join(lines, ", "))
		}
	}

	if edit && common.QuestionYN(common.Tr(questionDuplicate), true) {
		formatOptions = append(formatOptions, pformat.OptionRemoveDuplicates)
	}

	return
}

func checkMissingVars(p *pkgbuild.PKGBUILD, edit bool) (newNodes info.NodeInfoList) {
	missings, missingChecksum := p.GetMissingVariables()
	if missingChecksum && p.HasVariable(standard.SOURCE) {
		message(typeWarning, 0, 0, common.Tr(errMissingChecksum))
		fmt.Printf("  %s\n", commentAddManually)
	} else {
		message(typeInfo, 0, 0, common.Tr(infoMissingChecksum))
	}
	if len(missings) == 0 {
		message(typeInfo, 0, 0, common.Tr(infoMissingVar))
		return
	}

	for _, v := range missings {
		message(typeError, 0, 0, common.Tr(errMissingVar, v))
		if !(edit && common.QuestionYN(common.Tr(questionMissingVar, v), true)) {
			continue
		}
		r := strings.TrimSpace(common.Question(common.Tr(questionAddValue, v)))
		if len(r) == 0 {
			continue
		}
		parsed, err := syntax.NewParser().Parse(strings.NewReader(r), "")
		if err != nil {
			continue
		}
		expr, ok := parsed.Stmts[0].Cmd.(*syntax.CallExpr)
		if !ok || len(expr.Args) == 0 {
			continue
		}
		var line string
		if standard.IsArrayVariable(v) || len(expr.Args) > 1 {
			line = fmt.Sprintf("%s=(%s)", v, r)
		} else {
			line = fmt.Sprintf("%s=%s", v, r)
		}
		parsed, err = syntax.NewParser().Parse(strings.NewReader(line), "")
		if err != nil || len(parsed.Stmts) == 0 {
			continue
		}
		node, err := info.New(0, parsed.Stmts[0], p.Env())
		if err != nil {
			continue
		}
		newNodes = append(newNodes, node)
	}

	return
}

func checkMissingFuncs(p *pkgbuild.PKGBUILD, edit bool) {
	missings := p.GetMissingFunctions()
	if len(missings) == 0 {
		message(typeInfo, 0, 0, common.Tr(infoMissingFunc))
		return
	}

	for _, f := range missings {
		message(typeError, 0, 0, common.Tr(errMissingFunc, f))
		fmt.Printf("  %s\n", commentAddManually)
	}
}

func checkInfoTypes(p *pkgbuild.PKGBUILD, edit bool) (formatOptions []pformat.FormatOption) {
	clean := true
	mtype := map[info.NodeType]string{
		info.Function:  commentFunction,
		info.ArrayVar:  commentArrayVar,
		info.SingleVar: commentStringVar,
	}
	for _, node := range p.NodeInfoList {
		name := node.Name
		begin, end := node.Position()
		actualType, neededType := node.Type, node.Type
		if standard.IsStandardFunction(name) {
			neededType = info.Function
		} else if standard.IsArrayVariable(name) {
			neededType = info.ArrayVar
		} else if standard.IsStandardVariable(name) {
			neededType = info.SingleVar
		}
		if actualType == neededType {
			continue
		}
		clean = false
		message(
			typeWarning,
			int(begin.Line()),
			int(end.Line()),
			common.Tr(warnBadType, name, common.Tr(mtype[actualType]), common.Tr(mtype[neededType])),
		)
	}

	if clean {
		message(typeInfo, 0, 0, common.Tr(infoBadType))
	}

	return
}

func checkEmpty(p *pkgbuild.PKGBUILD, edit bool) (remove []int) {
	clean := true
	for _, node := range p.NodeInfoList {
		if node.Type == info.Function {
			continue
		}
		if node.Value == "" && len(node.Values) == 0 {
			clean = false
			begin, end := node.InnerPosition()
			message(typeWarning, int(begin.Line()), int(end.Line()), common.Tr(warnEmpty, node.Name))
			if edit && common.QuestionYN(common.Tr(questionRemoveEmpty, node.Name), true) {
				remove = append(remove, node.Id)
			}
		}
	}

	if clean {
		message(typeInfo, 0, 0, common.Tr(infoEmpty))
	}

	return
}

func checkPkgrel(p *pkgbuild.PKGBUILD, edit bool) {
	n, i := p.FindLast(standard.PKGREL, info.SingleVar)
	if i < 0 {
		return
	}
	b, e := n.InnerPosition()
	v := n.Value
	if v == "1" {
		message(typeInfo, int(b.Line()), int(e.Line()), common.Tr(infoVarClean, standard.PKGREL))
		return
	}
	message(typeWarning, int(b.Line()), int(e.Line()), common.Tr(warnPkgrel))
	if edit && common.QuestionYN(common.Tr(questionPkgrel), false) {
		p.UpdateValue(n.Id, "1")
	}
}

func checkArch(p *pkgbuild.PKGBUILD, edit bool) {
	n, i := p.FindLast(standard.ARCH, info.ArrayVar)
	if i < 0 {
		return
	}
	b, e := n.InnerPosition()
	v := n.Values
	if len(v) == 1 && v[0] == "x86_64" {
		message(typeInfo, int(b.Line()), int(e.Line()), common.Tr(infoVarClean, standard.ARCH))
		return
	}
	message(typeWarning, int(b.Line()), int(e.Line()), common.Tr(warnArch))
	if edit && common.QuestionYN(common.Tr(questionArch), true) {
		p.UpdateValue(n.Id, "'x86_64'")
	}
}

func checkDepend(node *info.NodeInfo, pkgname string, exceptions collection.Set[string]) {
	errors := make(map[string]int)
	values := make([]string, len(node.Values))
	copy(values, node.Values)

	for i, v := range values {
		v = formatPackage(v)
		values[i] = v
		if v == pkgname {
			errors[v] = 1
		} else if !exceptions.Contains(v) && !isPackageInRepo(v) {
			errors[v] = 2
		}
	}
	name := node.Name
	begin, end := node.InnerPosition()
	if len(errors) == 0 {
		message(typeInfo, int(begin.Line()), int(end.Line()), common.Tr(infoVarClean, name))
		return
	}

	message(typeWarning, int(begin.Line()), int(end.Line()), common.Tr(warnDepends, name))
	for _, v := range values {
		if t, ok := errors[v]; ok {
			m := warnPackageNotInRepo
			if t == 1 {
				m = warnPackageIsName
			}
			fmt.Printf("  %s\n", common.Tr(m, v))
		}
	}
}

func checkDepends(p *pkgbuild.PKGBUILD, edit bool) {
	var hasDepend bool
	pkgname := p.GetValue(standard.PKGNAME)
	names := collection.NewSet(
		standard.CONFLICTS,
		standard.PROVIDES,
		standard.REPLACES,
		standard.DEPENDS,
		standard.MAKEDEPENDS,
		standard.OPTDEPENDS,
		standard.CHECKDEPENDS,
	)
	nameDepends := collection.NewSet(
		standard.DEPENDS,
		standard.MAKEDEPENDS,
	)
	exceptions := loadExceptions()

	for _, node := range p.NodeInfoList {
		name := node.Name
		if node.Type != info.ArrayVar || !names.Contains(name) {
			continue
		}
		hasDepend = hasDepend || nameDepends.Contains(name)
		checkDepend(node, pkgname, exceptions)
	}
	if !hasDepend {
		message(typeWarning, 0, 0, common.Tr(warnMissingDepends))
	}
}

func checkInstall(p *pkgbuild.PKGBUILD, edit bool) {
	node, index := p.FindLast(standard.INSTALL, info.SingleVar)
	if index < 0 {
		return
	}
	begin, end := node.InnerPosition()
	if common.FileExists(node.Value) {
		message(typeInfo, int(begin.Line()), int(end.Line()), common.Tr(infoVarClean, standard.INSTALL))
		fmt.Printf("  %s\n", common.Tr(commentInstall, urlHook))
		return
	}
	message(typeWarning, int(begin.Line()), int(end.Line()), common.Tr(warnInstall, node.Value))
	if edit && common.QuestionYN(common.Tr(questionInstall, node.Value), true) {
		newFile := common.Question(common.Tr(questionInstall2))
		if newFile == "" {
			p.Remove(node.Id)
		} else {
			p.UpdateValue(node.Id, newFile)
		}
	}
}
