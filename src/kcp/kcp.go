package main

import (
	"fmt"
	"gettext"
	"kcpdb"
	"os"
	"os/signal"
	"parser/flag"
	"parser/pkgbuild"
	"path/filepath"
	"repo"
	"strings"
	"sysutil"
	"unicode/utf8"
)

//Flags informations
const (
	LONGDESCRIPTION = `Provides a tool to make the use of KaOS Community Packages.

With this tool, you can search, get and install a package from KaOS Community Packages.`
	APP_DESCRIPTION = "Tool in command-line for KaOS Community Packages"
	SYNOPSIS        = "(-h|-v|-u|(-l|-s <app>) [-fxNSIO]|-i <app> [-d]|-g <app>|-V <app>)"
	D_HELP          = "Print this help"
	D_VERSION       = "Print version"
	D_LIST          = "Display all packages of KCP"
	D_UPDATE        = "Refresh the local database"
	D_SEARCH        = "Search packages in KCP and display them"
	D_GET           = "Download needed files to build a package"
	D_INSTALL       = "Install a package from KCP"
	D_FAST          = "On display action, don't print KCP version"
	D_SORT          = "On display action, sort packages by stars descending"
	D_ASDEPS        = "On install action, install as a dependence"
	D_INSTALLED     = "On list action, display only installed packages"
	D_COMPLETE      = "On refreshing database action, force complete update"
	D_FORCEUPDATE   = "On display action, force refreshing local database"
	D_ONLYNAME      = "On display action, display only the name of the package"
	D_ONLYSTARRED   = "On display action, display only packages with at least one star"
	D_ONLYINSTALLED = "On display action, display only installed packages"
	D_ONLYOUTDATED  = "On display action, display only outdated packages"
	D_INFORMATION   = "Display informations about a package"
	VALUENAME       = "<app>"
)

var (
	flags                                                        *flag.Parser
	fHelp, fVersion, fList, fUpdate                              *bool
	fSearch, fGet, fInstall, fInfo                               *string
	fSorted, fOnlyName, fOnlyStar, fOnlyInstalled, fOnlyOutdated *bool
	fForceUpdate, fAsDepend, fDebug                              *bool
)

//Error messages
const (
	MSG_EDIT            = "Do you want to edit PKGBUILD?"
	MSG_EDIT_INSTALL    = "Do you want to edit %s?"
	MSG_ENTRIES_ADDED   = "%d entries added!"
	MSG_ENTRIES_DELETED = "%d entries deleted!"
	MSG_ENTRIES_UPDATED = "%d entries updated!"
	MSG_INTERRUPT       = "Interrupt by user..."
	MSG_NOPACKAGE       = "No package found"
	MSG_NOROOT          = "Don't launch this program as root!"
	MSG_ONLYONEINSTANCE = "Another instance of kcp is running!"
)

//Informations' labels
const (
	I_NAME        = "Name"
	I_VERSION     = "Version"
	I_DESCRIPTION = "Description"
	I_ARCH        = "Architecture"
	I_URL         = "URL"
	I_LICENSES    = "Licenses"
	I_PROVIDES    = "Provides"
	I_DEPENDS     = "Depends on"
	I_OPTDEPENDS  = "Optional Deps"
	I_CONFLICTS   = "Conflicts With"
	I_REPLACES    = "Replaces"
	I_INSTALL     = "Install Script"
	I_YES         = "Yes"
	I_NO          = "No"
)

var labels = []string{
	I_NAME,
	I_VERSION,
	I_DESCRIPTION,
	I_ARCH,
	I_URL,
	I_LICENSES,
	I_PROVIDES,
	I_DEPENDS,
	I_OPTDEPENDS,
	I_CONFLICTS,
	I_REPLACES,
	I_INSTALL,
}
var mlabel = map[string]string{
	I_NAME:        pkgbuild.PKGNAME,
	I_VERSION:     pkgbuild.PKGVER,
	I_DESCRIPTION: pkgbuild.PKGDESC,
	I_ARCH:        pkgbuild.ARCH,
	I_URL:         pkgbuild.URL,
	I_LICENSES:    pkgbuild.LICENSE,
	I_PROVIDES:    pkgbuild.PROVIDES,
	I_DEPENDS:     pkgbuild.DEPENDS,
	I_OPTDEPENDS:  pkgbuild.OPTDEPENDS,
	I_CONFLICTS:   pkgbuild.CONFLICTS,
	I_REPLACES:    pkgbuild.REPLACES,
	I_INSTALL:     pkgbuild.INSTALL,
}

//Helpers
var tr = gettext.Gettext

func trf(form string, e ...interface{}) string { return fmt.Sprintf(tr(form), e...) }

