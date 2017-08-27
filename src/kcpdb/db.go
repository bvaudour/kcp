//Package kcpdb provides utilities to manage the database of KCP repositories.
package kcpdb

import (
	"encoding/json"
	"fmt"
	"gettext"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sysutil"
)

var tr = gettext.Gettext

//Needed keys for database
const (
	DB_NAME         = "name"
	DB_DESCRIPTION  = "description"
	DB_STARS        = "stars"
	DB_LOCALVERSION = "localversion"
	DB_KCPVERSION   = "kcpversion"
	DB_PUSHED_AT    = "pushed_at"
)

//Output elements
const (
	INSTALLED_VERSION  = "[installed]"
	INSTALLED_VERSIONF = "[installed: %s]"
)

//Package groups the useful informations about a package.
type Package struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	LocalVersion string `json:"localversion"`
	KcpVersion   string `json:"kcpversion"`
	Stars        int64  `json:"stars"`
	PushedAt     string `json:"pushed_at"`
}

//String returns the string representation of a package.
func (p *Package) String() string {
	name := fmt.Sprintf("\033[1m%s\033[m", p.Name)
	kcpversion := fmt.Sprintf("\033[1;32m%s\033[m", p.KcpVersion)
	localversion := ""
	if p.LocalVersion != "" {
		if p.LocalVersion == p.KcpVersion {
			localversion = tr(INSTALLED_VERSION)
		} else {
			localversion = fmt.Sprintf(tr(INSTALLED_VERSIONF), p.LocalVersion)
		}
		localversion = fmt.Sprintf("\033[1;36m%s\033[m", localversion)
	}
	stars := fmt.Sprintf("\033[1;34m(%d)\033[m", p.Stars)
	description := ""
	if p.Description != "" {
		description = fmt.Sprintf("\n\t%s", p.Description)
	}
	return fmt.Sprintf("\033[1;35mkcp/\033[m%s %s %s %s %s", name, kcpversion, localversion, stars, description)
}

//Packages' list sorting tool.
func sortList(l []*Package, f func(*Package, *Package) bool) {
	sort.Slice(l, func(i, j int) bool { return f(l[i], l[j]) })
}

//Database is the database of the KCP packages.
type Database struct {
	LastSync string              `json:"lastsync"`
	List     map[string]*Package `json:"packages"`
}

//Add appends the given package to the database.
func (db *Database) Add(p *Package) { db.List[p.Name] = p }

//Names returns the packages' names of the database.
func (db Database) Names() []string {
	names := make([]string, 0, len(db.List))
	for n := range db.List {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

//Packages returns the packages contained in the database.
func (db Database) Packages() []*Package {
	packages := make([]*Package, 0, len(db.List))
	for _, p := range db.List {
		packages = append(packages, p)
	}
	return packages
}

//LastSynchronization returns the timestamp of the last sync date
func (db *Database) LastSynchronization() int64 { return sysutil.StrToTimestamp(db.LastSync) }

//Filter returns a new database matching all filters.
func (db *Database) Filter(filters ...func(*Package) bool) *Database {
	dbf := New()
	dbf.LastSync = db.LastSync
loop:
	for _, p := range db.List {
		for _, f := range filters {
			if !f(p) {
				continue loop
			}
		}
		dbf.Add(p)
	}
	return dbf
}

//Sorted returns a sorted list of package according to the given sort criteria.
func (db *Database) Sorted(f func(*Package, *Package) bool) []*Package {
	packages := db.Packages()
	sortList(packages, f)
	return packages
}

//Merge fusions the actual database with the new.
func (db *Database) Merge(dbn *Database) (updated int, added int, deleted int) {
	for n, p := range dbn.List {
		p_db, ok := db.List[n]
		if !ok {
			db.Add(p)
			added++
			continue
		}
		if p.Description != "" && p.Description != p_db.Description {
			ok = false
			p_db.Description = p.Description
		}
		if p.Stars != p_db.Stars {
			ok = false
			p_db.Stars = p.Stars
		}
		if p.LocalVersion != p_db.LocalVersion {
			ok = false
			p_db.LocalVersion = p.LocalVersion
		}
		if p.KcpVersion != "" && p.KcpVersion != p_db.KcpVersion {
			ok = false
			p_db.KcpVersion = p.KcpVersion
		}
		if p.PushedAt != p_db.PushedAt {
			ok = false
			p_db.PushedAt = p.PushedAt
		}
		if !ok {
			updated++
		}
	}
	for n := range db.List {
		if _, ok := dbn.List[n]; !ok {
			delete(db.List, n)
			deleted++
		}
	}
	return
}

//Save saves the database into a local file.
func (db *Database) SaveBD(file string) error {
	db.LastSync = sysutil.TimestampToString(sysutil.Now())
	b, e := json.Marshal(db)
	if e == nil {
		if e = os.MkdirAll(filepath.Dir(file), 0755); e == nil {
			e = ioutil.WriteFile(file, b, 0644)
		}
	}
	return e
}

//New returns an empty database.
func New() *Database {
	return &Database{
		List: make(map[string]*Package),
	}
}

//LoadDB loads a database from a local file.
func LoadBD(file string) (db *Database, e error) {
	db = New()
	var f *os.File
	if f, e = os.Open(file); e == nil {
		dc := json.NewDecoder(f)
		if e = dc.Decode(db); e == io.EOF {
			e = nil
		}
	}
	return
}

//Useful filters and sort functions
var (
	SortByName = func(p1, p2 *Package) bool { return p1.Name < p2.Name }
	SortByStar = func(p1, p2 *Package) bool { return p1.Stars > p2.Stars || (p1.Stars == p2.Stars && SortByName(p1, p2)) }

	FilterInstalled = func(p *Package) bool { return p.LocalVersion != "" }
	FilterOutdated  = func(p *Package) bool { return FilterInstalled(p) && p.LocalVersion != p.KcpVersion }
	FilterStar      = func(p *Package) bool { return p.Stars > 0 }

	FilterName = func(name string) func(*Package) bool {
		return func(p *Package) bool { return strings.Contains(p.Name, name) }
	}
	FilterDescription = func(name string) func(*Package) bool {
		return func(p *Package) bool { return strings.Contains(p.Description, name) }
	}
	FilterNameOrDescription = func(name string) func(*Package) bool {
		return func(p *Package) bool { return FilterName(name)(p) || FilterDescription(name)(p) }
	}
)
