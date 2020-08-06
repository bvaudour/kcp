package main

import (
	"fmt"
	"os"

	. "github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/pkgbuild"
)

func generate(clean, debug bool) {
	fp := JoinIfRelative(UserBaseDir, Config.Get("pckcp.protoFile"))
	if !FileExists(fp) {
		fp = JoinIfRelative(ConfigBaseDir, Config.Get("pckcp.protoFile"))
		if !FileExists(fp) {
			PrintError(Tr(errFileNotExist, fp))
			os.Exit(1)
		}
	}
	proto, err := os.Open(fp)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	defer proto.Close()
	p, err := pkgbuild.Decode(proto)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	if clean {
		p.Format()
	}
	if debug {
		fmt.Print(p.String())
		return
	}
	dest, err := os.Create("PKGBUILD")
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	defer dest.Close()
	if _, err := p.Encode(dest); err != nil {
		PrintError(err)
		os.Exit(1)
	}
}

func check(edit, debug bool) {
	f, err := os.Open("PKGBUILD")
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	defer f.Close()
	p, err := pkgbuild.Decode(f)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}

	// Checkers
	checkHeader(p, edit)
	checkDuplicates(p, edit)
	checkMissingFuncs(p, edit)
	checkMissingFuncs(p, edit)
	checkInfoTypes(p, edit)
	checkEmpty(p, edit)
	checkPkgrel(p, edit)
	checkArch(p, edit)
	checkDepends(p, edit)
	checkInstalls(p, edit)
	if edit && QuestionYN(Tr(questionFormat), true) {
		p.Format()
	}

	if debug {
		fmt.Println()
		fmt.Print(p.String())
		return
	}
	if !edit {
		return
	}
	destname := fmt.Sprintf("PKGBUILD%s", Config.Get("pckcp.suffixNewPKGBUILD"))
	dest, err := os.Create(destname)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	defer dest.Close()
	if _, err := p.Encode(dest); err != nil {
		PrintError(err)
		os.Exit(1)
	}
	PrintWarning(Tr(warnSaved, destname))
}
