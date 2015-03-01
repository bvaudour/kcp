package main

import (
	"fmt"
	"gettext"
	"kcp/api"
	"os"
	"parser/pargs"
)

// Filter methods
var filtInstalled = func(p *api.Package) bool { return p.LocalVersion != "" }
var filtOutdated = func(p *api.Package) bool { return p.LocalVersion != p.KcpVersion }
var filtStarred = func(p *api.Package) bool { return p.Stars > 0 }

func filter(c api.PCollection, r_lst bool, filt ...func(*api.Package) bool) api.PCollection {
	out := api.EmptyPCollection(r_lst)
	for _, p := range c.List() {
		ok := true
		for _, f := range filt {
			if ok = f(p); !ok {
				break
			}
		}
		if ok {
			out.Add(p)
		}
	}
	return out
}

func filterlist(c api.PCollection, filt ...func(*api.Package) bool) api.PList {
	return filter(c, true, filt...).List()
}

func filtermap(c api.PCollection, filt ...func(*api.Package) bool) api.PMap {
	return filter(c, false, filt...).Map()
}

// Database methods
func load() (api.PMap, bool) {
	m, e := api.LoadMapDB()
	return m, e == nil
}

func merge(db api.PMap, c api.PCollection) (updated int, added int, deleted int) {
	for _, p := range c.List() {
		p_db, ok := db[p.Name]
		if !ok {
			db.Add(p)
			added++
		} else {
			if p.Description != "" && p.Description != p_db.Description {
				ok = false
				p_db.Description = p.Description
			}
			if p.Stars != p_db.Stars {
				ok = false
				p_db.Stars = p.Stars
			}
			//if p.LocalVersion != "" && p.LocalVersion != p_db.LocalVersion {
			if p.LocalVersion != p_db.LocalVersion {
				ok = false
				p_db.LocalVersion = p.LocalVersion
			}
			if p.KcpVersion != "" && p.KcpVersion != p_db.KcpVersion {
				ok = false
				p_db.KcpVersion = p.KcpVersion
			}
			if !ok {
				updated++
			}
		}
	}
	newdb := c.Map()
	for n, _ := range db {
		if _, ok := newdb[n]; !ok {
			delete(db, n)
			deleted++
		}
	}
	return
}

func updatelocalv(db api.PMap, finish chan bool) {
	updated := api.LocalMapAll()
	for n, p := range db {
		if pu, ok := updated[n]; ok {
			p.LocalVersion = pu.LocalVersion
		} else {
			p.LocalVersion = ""
		}
	}
	finish <- true
}

func updatekcpv(db api.PMap, onlyinstalled bool, finish chan bool) {
	f := make(chan bool, len(db))
	for _, p := range db {
		go func(p *api.Package, f chan bool) {
			if !onlyinstalled || p.LocalVersion != "" {
				p.KcpVersion = api.KcpVersion(p.Name)
			}
			f <- true
		}(p, f)
	}
	for range f {
	}
	finish <- true
}

func updateversions(db api.PMap, onlyinstalled bool) <-chan bool {
	finish := make(chan bool)
	go func() {
		f := make(chan bool, 2)
		go updatelocalv(db, f)
		go updatekcpv(db, onlyinstalled, f)
		<-f
		<-f
		close(f)
		finish <- true
		close(finish)
	}()
	return finish
}

// Print method
func printerror(v interface{}) {
	fmt.Printf("\033[1;31m%v\033[m\n", v)
}

func printwarning(v interface{}) {
	fmt.Printf("\033[1;33m%v\033[m\n", v)
}

func nameprint(c api.PCollection, sorted bool) {
	l := c.List()
	if sorted {
		l = l.Sorted()
	}
	for _, p := range l {
		fmt.Println(p.Name)
	}
}

func allprint(c api.PCollection, sorted bool) {
	l := c.List()
	if sorted {
		l = l.Sorted()
	}
	for _, p := range l {
		fmt.Println(p)
	}
	if len(l) == 0 {
		fmt.Println()
		printwarning(t(api.MSG_NOPACKAGE))
	}
}

func packagesprint(c api.PCollection, sorted, onlyname bool) {
	if onlyname {
		nameprint(c, sorted)
	} else {
		allprint(c, sorted)
	}
}

