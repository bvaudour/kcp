//Package repo provides useful function to manage requests with the KaOS Community Packages.
package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"gettext"
	"io/ioutil"
	"kcpdb"
	"net/http"
	"os"
	"parser/pkgbuild"
	"sync"
	"sysutil"
)

//Needed URLs for requests
const (
	HEADER        = "application/vnd.github.v3+json"
	SEARCH_ALL    = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100&%s"
	URL_REPO      = "https://github.com/KaOS-Community-Packages/%s.git"
	URL_PKGBUILD  = "https://raw.githubusercontent.com/KaOS-Community-Packages/%s/master/PKGBUILD"
	APP_ID        = "&client_id=11f5f3d9dab26c7fff24"
	SECRET_ID     = "&client_secret=bb456e9fa4e2d0fe2df9e194974c98c2f9133ff5"
	IDENT         = APP_ID + SECRET_ID
	PKGBUILDPROTO = "https://raw.githubusercontent.com/kaos-addict/kaos-helpers/master/PKGBUILD.commented.kaos.proto"
)

//Json keys of github API
const (
	NAME          = "name"
	DESCRIPTION   = "description"
	STARS         = "stargazers_count"
	ITEMS         = "items"
	MESSAGE       = "message"
	DOCUMENTATION = "documentation_url"
	PUSHED_AT     = "pushed_at"
)

//Messages
const (
	MSG_NOT_FOUND   = "Package not found!"
	MSG_UNKNOWN     = "Unknown error!"
	MSG_PATH_EXISTS = "Dir %s already exists!"
	UNKNOWN_VERSION = "<unknown>"
	MSG_SYNC_ERROR  = "Failed on synchronize database. Please, try later."
)

type kcpPackage struct {
	Name         string `json:"name"`
	LocalVersion string `json:"localversion"`
	KcpVersion   string `json:"kcpversion"`
	Description  string `json:"description"`
	Stars        int64  `json:"stargazers_count"`
	PushedAt     string `json:"pushed_at"`
}

//List of ignore repos
var ignoreRepo = map[string]bool{
	"KaOS-Community-Packages.github.io": true,
}

var tr = gettext.Gettext

func o2e(b []byte) error {
	msg := struct {
		Message       string `json:"message"`
		Documentation string `json:"documentation_url"`
	}{}
	e := json.Unmarshal(b, &msg)
	if e != nil {
		return errors.New(tr(MSG_UNKNOWN))
	}
	return fmt.Errorf("%s\n%s\n", msg.Message, msg.Documentation)
}

func launchRequest(debug bool, header string, searchbase string, v ...interface{}) (b []byte, e error) {
	var req *http.Request
	if req, e = http.NewRequest("GET", fmt.Sprintf(searchbase, v...), nil); e != nil {
		return
	}
	if header != "" {
		req.Header.Add("Accept", header)
	}
	var resp *http.Response
	if resp, e = new(http.Client).Do(req); e != nil {
		return
	}
	if debug {
		resp.Write(os.Stdout)
	}
	b, e = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}

func listPkg(search string, debug bool, last_pushed int64) (db *kcpdb.Database, e error) {
	db = kcpdb.New()
	var wg sync.WaitGroup
	var mx = new(sync.RWMutex)
	end := false
	for i := 1; ; i++ {
		if end {
			break
		}
		b, err := launchRequest(debug, HEADER, search, i, IDENT)
		if err != nil {
			end = true
			e = err
			return
		}
		var packages []kcpPackage
		err = json.Unmarshal(b, &packages)
		if err != nil {
			end = true
			e = o2e(b)
			return
		}
		if len(packages) == 0 {
			end = true
			break
		}
		for _, p := range packages {
			go func(p kcpPackage) {
				wg.Add(1)
				defer wg.Done()
				if p.Name != "" && !ignoreRepo[p.Name] {
					pp := &kcpdb.Package{
						Name:         p.Name,
						Description:  p.Description,
						Stars:        p.Stars,
						PushedAt:     p.PushedAt,
						LocalVersion: sysutil.InstalledVersion(p.Name),
					}
					if d := sysutil.StrToTimestamp(p.PushedAt); d > last_pushed {
						pp.KcpVersion = kcpVersion(pp.Name)
					}
					mx.Lock()
					db.Add(pp)
					mx.Unlock()
				}
			}(p)
		}
		if end {
			break
		}
	}
	wg.Wait()
	return
}

//Read remote PKGBUILD to get version.
func kcpVersion(app string) string {
	if b, e := Pkgbuild(app); e == nil {
		if v, ok := pkgbuild.Version(b); ok {
			return v
		}
	}
	return "" //UNKNOWN_VERSION
}

//Pkgbuild returns the PKGBUILD of the given repo.
func Pkgbuild(app string) ([]byte, error) {
	b, e := launchRequest(false, "", URL_PKGBUILD, app)
	if e == nil && string(b) == "Not Found" {
		e = errors.New(tr(MSG_NOT_FOUND))
	}
	return b, e
}

//PkgbuildProto returns a PKGBUILD prototype.
func PkgbuildProto() ([]byte, error) {
	b, e := launchRequest(false, "", PKGBUILDPROTO)
	if e == nil && string(b) == "Not Found" {
		e = errors.New(tr(MSG_NOT_FOUND))
	}
	return b, e
}

//List returns the complete list of repos in KCP.
func List(debug bool, pushed_at int64) (db *kcpdb.Database, e error) {
	return listPkg(SEARCH_ALL, debug, pushed_at)
}

//Exists checks the existence of the given repo.
func Exists(app string) bool {
	_, e := Pkgbuild(app)
	return e == nil
}

//Clone clones the given KCP's repo.
func Clone(app string) error {
	if _, e := os.Stat(app); e == nil {
		return fmt.Errorf(tr(MSG_PATH_EXISTS), app)
	}
	if !Exists(app) {
		return errors.New(tr(MSG_NOT_FOUND))
	}
	return sysutil.LaunchCommand("git", "clone", fmt.Sprintf(URL_REPO, app))
}
