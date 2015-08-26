//Package kcpdb provides utilities to manage the database of KCP repositories.
package kcpdb

import (
	"fmt"
	"gettext"
	"io/ioutil"
	"os"
	"parser/json"
	"path/filepath"
	"sort"
	"strings"
)

var tr = gettext.Gettext

//Needed keys for database
const (
	DB_NAME         = "name"
	DB_DESCRIPTION  = "description"
	DB_STARS        = "stars"
	DB_LOCALVERSION = "localversion"
	DB_KCPVERSION   = "kcpversion"
)

//Output elements
const (
	INSTALLED_VERSION  = "[installed]"
	INSTALLED_VERSIONF = "[installed: %s]"
)

//Package groups the useful informations about a package.
type Package struct {
	Name         string
	Description  string
	LocalVersion string
	KcpVersion   string
	Stars        int64
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

//Object converts the package to a json object
func (p *Package) Object() json.Object {
	o := make(json.Object)
	o[DB_NAME] = p.Name
	o[DB_DESCRIPTION] = p.Description
	o[DB_LOCALVERSION] = p.LocalVersion
	o[DB_KCPVERSION] = p.KcpVersion
	o[DB_STARS] = p.Stars
	return o
}

//LoadPkg converts an object to a package type.
func LoadPkg(o json.Object) (p *Package) {
	if s, e := o.GetString(DB_NAME); e == nil {
		p = new(Package)
		p.Name = s
		p.Description, _ = o.GetString(DB_DESCRIPTION)
		p.LocalVersion, _ = o.GetString(DB_LOCALVERSION)
		p.KcpVersion, _ = o.GetString(DB_KCPVERSION)
		p.Stars, _ = o.GetInt64(DB_STARS)
	}
	return
}

//Packages' list sorting tool.
type plSorter struct {
	l []*Package
	f func(*Package, *Package) bool
}

func (s *plSorter) Len() int           { return len(s.l) }
func (s *plSorter) Less(i, j int) bool { return s.f(s.l[i], s.l[j]) }
func (s *plSorter) Swap(i, j int)      { s.l[i], s.l[j] = s.l[j], s.l[i] }

func sortList(l []*Package, f func(*Package, *Package) bool) {
	s := &plSorter{l, f}
	sort.Sort(s)
}

//Database is the database of the KCP packages.
type Database map[string]*Package

//Add appends the given package to the database.
func (db Database) Add(p *Package) { db[p.Name] = p }

//Names returns the packages' names of the database.
func (db Database) Names() []string {
	names := make([]string, 0, len(db))
	for n := range db {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

//Packages returns the packages contained in the database.
func (db Database) Packages() []*Package {
	packages := make([]*Package, 0, len(db))
	for _, p := range db {
		packages = append(packages, p)
	}
	return packages
}

//Filter returns a new database matching all filters.
func (db Database) Filter(filters ...func(*Package) bool) Database {
	dbf := New()
loop:
	for n, p := range db {
		for _, f := range filters {
			if !f(p) {
				continue loop
			}
		}
		dbf[n] = p
	}
	return dbf
}

//Sorted returns a sorted list of package according to the given sort criteria.
func (db Database) Sorted(f func(*Package, *Package) bool) []*Package {
	packages := db.Packages()
	sortList(packages, f)
	return packages
}

//Merge fusions the actual database with the new.
func (db Database) Merge(dbn Database) (updated int, added int, deleted int) {
	for n, p := range dbn {
		p_db, ok := db[n]
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
		if !ok {
			updated++
		}
	}
	for n := range db {
		if _, ok := dbn[n]; !ok {
			delete(db, n)
			deleted++
		}
	}
	return
}

//Save saves the database into a local file.
func (db Database) SaveBD(file string) error {
	o := make([]json.Object, 0)
	for _, p := range db {
		o = append(o, p.Object())
	}
	b, e := json.Marshal(o)
	if e == nil {
		if e = os.MkdirAll(filepath.Dir(file), 0755); e == nil {
			e = ioutil.WriteFile(file, b, 0644)
		}
	}
	return e
}

//New returns an empty database.
func New() Database { return make(Database) }

//LoadDB loads a database from a local file.
func LoadBD(file string) (db Database, e error) {
	db = New()
	var f *os.File
	if f, e = os.Open(file); e == nil {
		var o []json.Object
		if o, e = json.ArrayObjectReader(f); e == nil {
			for _, v := range o {
				if p := LoadPkg(v); p != nil {
					db.Add(p)
				}
			}
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
