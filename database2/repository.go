package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bvaudour/kcp/color"
)

// Repository is a connector to access to the repos infos
// of a github organization.
type Repository struct {
	organization string
	ctx
}

// NewRepository creates a connector to an organization.
// If optional user and password are given, requests are done
// with authentification in order to have a better rate limit.
func NewRepository(organization string, opt ...string) Repository {
	var user, password string
	if len(opt) >= 2 {
		user, password = opt[0], opt[1]
	}
	return Repository{
		organization: organization,
		ctx: ctx{
			username: user,
			password: password,
			accept:   acceptHeader,
		},
	}
}

// CountPublicRepos returns the number of repositories
// owned by the organization.
func (r Repository) CountPublicRepos() (nb int, err error) {
	var buf io.Reader

	if buf, err = execRequest(baseOrganizationURL, r.ctx, r.organization); err == nil {
		var result struct {
			PublicRepos int `json:"public_repos"`
		}
		dec := json.NewDecoder(buf)
		if err = dec.Decode(&result); err == nil {
			nb = result.PublicRepos
		}
	}

	return
}

// CountPages gives the number of pages of repos and of repos in the organization
func (r Repository) CountPages(limit int) (pages, repos int, err error) {
	if limit == 0 {
		panic("Limit should be > 0")
	}

	if repos, err = r.CountPublicRepos(); err == nil && repos > 0 {
		pages = (repos-1)/limit + 1
	}

	return
}

// GetPage returns the remote packages’ infos on
// the repositories list page of the organization.
func (r *Repository) GetPage(page, limit int, debug ...bool) (packages Packages, err error) {
	var buf io.Reader

	buf, err = execRequest(baseReposURL, r.ctx, r.organization, page, limit)
	if err == nil {
		dec := json.NewDecoder(buf)
		err = dec.Decode(&packages)
	}

	if len(debug) > 0 && debug[0] {
		var t string
		if err == nil {
			t = color.Green.Colorize("[Success]")
		} else {
			t = color.Red.Format("[Error: %s]", err)
		}
		fmt.Fprintf(
			os.Stderr,
			"%s "+baseReposURL+"\n",
			t,
			r.organization,
			page,
			limit,
		)
	}

	return
}
