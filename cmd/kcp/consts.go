package main

// Flags informations
const (
	appLongDescription = `Provides a tool to make use of KaOS Community Packages.

With this tool, you can search, get and install a package from KaOS Community Packages.`
	appDescription = "Tool in command-line for KaOS Community Packages"
	synopsis       = "(-h|-v|-u|(-l|-s <app>) [-fxNSIO]|-i <app> [-d]|-g <app>|-V <app>)"
	dHelp          = "Print this help"
	dVersion       = "Print version"
	dList          = "Display all packages of KCP"
	dUpdate        = "Refresh the local database"
	dSearch        = "Search packages in KCP and display them"
	dGet           = "Download needed files to build a package"
	dInstall       = "Install a package from KCP"
	dFast          = "On display action, don't print KCP version"
	dSort          = "On display action, sort packages by stars descending"
	dAsDeps        = "On install action, install as a dependence"
	dInstalled     = "On list action, display only installed packages"
	dComplete      = "On refreshing database action, force complete update"
	dForceUpdate   = "On display action, force refreshing local database"
	dOnlyName      = "On display action, display only the name of the package"
	dOnlystarred   = "On display action, display only packages with at least one star"
	dOnlyInstalled = "On display action, display only installed packages"
	dOnlyOutdated  = "On display action, display only outdated packages"
	dInformation   = "Display informations about a package"
	dValueName     = "<app>"
)

// Messages
const (
	errNoRoot                = "Don't launch this program as root!"
	errNoPackage             = "No package found"
	errNoPackageOrNeedUpdate = "No package found. Check if the database is updated."
	errOnlyOneInstance       = "Another instance of kcp is running!"
	errFailedCreateLocker    = "Failed to create locker file!"
	errInterrupt             = "Interrupt by user…"

	msgCloned      = "Package %s cloned in %s."
	msgEdit        = "Do you want to edit PKGBUILD?"
	msgEditInstall = "Do you want to edit %s?"
)
