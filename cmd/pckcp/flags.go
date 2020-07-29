package main

import (
	"fmt"
	"os"

	. "github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/flag"
)

var (
	flags                                             *flag.Parser
	fHelp, fVersion, fEdit, fDebug, fGenerate, fClean *bool
)

func initFlags() {
	flags = flag.NewParser(Tr(appDescription), Version)
	flags.Set(flag.Synopsis, Tr(synopsis))
	flags.Set(flag.LongDescription, Tr(appLongDescription))

	fHelp, _ = flags.Bool("-h", "--help", Tr(help))
	fVersion, _ = flags.Bool("-v", "--version", Tr(version))
	fEdit, _ = flags.Bool("-e", "--edit", Tr(interactiveEdit))
	fGenerate, _ = flags.Bool("-g", "--generate", Tr(generatePrototype))

	fClean, _ = flags.Bool("-c", "--clean", Tr(cleanUseless))
	flags.Require("-c", "-g")

	fDebug, _ = flags.Bool("-d", "--debug", "")
	flags.GetFlag("--debug").Set(flag.Hidden, true)
	flags.Group("-h", "-e", "-v", "-g")
}

func parseFlags() {
	if err := flags.Parse(os.Args); err != nil {
		PrintError(err)
		fmt.Fprintln(os.Stderr)
		flags.PrintHelp()
		os.Exit(1)
	}
}