// Global variables
var argparser *pargs.Parser
var flag_help, flag_version, flag_list, flag_update *bool
var flag_search, flag_get, flag_install *string
var flag_complete, flag_sorted, flag_forceupdate *bool
var flag_onlyname, flag_onlystarred, flag_onlyinstalled, flag_onlyoutdated *bool
var flag_asdeps *bool
var flag_debug *bool
var flag_fast *bool

// Parser's descriptions
const (
	LONGDESCRIPTION = `Provides a tool to make the use of KaOS Community Packages.

With this tool, you can search, get and install a package from KaOS Community Packages.`
	VERSION         = "0.52"
	AUTHOR          = "B. VAUDOUR"
	APP_DESCRIPTION = "Tool in command-line for KaOS Community Packages"
	SYNOPSIS        = "[OPTIONS] [APP]"
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
	VALUENAME       = "<app>"
)

func t(arg string) string { return api.Translate(arg) }

// Actions
func help() {
	argparser.PrintHelp()
}

func version() {
	argparser.PrintVersion()
}

func list() {
	var m api.PMap
	var e error
	if *flag_forceupdate {
		//if *flag_complete {
		m, e = api.KcpMapAllWithVersions(*flag_debug)
		//} else {
		//	m, e = api.KcpMapAll(*flag_debug)
		//}
		//updatelocalv(m)
		//updatekcpv(m, !*flag_complete)
	} else if db, ok := load(); ok {
		m = db
	} else {
		m, e = api.KcpMapAllWithVersions(*flag_debug)
		//updatelocalv(m)
		//updatekcpv(m, true)
	}
	if e != nil {
		printerror(e)
		os.Exit(1)
	} else {
		api.SaveDB(m)
		filters := make([]func(*api.Package) bool, 0)
		if *flag_onlyinstalled {
			filters = append(filters, filtInstalled)
		} else if *flag_onlyoutdated {
			filters = append(filters, filtInstalled, filtOutdated)
		}
		if *flag_onlystarred {
			filters = append(filters, filtStarred)
		}
		if len(filters) > 0 {
			m = filtermap(m, filters...)
		}
		packagesprint(m, *flag_sorted, *flag_onlyname)
	}
}

func update() {
	mlocal, _ := api.LoadMapDB()
	var mkcp api.PMap
	var e error
	//if *flag_complete {
	mkcp, e = api.KcpMapAllWithVersions(*flag_debug)
	//} else {
	//  mkcp, e = api.KcpMapAll(*flag_debug)
	//}
	if e != nil {
		printerror(e)
		os.Exit(1)
	}
	//u := updateversions(mkcp, !*flag_complete)
	//<-u
	//updatelocalv(mkcp)
	//updatekcpv(mkcp, !*flag_complete)
	u, a, d := merge(mlocal, mkcp)
	if e = api.SaveDB(mlocal); e != nil {
		printerror(e)
		os.Exit(1)
	} else {
		fmt.Printf(t(api.MSG_ENTRIES_UPDATED)+"\n", u)
		fmt.Printf(t(api.MSG_ENTRIES_ADDED)+"\n", a)
		fmt.Printf(t(api.MSG_ENTRIES_DELETED)+"\n", d)
	}
}

func search(app string) {
	l, e := api.KcpListSearch(app, *flag_debug)
	if e != nil {
		printerror(e)
		os.Exit(1)
	} else {
		filters := make([]func(*api.Package) bool, 0)
		if *flag_onlyinstalled {
			filters = append(filters, filtInstalled)
		} else if *flag_onlyoutdated {
			filters = append(filters, filtInstalled, filtOutdated)
		}
		if *flag_onlystarred {
			filters = append(filters, filtStarred)
		}
		if len(filters) > 0 {
			l = filterlist(&l, filters...)
		}
		for _, p := range l {
			p.KcpVersion = api.KcpVersion(p.Name)
		}
		packagesprint(&l, *flag_sorted, *flag_onlyname)
	}
}

func get(app string) {
	e := api.Get(app)
	if e != nil {
		printerror(e)
		os.Exit(1)
	}
}

