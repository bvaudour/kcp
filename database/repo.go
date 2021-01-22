package database

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bvaudour/kcp/color"
	"github.com/google/go-github/v33/github"
)

const (
	rawURL = "https://raw.githubusercontent.com/KaOS-Community-Packages/%s/master/PKGBUILD"
)

//Repository is a connector to access to the repos infos
//of a github organization.
type Repository struct {
	organization string
	client       *github.Client
	ctx          context.Context
}

//NewRepository creates a connector to an organization.
//If optional user and password are given, requests are done
//with authentification in order to have a better rate limit.
func NewRepository(organization string, opt ...string) *Repository {
	var user, password string
	if len(opt) >= 2 {
		user, password = opt[0], opt[1]
	}
	var client *http.Client
	if user != "" && password != "" {
		auth := github.BasicAuthTransport{
			Username: user,
			Password: password,
		}
		client = auth.Client()
	}
	return &Repository{
		organization: organization,
		client:       github.NewClient(client),
		ctx:          context.Background(),
	}
}

func (r *Repository) getRepos(opt *github.RepositoryListByOrgOptions) (repos []*github.Repository, resp *github.Response, err error) {
	return r.client.Repositories.ListByOrg(r.ctx, r.organization, opt)
}

//GetPage the returns the remote packages’ infos on
//the repositories list page of the organization.
func (r *Repository) GetPage(opt *github.RepositoryListByOrgOptions, debug ...bool) (packages Packages, nextPage int, err error) {
	var repos []*github.Repository
	var resp *github.Response
	if repos, resp, err = r.getRepos(opt); err == nil {
		nextPage = resp.NextPage
		packages = make(Packages, len(repos))
		for i, repo := range repos {
			var description string
			if repo.Description != nil {
				description = *repo.Description
			}
			packages[i] = &Package{
				Name:        *repo.Name,
				Description: description,
				CreatedAt:   repo.CreatedAt.Time,
				UpdatedAt:   repo.UpdatedAt.Time,
				PushedAt:    repo.PushedAt.Time,
				RepoUrl:     *repo.HTMLURL,
				CloneUrl:    *repo.CloneURL,
				SshUrl:      *repo.SSHURL,
				PkgbuildUrl: fmt.Sprintf(rawURL, *repo.Name),
				Stars:       *repo.StargazersCount,
			}
		}
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
			"%s https://api.github.com/%s/repos?page=%d&per_page=%d\n",
			t,
			r.organization,
			opt.Page,
			opt.PerPage,
		)
	}
	return
}