func checkUser() {
	if e := os.Geteuid(); e == 0 {
		sysutil.PrintError(tr(MSG_NOROOT))
		os.Exit(1)
	}
}
func dbPath() string { return filepath.Join(os.Getenv("HOME"), sysutil.KCP_DB) }
func displayCount(msgcnt string, c int) {
	if c > 0 {
		sysutil.PrintWarning(trf(msgcnt, c))
	}
}
func filters() (f []func(*kcpdb.Package) bool) {
	if *fOnlyStar {
		f = append(f, kcpdb.FilterStar)
	}
	if *fOnlyInstalled {
		f = append(f, kcpdb.FilterInstalled)
	}
	if *fOnlyOutdated {
		f = append(f, kcpdb.FilterOutdated)
	}
	return
}
func displayPackages(db kcpdb.Database) {
	switch {
	case *fOnlyName:
		for _, p := range db.Names() {
			fmt.Println(p)
		}
	case *fSorted:
		for _, p := range db.Sorted(kcpdb.SortByStar) {
			fmt.Println(p)
		}
	default:
		for _, p := range db.Sorted(kcpdb.SortByName) {
			fmt.Println(p)
		}
	}
	if len(db) == 0 {
		sysutil.PrintWarning("")
		sysutil.PrintWarning(tr(MSG_NOPACKAGE))
	}
}
func sizeLabel() (s int) {
	for _, e := range labels {
		if i := utf8.RuneCountInString(tr(e)); i > s {
			s = i
		}
	}
	return
}
func pathOf(file string) string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, file)
}

//Commands
func list() {
	var db kcpdb.Database
	var e error
	save := false
	if db, e = kcpdb.LoadBD(dbPath()); e != nil {
		db, _ = repo.List(false)
		save = true
	} else if *fForceUpdate {
		var rdb kcpdb.Database
		rdb, e = repo.List(false)
		if e == nil {
			db.Merge(rdb)
			save = true
		}
	}
	if e != nil {
		sysutil.PrintError(e)
		os.Exit(1)
	}
	displayPackages(db.Filter(filters()...))
	if save {
		if e = db.SaveBD(dbPath()); e != nil {
			sysutil.PrintError(e)
			os.Exit(1)
		}
	}
}
func update() {
	db, _ := kcpdb.LoadBD(dbPath())
	rdb, e := repo.List(false)
	if e == nil {
		u, a, d := db.Merge(rdb)
		displayCount(MSG_ENTRIES_UPDATED, u)
		displayCount(MSG_ENTRIES_ADDED, a)
		displayCount(MSG_ENTRIES_DELETED, d)
		e = db.SaveBD(dbPath())
	}
	if e != nil {
		sysutil.PrintError(e)
		os.Exit(1)
	}
}
func search() {
	var db kcpdb.Database
	var e error
	save := false
	if db, e = kcpdb.LoadBD(dbPath()); e != nil {
		db, _ = repo.List(false)
		save = true
	} else if *fForceUpdate {
		var rdb kcpdb.Database
		rdb, e = repo.List(false)
		if e == nil {
			db.Merge(rdb)
			save = true
		}
	}
	if e != nil {
		sysutil.PrintError(e)
		os.Exit(1)
	}
	filt := append(filters(), kcpdb.FilterName(*fSearch), kcpdb.FilterNameOrDescription(*fSearch))
	displayPackages(db.Filter(filt...))
	if save {
		if e = db.SaveBD(dbPath()); e != nil {
			sysutil.PrintError(e)
			os.Exit(1)
		}
	}
}
func info() {
	var e error
	var b []byte
	if b, e = repo.Pkgbuild(*fInfo); e == nil {
		var p *pkgbuild.Pkgbuild
		if p, e = pkgbuild.ParseBytes(b); e == nil {
			s := sizeLabel() + 1
			for _, e := range labels {
				l, k, v := tr(e), mlabel[e], ""
				l += strings.Repeat(" ", s-utf8.RuneCountInString(l))
				switch k {
				case pkgbuild.INSTALL:
					v = p.Variable(k)
					if v == "" {
						v = I_NO
					} else {
						v = I_YES
					}
				case pkgbuild.PKGVER:
					v = p.Version()
				default:
					v = p.Variable(k)
				}
				if v == "" {
					v = "--"
				}
				fmt.Printf("\033[1m%s:\033[m %s\n", l, v)
			}
		}
	}
	if e != nil {
		sysutil.PrintError(e)
		os.Exit(1)
	}
}
func get() {
	if e := repo.Clone(*fGet); e != nil {
		sysutil.PrintError(e)
		os.Exit(1)
	}
}
func install() {
	var e error
	tmpdir := os.TempDir()
	os.Chdir(tmpdir)
	lck, wdir := pathOf(sysutil.KCP_LOCK), pathOf(*fInstall)
	if _, e = os.Open(lck); e == nil {
		sysutil.PrintError(tr(MSG_ONLYONEINSTANCE))
		os.Exit(1)
	}
	os.Create(lck)
	rem := func() {
		os.RemoveAll(wdir)
		os.Remove(lck)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		rem()
		sysutil.PrintError(tr(MSG_INTERRUPT))
		os.Exit(1)
	}()
	defer rem()
	e = repo.Clone(*fInstall)
	if e != nil {
		rem()
		sysutil.PrintError(e)
		os.Exit(1)
	}
	os.Chdir(wdir)
	if sysutil.QuestionYN(tr(MSG_EDIT), true) {
		if e := sysutil.EditFile("PKGBUILD"); e != nil {
			rem()
			sysutil.PrintError(e)
			os.Exit(1)
		}
	}
	m, _ := filepath.Glob("*.install")
	for _, i := range m {
		if sysutil.QuestionYN(trf(MSG_EDIT_INSTALL, i), false) {
			if e := sysutil.EditFile(i); e != nil {
				rem()
				sysutil.PrintError(e)
				os.Exit(1)
			}
		}
	}
	if *fAsDepend {
		e = sysutil.LaunchCommand("makepkg", "-si", "--asdeps")
	} else {
		e = sysutil.LaunchCommand("makepkg", "-si")
	}
	if e != nil {
		rem()
		sysutil.PrintError(e)
		os.Exit(1)
	}
}

