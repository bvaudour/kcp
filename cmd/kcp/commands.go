package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bvaudour/kcp/color"
	. "github.com/bvaudour/kcp/common"
	database "github.com/bvaudour/kcp/database2"
	"github.com/leonelquinteros/gotext"
)

func getIgnore() []string {
	return strings.Fields(Config.Get("kcp.ignore"))
}

func getDbPath() string {
	return JoinIfRelative(UserBaseDir, Config.Get("kcp.dbFile"))
}

func useSsh() bool {
	return Config.Get("kcp.cloneMethod") == "ssh"
}

func getDb() (db *database.Database, err error) {
	fpath, ignore := getDbPath(), getIgnore()
	return database.Load(fpath, ignore...)
}

func updateDb(db *database.Database, debug bool) (counters database.Counter, err error) {
	return db.Update(Organization, debug, User, Password)
}

func loadDb(debug, forceUpdate bool) *database.Database {
	db, err := getDb()
	if debug {
		fmt.Fprintln(os.Stderr, "Trying to open db")
	}
	if err != nil || forceUpdate {
		if debug {
			fmt.Fprintln(os.Stderr, "Trying to update db")
		}
		updateDb(db, debug)
	}
	return db
}

func saveDb(db *database.Database) error {
	return database.Save(getDbPath(), db)
}

func filter(debug, forceUpdate, onlyName bool, f []database.FilterFunc, s []database.SorterFunc) {
	db := loadDb(debug, forceUpdate)
	if forceUpdate {
		saveDb(db)
	}
	l := db.Filter(f...).Sort(s...)
	if len(l) == 0 {
		PrintWarning(Tr(errNoPackage))
		return
	}
	if onlyName {
		names := l.Names()
		fmt.Println(strings.Join(names, "\n"))
		return
	}
	fmt.Println(l)
}

func getFilters(onlyStarred, onlyInstalled, onlyOutDated bool) []database.FilterFunc {
	var filters []database.FilterFunc
	if onlyStarred {
		filters = append(filters, database.FilterStarred)
	}
	if onlyOutDated {
		filters = append(filters, database.FilterOutdated)
	} else if onlyInstalled {
		filters = append(filters, database.FilterInstalled)
	}
	return filters
}

func getFiltersSearch(search string) []database.FilterFunc {
	var filters []database.FilterFunc
	if len(search) > 0 {
		search = strings.ToLower(search)
		filters = append(filters, func(p *database.Package) bool {
			name := strings.ToLower(p.Name)
			if strings.Contains(name, search) {
				return true
			}
			description := strings.ToLower(p.Description)
			return strings.Contains(description, search)
		})
	}
	return filters
}

func getSorters(sortByStar bool) []database.SorterFunc {
	var sorters []database.SorterFunc
	if sortByStar {
		sorters = append(sorters, database.SortByStar)
	}
	sorters = append(sorters, database.SortByName)
	return sorters
}

func update(debug bool) {
	if debug {
		fmt.Fprintln(os.Stderr, "Open db")
	}
	db, _ := getDb()
	if debug {
		fmt.Fprintln(os.Stderr, "Trying to update db")
	}
	counters, err := updateDb(db, debug)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	if err = saveDb(db); err != nil {
		PrintError(err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(color.Yellow.Colorize(counters))
}

func list(
	debug,
	forceUpdate,
	onlyName,
	onlyStarred,
	onlyInstalled,
	onlyOutDated,
	sortByStar bool,
) {
	filters := getFilters(onlyStarred, onlyInstalled, onlyOutDated)
	sorters := getSorters(sortByStar)
	filter(debug, forceUpdate, onlyName, filters, sorters)
}

func search(
	debug,
	forceUpdate,
	onlyName,
	onlyStarred,
	onlyInstalled,
	onlyOutDated,
	sortByStar bool,
	substr string,
) {
	filters := getFilters(onlyStarred, onlyInstalled, onlyOutDated)
	filters = append(filters, getFiltersSearch(substr)...)
	sorters := getSorters(sortByStar)
	filter(debug, forceUpdate, onlyName, filters, sorters)
}

func info(debug bool, app string) {
	db := loadDb(debug, false)
	p, ok := db.Get(app)
	if !ok {
		PrintWarning(Tr(errNoPackage))
		os.Exit(1)
	}
	fmt.Println(p.Detail())
}

func get(debug bool, app string) {
	db := loadDb(debug, false)
	p, ok := db.Get(app)
	if !ok {
		PrintWarning(Tr(errNoPackageOrNeedUpdate))
		os.Exit(1)
	}
	wd, _ := os.Getwd()
	fullDir, err := p.Clone(wd, useSsh())
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	fmt.Println(Tr(msgCloned, app, fullDir))
}

func install(debug bool, app string, asdep bool) {
	db := loadDb(debug, false)
	p, ok := db.Get(app)
	if !ok {
		PrintWarning(Tr(errNoPackageOrNeedUpdate))
		os.Exit(1)
	}
	wd := Config.Get("kcp.tmpDir")
	if err := os.MkdirAll(wd, 0755); err != nil {
		PrintError(err)
		os.Exit(1)
	}
	locker := JoinIfRelative(wd, Config.Get("kcp.lockerFile"))
	if _, err := os.Open(locker); err == nil {
		PrintError(Tr(errOnlyOneInstance))
		os.Exit(1)
	}
	_, err := os.Create(locker)
	if err != nil {
		PrintError(Tr(errFailedCreateLocker))
		os.Exit(1)
	}
	installDir, err := p.Clone(wd, useSsh())
	remove := func() {
		os.Remove(locker)
		os.RemoveAll(installDir)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGHUP)
	go func() {
		<-c
		remove()
		PrintError(Tr(errInterrupt))
		os.Exit(1)
	}()
	defer remove()
	if err != nil {
		remove()
		PrintError(err)
		os.Exit(1)
	}
	os.Chdir(installDir)
	if QuestionYN(Tr(msgEdit), true) {
		if err = EditFile("PKGBUILD"); err != nil {
			remove()
			PrintError(err)
			os.Exit(1)
		}
	}
	m, _ := filepath.Glob("*.install")
	for _, i := range m {
		if QuestionYN(Tr(msgEditInstall, i), false) {
			if err := EditFile(i); err != nil {
				remove()
				PrintError(err)
				os.Exit(1)
			}
		}
	}
	args := []string{"-si"}
	if asdep {
		args = append(args, "--asdeps")
	}
	if err := LaunchCommand("makepkg", args...); err != nil {
		remove()
		PrintError(err)
		os.Exit(1)
	}
	p.LocalVersion = p.GetLocaleVersion()
	saveDb(db)
}

func debugLocales() {
	b, d, l := gotext.GetLibrary(), gotext.GetDomain(), gotext.GetLanguage()
	f := filepath.Join(b, l, "LC_MESSAGES", d+".mo")
	fmt.Println("Debug locale configuration:")
	fmt.Println("- Base path:", b)
	fmt.Println("- Domain:", d)
	fmt.Println("- Language used:", l)
	fmt.Println("- Mo file:", f)
	fmt.Println("- Mo file exists:", FileExists(f))
}
