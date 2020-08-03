package main

func main() {
	checkUser()
	initFlags()
	parseFlags()
	switch {
	case *fHelp:
		flags.PrintHelp()
	case *fVersion:
		flags.PrintVersion()
	case *fUpdate:
		update(*fDebug)
	case *fList:
		list(*fDebug, *fForceUpdate, *fOnlyName, *fOnlyStar, *fOnlyInstalled, *fOnlyOutdated, *fSorted)
	case *fSearch != "":
		search(*fDebug, *fForceUpdate, *fOnlyName, *fOnlyStar, *fOnlyInstalled, *fOnlyOutdated, *fSorted, *fSearch)
	case *fInfo != "":
		info(*fDebug, *fInfo)
	case *fGet != "":
		get(*fDebug, *fGet)
	case *fInstall != "":
		install(*fDebug, *fInstall, *fAsDepend)
	}
}
