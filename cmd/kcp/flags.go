package main

import (
	"os"

	. "github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/flag"
)

var (
	flags                                                        *flag.Parser
	fHelp, fVersion, fList, fUpdate                              *bool
	fSearch, fGet, fInstall, fInfo                               *string
	fSorted, fOnlyName, fOnlyStar, fOnlyInstalled, fOnlyOutdated *bool
	fForceUpdate, fAsDepend, fDebug                              *bool
)

func initFlags() {
	flags = flag.NewParser(Tr(appDescription), Version)
	flags.Set(flag.Synopsis, Tr(synopsis))
	flags.Set(flag.LongDescription, Tr(appLongDescription))

	fHelp, _ = flags.Bool("-h", "--help", Tr(dHelp))
	fVersion, _ = flags.Bool("-v", "--version", Tr(dVersion))
	fList, _ = flags.Bool("-l", "--list", Tr(dList))
	fUpdate, _ = flags.Bool("-u", "--update-database", Tr(dUpdate))
	fSearch, _ = flags.String("-s", "--search", Tr(dSearch), Tr(dValueName), "")
	fGet, _ = flags.String("-g", "--get", Tr(dGet), Tr(dValueName), "")
	fInstall, _ = flags.String("-i", "--install", Tr(dInstall), Tr(dValueName), "")
	fSorted, _ = flags.Bool("-x", "--sort", Tr(dSort))
	fForceUpdate, _ = flags.Bool("-f", "--force-update", Tr(dForceUpdate))
	fOnlyName, _ = flags.Bool("-N", "--only-name", Tr(dOnlyName))
	fOnlyStar, _ = flags.Bool("-S", "--only-starred", Tr(dOnlystarred))
	fOnlyInstalled, _ = flags.Bool("-I", "--only-installed", Tr(dOnlyInstalled))
	fOnlyOutdated, _ = flags.Bool("-O", "--only-outdated", Tr(dOnlyOutdated))
	fAsDepend, _ = flags.Bool("-d", "--asdeps", Tr(dAsDeps))
	fInfo, _ = flags.String("-V", "--information", Tr(dInformation), Tr(dValueName), "")
	fDebug, _ = flags.Bool("", "--debug", "")

	flags.Group("-h", "-v", "-l", "-s", "-g", "-i", "-u", "--information")
	flags.Require("--sort", "-l", "-s")
	flags.Require("--force-update", "-l", "-s")
	flags.Require("--only-name", "-l", "-s")
	flags.Require("--only-starred", "-l", "-s")
	flags.Require("--only-installed", "-l", "-s")
	flags.Require("--only-outdated", "-l", "-s")
	flags.Require("--asdeps", "-i")
	flags.GetFlag("--debug").Set(flag.Hidden, true)
}

func parseFlags() {
	if err := flags.Parse(os.Args); err != nil {
		PrintError(err)
		flags.PrintHelp()
		os.Exit(1)
	}
}

func checkUser() {
	if e := os.Geteuid(); e == 0 {
		PrintError(Tr(errNoRoot))
		os.Exit(1)
	}
}
