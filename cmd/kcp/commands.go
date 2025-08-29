package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"codeberg.org/bvaudour/kcp/common"
	"codeberg.org/bvaudour/kcp/database"
	"git.kaosx.ovh/benjamin/format"
	"github.com/leonelquinteros/gotext"
)

func getIgnore() []string {
	return strings.Fields(common.Config.Get("kcp.ignore"))
}

func getDbPath() string {
	return common.JoinIfRelative(common.UserBaseDir, common.Config.Get("kcp.dbFile"))
}

func useSsh() bool {
	return common.Config.Get("kcp.cloneMethod") == "ssh"
}

func getDb() (database.Database, error) {
	fpath, ignore := getDbPath(), getIgnore()
	return database.Load(fpath, ignore...)
}

func updateDb(db *database.Database, debug bool) (database.Counter, error) {
	return db.Update(database.NewConnector(), debug)
}

func loadDb(debug, forceUpdate bool) database.Database {
	db, err := getDb()
	if debug {
		fmt.Fprintln(os.Stderr, "Trying to open db")
	}
	if err != nil || forceUpdate {
		if debug {
			fmt.Fprintln(os.Stderr, "Trying to update db")
		}
		updateDb(&db, debug)
	}
	return db
}

func saveDb(db database.Database) error {
	return database.Save(getDbPath(), db)
}

func filter(debug, forceUpdate, onlyName bool, f []database.FilterFunc, s []database.SorterFunc) {
	db := loadDb(debug, forceUpdate)
	saveDb(db)
	l := db.Filter(f...).Sort(s...)
	if len(l) == 0 {
		common.PrintWarning(common.Tr(errNoPackage))
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
		filters = append(filters, func(p database.Package) bool {
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
	db, err := getDb()
	if err != nil {
		db = database.New(getIgnore()...)
	}
	if debug {
		fmt.Fprintln(os.Stderr, "Trying to update db")
	}
	counters, err := updateDb(&db, debug)
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	if err = saveDb(db); err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	fmt.Println()
	format.FormatOf("yellow").Println(counters)
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
		common.PrintWarning(common.Tr(errNoPackage))
		os.Exit(1)
	}
	fmt.Println(p.Detail())
}

func get(debug bool, app string) {
	db := loadDb(debug, false)
	p, ok := db.Get(app)
	if !ok {
		common.PrintWarning(common.Tr(errNoPackageOrNeedUpdate))
		os.Exit(1)
	}
	wd, _ := os.Getwd()
	fullDir, err := p.Clone(wd, useSsh())
	if err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	fmt.Println(common.Tr(msgCloned, app, fullDir))
}

func install(debug bool, app string, asdep bool) {
	db := loadDb(debug, false)
	p, ok := db.Get(app)
	if !ok {
		common.PrintWarning(common.Tr(errNoPackageOrNeedUpdate))
		os.Exit(1)
	}
	wd := common.Config.Get("kcp.tmpDir")
	if err := os.MkdirAll(wd, 0755); err != nil {
		common.PrintError(err)
		os.Exit(1)
	}
	locker := common.JoinIfRelative(wd, common.Config.Get("kcp.lockerFile"))
	if _, err := os.Open(locker); err == nil {
		common.PrintError(common.Tr(errOnlyOneInstance))
		os.Exit(1)
	}
	_, err := os.Create(locker)
	if err != nil {
		common.PrintError(common.Tr(errFailedCreateLocker))
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
		common.PrintError(common.Tr(errInterrupt))
		os.Exit(1)
	}()
	defer remove()
	if err != nil {
		remove()
		common.PrintError(err)
		os.Exit(1)
	}
	os.Chdir(installDir)
	if common.QuestionYN(common.Tr(msgEdit), true) {
		if err = common.EditFile("PKGBUILD"); err != nil {
			remove()
			common.PrintError(err)
			os.Exit(1)
		}
	}
	m, _ := filepath.Glob("*.install")
	for _, i := range m {
		if common.QuestionYN(common.Tr(msgEditInstall, i), false) {
			if err := common.EditFile(i); err != nil {
				remove()
				common.PrintError(err)
				os.Exit(1)
			}
		}
	}
	args := []string{"-si"}
	if asdep {
		args = append(args, "--asdeps")
	}
	if err := common.LaunchCommand("makepkg", args...); err != nil {
		remove()
		common.PrintError(err)
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
	fmt.Println("- Mo file exists:", common.FileExists(f))
}
