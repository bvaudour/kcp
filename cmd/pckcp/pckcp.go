package main

func main() {
	initFlags()
	parseFlags()

	switch {
	case *fHelp:
		flags.PrintHelp()
	case *fVersion:
		flags.PrintVersion()
	case *fGenerate:
		generate(*fClean, *fDebug, *fOutput)
	case *fFormat:
		format(*fDebug, *fOutput)
	default:
		check(*fEdit, *fDebug, *fOutput)
	}
}
