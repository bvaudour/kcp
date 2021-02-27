package main

import (
	"fmt"
	"path"

	"github.com/bvaudour/kcp/common"
	"github.com/leonelquinteros/gotext"
)

func main() {
	checkUser()
	initFlags()
	parseFlags()
	switch {
	case *fHelp:
		if *fDebug {
			b, d, l := gotext.GetLibrary(), gotext.GetDomain(), gotext.GetLanguage()
			f := path.Join(d, l, "LC_MESSAGES", l+".mo")
			fmt.Println("Debug locale configuration:")
			fmt.Println("- Base path:", b)
			fmt.Println("- Domain:", d)
			fmt.Println("- Language used:", l)
			fmt.Println("- Mo file:")
			fmt.Println(" - Mo file exists:", common.FileExists(f))
		}
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
