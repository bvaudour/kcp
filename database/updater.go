package database

import (
	"strings"
	"sync"
	"time"

	"github.com/bvaudour/kcp/common"
	"github.com/google/go-github/v32/github"
)

//Updater provides utilities
//to update a local database
//using a remote connector to an organization.
type Updater struct {
	db   *Database
	repo *Repository
}

//Counter is a counter of updated packages.
type Counter struct {
	Updated int
	Deleted int
	Added   int
}

//String returns the string representation of the counter
func (c Counter) String() string {
	var out []string
	if c.Added > 0 {
		out = append(out, common.Tr(msgAdded, c.Added))
	}
	if c.Deleted > 0 {
		out = append(out, common.Tr(msgDeleted, c.Deleted))
	}
	if c.Updated > 0 {
		out = append(out, common.Tr(msgUpdated, c.Updated))
	}
	return strings.Join(out, "\n")
}

//Diff refresh the counter comparing 2 package sets.
func (c *Counter) Diff(ps1, ps2 *PackageSet) {
	ps1.Iterate(func(p *Package) {
		if !ps2.Contains(p.Name) {
			c.Deleted++
		}
	})
	ps2.Iterate(func(p2 *Package) {
		p1, exists := ps1.Get(p2.Name)
		if !exists {
			c.Added++
		} else if !p1.Eq(p2) {
			c.Updated++
		}
	})
}

//NewUpdater return a new updater.
func NewUpdater(db *Database, repo *Repository) *Updater {
	return &Updater{
		db:   db,
		repo: repo,
	}
}

func (u *Updater) updatePackage(p *Package, ps *PackageSet) {
	p.LocalVersion = p.GetLocaleVersion()
	p1, ok := ps.Get(p.Name)
	if !ok || u.db.LastUpdate.IsZero() || u.db.LastUpdate.Before(p.PushedAt) {
		if pkg, err := p.GetPKGBUID(); err == nil {
			p.udpateFromPKGBUILD(pkg)
			return
		}
	} else {
		p.updateFromPackage(p1)
	}
}

//Update updates the local database
//and returns the counter of modifications.
func (u *Updater) Update() (c Counter, err error) {
	ps1 := u.db.ToSet()
	ignore := s2m(u.db.IgnoreRepos)
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	ps2 := NewPackageSet()
	var pl Packages
	var nextPage int
	var wg sync.WaitGroup
	for {
		if pl, nextPage, err = u.repo.GetPage(opt); err != nil {
			break
		}
		for p := range pl.Iterator() {
			go (func(p *Package) {
				wg.Add(1)
				defer wg.Done()
				if ignore[p.Name] {
					return
				}
				ps2.Add(p)
				u.updatePackage(p, ps1)
			})(p)
		}
		if nextPage == 0 {
			break
		}
		opt.Page = nextPage
	}
	wg.Wait()
	if err == nil {
		c.Diff(ps1, ps2)
		u.db.LastUpdate = time.Now()
		u.db.Packages = ps2.ToList()
	}
	return
}