func install(app string, asdeps bool) {
	e := api.Install(app, asdeps)
	if e != nil {
		printerror(e)
		os.Exit(1)
	} else if db, ok := load(); ok {
		f := make(chan bool)
		go updatelocalv(db, f)
		<-f
		api.SaveDB(db)
	}
}

func checkUser() {
	if e := os.Geteuid(); e == 0 {
		printerror(t(api.MSG_NOROOT))
		os.Exit(1)
	}
}

func init() {
	// Init the locales
	os.Setenv("LANGUAGE", os.Getenv("LC_MESSAGES"))
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.BindTextdomain("kcp", api.LOCALE_DIR)
	gettext.Textdomain("kcp")

	// Init the args parser
	argparser = pargs.New(t(APP_DESCRIPTION), VERSION)
	argparser.Set(pargs.AUTHOR, AUTHOR)
	argparser.Set(pargs.SYNOPSIS, t(SYNOPSIS))
	argparser.Set(pargs.LONGDESCRIPTION, t(LONGDESCRIPTION))
	flag_help, _ = argparser.Bool("-h", "--help", t(D_HELP))
	flag_version, _ = argparser.Bool("-v", "--version", t(D_VERSION))
	flag_list, _ = argparser.Bool("-l", "--list", t(D_LIST))
	flag_update, _ = argparser.Bool("-u", "--update-database", t(D_UPDATE))
	flag_search, _ = argparser.String("-s", "--search", t(D_SEARCH), VALUENAME, "")
	flag_get, _ = argparser.String("-g", "--get", t(D_GET), VALUENAME, "")
	flag_install, _ = argparser.String("-i", "--install", t(D_INSTALL), VALUENAME, "")
	flag_complete, _ = argparser.Bool("-c", "--complete", t(D_COMPLETE))
	flag_sorted, _ = argparser.Bool("-x", "--sort", t(D_SORT))
	flag_forceupdate, _ = argparser.Bool("-f", "--force-update", t(D_FORCEUPDATE))
	flag_onlyname, _ = argparser.Bool("-N", "--only-name", t(D_ONLYNAME))
	flag_onlystarred, _ = argparser.Bool("-S", "--only-starred", t(D_ONLYSTARRED))
	flag_onlyinstalled, _ = argparser.Bool("-I", "--only-installed", t(D_ONLYINSTALLED))
	flag_onlyoutdated, _ = argparser.Bool("-O", "--only-outdated", t(D_ONLYOUTDATED))
	flag_asdeps, _ = argparser.Bool("-d", "--asdeps", t(D_ASDEPS))
	flag_debug, _ = argparser.Bool("-D", "--debug", "debug mode")

	// Init flags groups/requirements
	argparser.Group("-h", "-v", "-l", "-s", "-g", "-i", "-u")
	argparser.Require("--complete", "-u", "-f")
	argparser.Require("--sort", "-l", "-s")
	argparser.Require("--force-update", "-l")
	argparser.Require("--only-name", "-l", "-s")
	argparser.Require("--only-starred", "-l", "-s")
	argparser.Require("--only-installed", "-l", "-s")
	argparser.Require("--only-outdated", "-l", "-s")
	argparser.Require("--asdeps", "-i")
	argparser.GetFlag("--debug").Set(pargs.HIDDEN, true)
	// To ensure compatibility with completion
	//flag_fast, _ = argparser.Bool("", "--fast", "")
	//argparser.GetFlag("--fast").Set(pargs.HIDDEN, true)
	//argparser.GetFlag("--complete").Set(pargs.HIDDEN, true)
}

func main() {
	checkUser()
	e := argparser.Parse(os.Args)
	switch {
	case e != nil:
		printerror(e)
		fmt.Println()
		help()
		os.Exit(1)
	case *flag_help:
		help()
	case *flag_version:
		version()
	case *flag_list:
		//if *flag_fast {
		//	*flag_onlyname = true
		//}
		list()
	case *flag_update:
		update()
	case *flag_search != "":
		search(*flag_search)
	case *flag_get != "":
		get(*flag_get)
	case *flag_install != "":
		install(*flag_install, *flag_asdeps)
	default:
		help()
		os.Exit(1) // Should not happen
	}
}
