package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/bvaudour/kcp/color"
)

const (
	baseUrl             = "https://api.github.com/orgs"
	baseRawURL          = "https://raw.githubusercontent.com/%s/%s/%s/PKGBUILD"
	baseOrganizationURL = baseUrl + "/%s"
	baseReposURL        = baseOrganizationURL + "/repos?page=%d&per_page=%d"
	acceptHeader        = "application/vnd.github.v3+json"
	defaultLimit        = 100
)

type ctx struct {
	username string
	password string
	accept   string
}

func execRequest(url string, opt ctx, args ...interface{}) (io.Reader, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf(url, args...), nil)
	if err != nil {
		return nil, err
	}
	if opt.username != "" && opt.password != "" {
		request.SetBasicAuth(opt.username, opt.password)
	}
	if opt.accept != "" {
		request.Header.Set("Accept", opt.accept)
	}
	response, err := new(http.Client).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err == nil {
		return bytes.NewBuffer(b), nil
	}
	return bytes.NewBuffer([]byte{}), nil
}

//Repository is a connector to access to the repos infos
//of a github organization.
type Repository struct {
	organization string
	ctx
}

//NewRepository creates a connector to an organization.
//If optional user and password are given, requests are done
//with authentification in order to have a better rate limit.
func NewRepository(organization string, opt ...string) *Repository {
	var user, password string
	if len(opt) >= 2 {
		user, password = opt[0], opt[1]
	}
	return &Repository{
		organization: organization,
		ctx: ctx{
			username: user,
			password: password,
			accept:   acceptHeader,
		},
	}
}

func (r *Repository) countPublicRepos() (nb int, err error) {
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

func (r *Repository) countPages(limit int) (pages int, err error) {
	if limit == 0 {
		panic("Limit should be > 0")
	}
	var nbRepos int
	if nbRepos, err = r.countPublicRepos(); err == nil && nbRepos > 0 {
		pages = (nbRepos-1)/limit + 1
	}
	return
}

//GetPage returns the remote packages’ infos on
//the repositories list page of the organization.
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

//GetPublicRepos returns all remote packages’ infos on
//the repositories list of the organization.
func (r *Repository) GetPublicRepos(debug ...bool) (packages Packages, err error) {
	limit := defaultLimit
	var nbPages int
	if nbPages, err = r.countPages(limit); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"%s %s\n",
			color.Red.Format("[Error: %s]", err),
			"Failed to count the pages of the repository list",
		)
		return
	}
	var wg sync.WaitGroup
	wg.Add(nbPages)
	var mtx sync.Mutex
	for page := 1; page <= nbPages; page++ {
		go (func(page int) {
			defer wg.Done()
			pl, err2 := r.GetPage(page, limit, debug...)
			if err2 != nil {
				err = err2
				return
			}
			mtx.Lock()
			defer mtx.Unlock()
			packages = append(packages, pl...)
		})(page)
	}
	wg.Wait()
	return
}
