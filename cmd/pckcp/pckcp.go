package main

func main() {
	initFlags()
	parseFlags()

	if *fHelp {
		flags.PrintHelp()
		return
	}
	if *fVersion {
		flags.PrintVersion()
		return
	}
	if *fGenerate {
		generate(*fClean, *fDebug)
		return
	}
	check(*fEdit, *fDebug)
}
