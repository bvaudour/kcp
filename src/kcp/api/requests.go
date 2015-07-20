package api

import (
	"errors"
	"fmt"
	"gettext"
	"os"
	"os/signal"
)

// Translation methods
func Translate(arg string) string                        { return gettext.Gettext(arg) }
func Translatef(form string, args ...interface{}) string { return fmt.Sprintf(Translate(form), args...) }

// Local requests
func LocalListAll() PList {
	return localAll(true).List()
}

func LocalMapAll() PMap {
	return localAll(false).Map()
}

func LocalListSearch(app string) PList {
	return localSearch(app, true).List()
}

func LocalMapSearch(app string) PMap {
	return localSearch(app, false).Map()
}

func LocalVersion(app string) string {
	v, _ := localVersion(app)
	return v
}

// Remote requests
func KcpListAll(debug bool) (PList, error) {
	c, e := remoteAll(SEARCH_ALL, debug, true, false)
	return c.List(), e
}

func KcpMapAll(debug bool) (PMap, error) {
	c, e := remoteAll(SEARCH_ALL, debug, false, false)
	return c.Map(), e
}

func KcpListAllWithVersions(debug bool) (PList, error) {
	c, e := remoteAll(SEARCH_ALL, debug, true, true)
	return c.List(), e
}

func KcpMapAllWithVersions(debug bool) (PMap, error) {
	c, e := remoteAll(SEARCH_ALL, debug, false, true)
	return c.Map(), e
}

func KcpListStarred(debug bool) (PList, error) {
	c, e := remoteAll(SEARCH_ALLST, debug, true, true)
	return c.List(), e
}

func KcpMapStarred(debug bool) (PMap, error) {
	c, e := remoteAll(SEARCH_ALLST, debug, false, true)
	return c.Map(), e
}

func KcpListSearch(app string, debug bool) (PList, error) {
	c, e := remoteSearch(app, debug, true)
	return c.List(), e
}

func KcpMapSearch(app string, debug bool) (PMap, error) {
	c, e := remoteSearch(app, debug, false)
	return c.Map(), e
}

func KcpVersion(app string) string {
	v, _ := remoteVersion(app)
	return v
}

func KcpVersion2(app string) string {
	v, _ := remoteVersion2(app)
	return v
}

// Git clone
func Get(app string, stderr bool) error {
	if !repoExists(app) {
		return errors.New(Translate(MSG_NOT_FOUND))
	}
	path := pathOf(app)
	if pathExists(path) {
		return errors.New(Translatef(MSG_DIREXISTS, path))
	}
	return cloneRepo(app, stderr)
}

// Install app
func Install(app string, asdeps bool) error {
	tmpdir := os.TempDir()
	cd(tmpdir)

	lck, wdir := pathOf(KCP_LOCK), pathOf(app)
	if _, e := os.Open(lck); e == nil {
		return errors.New(Translate(MSG_ONLYONEINSTANCE))
	}
	os.Create(lck)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		endInstall(wdir, lck)
		fmt.Println(Translate(MSG_INTERRUPT))
	}()
	defer endInstall(wdir, lck)

	if e := Get(app, true); e != nil {
		endInstall(wdir, lck)
		return e
	}
	cd(wdir)
	if question(Translate(MSG_EDIT), true) {
		if e := editPkgbuild(); e != nil {
			endInstall(wdir, lck)
			return e
		}
	}
	for _, inst := range searchInstallFiles() {
		if question(Translatef(MSG_EDIT_INSTALL, inst), false) {
			if e := editFile(inst); e != nil {
				endInstall(wdir, lck)
				return e
			}
		}
	}

	var e error
	if asdeps {
		e = launchCommand("makepkg", true, "-si", "--asdeps")
	} else {
		e = launchCommand("makepkg", true, "-si")
	}
	endInstall(wdir, lck)
	return e
}

// Database Management
func LoadListDB() (PList, error) {
	c, e := loadDB(pathJoin(userDir(), KCP_DB), true)
	return c.List(), e
}

func LoadMapDB() (PMap, error) {
	c, e := loadDB(pathJoin(userDir(), KCP_DB), false)
	return c.Map(), e
}

func SaveDB(c PCollection) error { return saveDB(pathJoin(userDir(), KCP_DB), c) }
