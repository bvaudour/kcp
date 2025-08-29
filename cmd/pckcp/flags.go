package main

import (
	"os"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/flag"
)

var (
	flags                                                      *flag.Parser
	fHelp, fVersion, fEdit, fDebug, fGenerate, fClean, fFormat *bool
	fOutput                                                    *string
)

func initFlags() {
	flags = flag.NewParser(common.Tr(appDescription), common.Version)
	flags.Set(flag.Synopsis, common.Tr(synopsis))
	flags.Set(flag.LongDescription, common.Tr(appLongDescription))

	fHelp, _ = flags.Bool("-h", "--help", common.Tr(help))
	fVersion, _ = flags.Bool("-v", "--version", common.Tr(version))
	fEdit, _ = flags.Bool("-e", "--edit", common.Tr(interactiveEdit))
	fGenerate, _ = flags.Bool("-g", "--generate", common.Tr(generatePrototype))
	fFormat, _ = flags.Bool("-f", "--format", common.Tr(formatFile))
	fOutput, _ = flags.String("-o", "--output", common.Tr(formatedOutput), common.Tr(dFileName), "")

	fClean, _ = flags.Bool("-c", "--clean", common.Tr(cleanUseless))
	flags.Require("-c", "-g")

	fDebug, _ = flags.Bool("-d", "--debug", "")
	flags.GetFlag("--debug").Set(flag.Hidden, true)
	flags.Group("-h", "-e", "-v", "-g", "-f")
}

func parseFlags() {
	if err := flags.Parse(os.Args); err != nil {
		common.PrintError(err)
		flags.PrintHelp()
		os.Exit(1)
	}
}