func init() {
	// Init the locales
	os.Setenv("LANGUAGE", os.Getenv("LC_MESSAGES"))
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.BindTextdomain("kcp", sysutil.LOCALE_DIR)
	gettext.Textdomain("kcp")

	//Init the flags
	flags = flag.NewParser(tr(APP_DESCRIPTION), sysutil.VERSION)
	flags.Set(flag.SYNOPSIS, tr(SYNOPSIS))
	flags.Set(flag.LONGDESCRIPTION, tr(LONGDESCRIPTION))
	flags.Set(flag.AUTHOR, sysutil.AUTHOR)

	fHelp, _ = flags.Bool("-h", "--help", tr(D_HELP))
	fVersion, _ = flags.Bool("-v", "--version", tr(D_VERSION))
	fList, _ = flags.Bool("-l", "--list", tr(D_LIST))
	fUpdate, _ = flags.Bool("-u", "--update-database", tr(D_UPDATE))
	fSearch, _ = flags.String("-s", "--search", tr(D_SEARCH), VALUENAME, "")
	fGet, _ = flags.String("-g", "--get", tr(D_GET), VALUENAME, "")
	fInstall, _ = flags.String("-i", "--install", tr(D_INSTALL), VALUENAME, "")
	fSorted, _ = flags.Bool("-x", "--sort", tr(D_SORT))
	fForceUpdate, _ = flags.Bool("-f", "--force-update", tr(D_FORCEUPDATE))
	fOnlyName, _ = flags.Bool("-N", "--only-name", tr(D_ONLYNAME))
	fOnlyStar, _ = flags.Bool("-S", "--only-starred", tr(D_ONLYSTARRED))
	fOnlyInstalled, _ = flags.Bool("-I", "--only-installed", tr(D_ONLYINSTALLED))
	fOnlyOutdated, _ = flags.Bool("-O", "--only-outdated", tr(D_ONLYOUTDATED))
	fAsDepend, _ = flags.Bool("-d", "--asdeps", tr(D_ASDEPS))
	fInfo, _ = flags.String("-V", "--information", tr(D_INFORMATION), VALUENAME, "")
	fDebug, _ = flags.Bool("", "--debug", "")

	flags.Group("-h", "-v", "-l", "-s", "-g", "-i", "-u", "--information")
	flags.Require("--sort", "-l", "-s")
	flags.Require("--force-update", "-l", "-s")
	flags.Require("--only-name", "-l", "-s")
	flags.Require("--only-starred", "-l", "-s")
	flags.Require("--only-installed", "-l", "-s")
	flags.Require("--only-outdated", "-l", "-s")
	flags.Require("--asdeps", "-i")
	flags.GetFlag("--debug").Set(flag.HIDDEN, true)
}

func main() {
	checkUser()
	e := flags.Parse(os.Args)
	switch {
	case e != nil:
		sysutil.PrintError(e)
		fmt.Println()
		flags.PrintHelp()
		os.Exit(1)
	case *fHelp:
		flags.PrintHelp()
	case *fVersion:
		flags.PrintVersion()
	case *fUpdate:
		update()
	case *fList:
		list()
	case *fSearch != "":
		search()
	case *fInfo != "":
		info()
	case *fGet != "":
		get()
	case *fInstall != "":
		install()
	default:
		flags.PrintHelp()
		os.Exit(1) // Should not happen
	}
}
