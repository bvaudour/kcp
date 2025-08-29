package main

import (
	"fmt"
	"io"
	"os"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/pkgbuild"
	pformat "codeberg.org/bvaudour/kcp/pkgbuild/format"
)

func generate(clean, debug bool, output string) {
	fp := common.JoinIfRelative(common.UserBaseDir, common.Config.Get("pckcp.protoFile"))
	if !common.FileExists(fp) {
		fp = common.JoinIfRelative(common.ConfigBaseDir, common.Config.Get("pckcp.protoFile"))
		if !common.FileExists(fp) {
			common.PrintError(common.Tr(errFileNotExist, fp))
			os.Exit(1)
		}
	}
	proto, err := os.Open(fp)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	defer proto.Close()
	p, err := pkgbuild.Decode(proto)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	if clean {
		p.Format(pformat.NewFormater(pformat.OptionRemoveInnerComments, pformat.OptionRemoveOuterComments))
	}
	if debug {
		p.Debug(os.Stderr)
		return
	}

	if output == "" {
		output = "PKGBUILD"
	}
	dest, err := os.Create(output)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	defer dest.Close()
	if err := p.Encode(dest); err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
}

func check(edit, debug bool, output string) {
	f, err := os.Open("PKGBUILD")
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	defer f.Close()

	p, err := pkgbuild.Decode(f)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}

	// Checkers
	var options []pformat.FormatOption
	options = append(options, checkHeader(p, edit)...)
	options = append(options, checkDuplicates(p, edit)...)
	newNodes := checkMissingVars(p, edit)
	checkMissingFuncs(p, edit)
	options = append(options, checkInfoTypes(p, edit)...)
	rempve := checkEmpty(p, edit)
	checkPkgrel(p, edit)
	checkArch(p, edit)
	checkDepends(p, edit)
	checkInstall(p, edit)

	if edit {
		p.Add(newNodes...)
		p.Remove(rempve...)
		if common.QuestionYN(common.Tr(questionFormat), true) {
			options = append(options,
				pformat.OptionRemoveOuterComments,
				pformat.OptionRemoveInnerComments,
				pformat.OptionFormatWords,
				pformat.OptionFormatBlank,
				pformat.OptionKeepFirstBlank,
				pformat.OptionIndentWithTabs,
				pformat.OptionReorder,
			)
		}
		p.Format(pformat.NewFormater(options...))
	}

	if debug {
		fmt.Fprintln(os.Stderr)
		p.Debug(os.Stderr)
		return
	}

	if !edit {
		return
	}

	if output == "" {
		output = fmt.Sprintf("PKGBUILD%s", common.Config.Get("pckcp.suffixNewPKGBUILD"))
	}

	dest, err := os.Create(output)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	defer dest.Close()

	if err := p.Encode(dest); err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	common.PrintWarning(common.Tr(warnSaved, output))
}

func format(debug bool, output string) {
	f, err := os.Open("PKGBUILD")
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	defer f.Close()

	p, err := pkgbuild.Decode(f)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}

	p.Format(pformat.NewFormater(
		pformat.OptionRemoveInnerComments,
		pformat.OptionRemoveOuterComments,
		pformat.OptionRemoveDuplicates,
		pformat.OptionFormatWords,
		pformat.OptionIndentWithTabs,
		pformat.OptionReorder,
		pformat.OptionFormatBlank,
		pformat.OptionKeepFirstBlank,
	))

	if debug {
		p.Debug(os.Stderr)
		return
	}

	var w io.Writer
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			common.PrintError(err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	} else {
		w = os.Stdout
	}

	if err := p.Encode(w); err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	if output != "" {
		common.PrintWarning(common.Tr(warnSaved, output))
	}
}
