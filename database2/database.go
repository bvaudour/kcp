package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bvaudour/kcp/color"
)

// Database is the decoded structure
// of a json database of packages.
type Database struct {
	LastUpdate    time.Time `json:"last_update"`
	IgnoreRepos   []string  `json:"ignore_repos"`
	BrokenDepends []string  `json:"broken_depends"`
	Packages      `json:"packages"`
}

// New returns a new empty database initialized
// by repositories of the organzation to ignore.
func New(ignored ...string) Database {
	return Database{
		IgnoreRepos: ignored,
	}
}

// Decode decodes the given file to the database
func (db *Database) Decode(r io.Reader) error {
	dec := json.NewDecoder(r)

	return dec.Decode(db)
}

// Encode encodes the database to json
// and write it to the given file.
func (db Database) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)

	return enc.Encode(db)
}

// Load decodes the file in the given path and
// returns the decoded database.
func Load(fpath string, ignored ...string) (db Database, err error) {
	var f *os.File
	db = New(ignored...)

	if f, err = os.Open(fpath); err != nil {
		return
	}
	defer f.Close()

	err = db.Decode(f)
	db.IgnoreRepos = ignored

	return
}

// Save writes the database into the file on the given path.
func Save(fpath string, db Database) (err error) {
	var f *os.File

	if f, err = os.Create(fpath); err != nil {
		return
	}
	defer f.Close()

	return db.Encode(f)
}

// UpdateBroken updates the broken depends.
func (db *Database) UpdateBroken() {
	db.BrokenDepends = db.Packages.SearchBroken()
}

// UpdateRemote updates the database from a github organization.
// If optional user and password are given, requests are done
// with authentification in order to have a better rate limit.
func (db *Database) UpdateRemote(organization string, debug bool, opt ...string) (counter Counter, err error) {
	limit, routines := defaultLimit, defaultRoutines
	repo := NewRepository(organization, opt...)

	var nbPages, nbRepos int
	if nbPages, nbRepos, err = repo.CountPages(limit); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"%s %s\n",
			color.Red.Format("[Error: %s]", err),
			"Failed to count the pages of the repository list",
		)
		return
	}

	var newPackages Packages
	ignored := sliceToSet(db.IgnoreRepos)
	lastUpdate, newUpdate := db.LastUpdate, time.Now().Add(time.Hour*-24)

	packages := make(chan Package, (nbPages-1)*limit+1)
	buffer := make(chan Package, nbRepos)

	go (func() {
		for {
			p, ok := <-buffer
			if !ok {
				return
			}
			newPackages.Push(p)
		}
	})()

	for i := 0; i < routines; i++ {
		go (func() {
			for {
				p, ok := <-packages
				if !ok {
					return
				}
				if ignored[p.Name] {
					continue
				}

				p.PkgbuildUrl = fmt.Sprintf(baseRawURL, repo.organization, p.Name, p.Branch)
				p.LocalVersion = p.GetLocaleVersion()
				if p.noChange = p.UpdatedAt.Before(lastUpdate); !p.noChange {
					if file, e := p.GetPKGBUID(debug); e == nil {
						p.updateFromPKGBUILD(file)
					}
				}
				buffer <- p
			}
		})()
	}

	for i := 1; i <= nbPages; i++ {
		go (func(i int) {
			page, e := repo.GetPage(i, limit, debug)
			if e != nil {
				err = e
				return
			}
			for _, p := range page {
				packages <- p
			}
		})(i)
	}

	close(packages)
	close(buffer)

	if err != nil {
		return
	}

	psOld := db.Packages.ToSet()

	for i, p := range newPackages {
		p0, ok := psOld[p.Name]
		if !ok {
			counter.Added++
		} else if !p.noChange {
			counter.Updated++
		} else {
			newPackages[i].updateFromPackage(p0)
		}
	}

	psNew := newPackages.ToSet()

	for n := range psOld {
		if _, ok := psNew[n]; !ok {
			counter.Deleted++
		}
	}

	db.Packages, db.LastUpdate = newPackages, newUpdate

	return
}

// Update checks if updates are available in the database.
func (db *Database) Update(organization string, debug bool, opt ...string) (counter Counter, err error) {
	if counter, err = db.UpdateRemote(organization, debug, opt...); err == nil {
		db.UpdateBroken()
	}

	return
}
