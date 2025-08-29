package main

import (
	"os"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/flag"
)

var (
	flags                                                        *flag.Parser
	fHelp, fVersion, fList, fUpdate                              *bool
	fSearch, fGet, fInstall, fInfo                               *string
	fSorted, fOnlyName, fOnlyStar, fOnlyInstalled, fOnlyOutdated *bool
	fForceUpdate, fAsDepend, fDebug                              *bool
)

func initFlags() {
	flags = flag.NewParser(common.Tr(appDescription), common.Version)
	flags.Set(flag.Synopsis, common.Tr(synopsis))
	flags.Set(flag.LongDescription, common.Tr(appLongDescription))

	fHelp, _ = flags.Bool("-h", "--help", common.Tr(dHelp))
	fVersion, _ = flags.Bool("-v", "--version", common.Tr(dVersion))
	fList, _ = flags.Bool("-l", "--list", common.Tr(dList))
	fUpdate, _ = flags.Bool("-u", "--update-database", common.Tr(dUpdate))
	fSearch, _ = flags.String("-s", "--search", common.Tr(dSearch), common.Tr(dValueName), "")
	fGet, _ = flags.String("-g", "--get", common.Tr(dGet), common.Tr(dValueName), "")
	fInstall, _ = flags.String("-i", "--install", common.Tr(dInstall), common.Tr(dValueName), "")
	fSorted, _ = flags.Bool("-x", "--sort", common.Tr(dSort))
	fForceUpdate, _ = flags.Bool("-f", "--force-update", common.Tr(dForceUpdate))
	fOnlyName, _ = flags.Bool("-N", "--only-name", common.Tr(dOnlyName))
	fOnlyStar, _ = flags.Bool("-S", "--only-starred", common.Tr(dOnlystarred))
	fOnlyInstalled, _ = flags.Bool("-I", "--only-installed", common.Tr(dOnlyInstalled))
	fOnlyOutdated, _ = flags.Bool("-O", "--only-outdated", common.Tr(dOnlyOutdated))
	fAsDepend, _ = flags.Bool("-d", "--asdeps", common.Tr(dAsDeps))
	fInfo, _ = flags.String("-V", "--information", common.Tr(dInformation), common.Tr(dValueName), "")
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
		common.PrintError(err)
		flags.PrintHelp()
		os.Exit(1)
	}
}

func checkUser() {
	if e := os.Geteuid(); e == 0 {
		common.PrintError(common.Tr(errNoRoot))
		os.Exit(1)
	}
}
