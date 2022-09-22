package database

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bvaudour/kcp/common"
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

func (u *Updater) updatePackage(p *Package, ps *PackageSet, debug ...bool) {
	p.LocalVersion = p.GetLocaleVersion()
	p1, ok := ps.Get(p.Name)
	if !ok || u.db.LastUpdate.IsZero() || u.db.LastUpdate.Before(p.PushedAt) {
		if pkg, err := p.GetPKGBUID(debug...); err == nil {
			p.udpateFromPKGBUILD(pkg)
			return
		}
	} else {
		p.updateFromPackage(p1)
	}
}

//Update updates the local database
//and returns the counter of modifications.
func (u *Updater) Update(debug ...bool) (c Counter, err error) {
	var pl Packages
	if pl, err = u.repo.GetPublicRepos(debug...); err != nil {
		return
	}
	ps1 := u.db.ToSet()
	ps2 := NewPackageSet()
	ignore := s2m(u.db.IgnoreRepos)
	var wg sync.WaitGroup
	wg.Add(len(pl))
	limitRoutines := make(chan struct{}, 20)
	for _, p := range pl {
		go (func(p *Package) {
			limitRoutines <- struct{}{}
			defer (func() { <-limitRoutines })()
			defer wg.Done()
			if ignore[p.Name] {
				return
			}
			p.PkgbuildUrl = fmt.Sprintf(baseRawURL, u.repo.organization, p.Name, p.Branch)
			ps2.Add(p)
			u.updatePackage(p, ps1, debug...)
		})(p)
	}
	wg.Wait()
	close(limitRoutines)
	c.Diff(ps1, ps2)
	u.db.BrokenDepends = ps2.searchBroken()
	u.db.LastUpdate = time.Now()
	u.db.Packages = ps2.ToList()
	return
}
