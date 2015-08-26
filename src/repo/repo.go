//Package repo provides useful function to manage requests with the KaOS Community Packages.
package repo

import (
	"errors"
	"fmt"
	"gettext"
	"io/ioutil"
	"kcpdb"
	"net/http"
	"os"
	"parser/json"
	"parser/pkgbuild"
	"sync"
	"sysutil"
)

//Needed URLs for requests
const (
	HEADER       = "application/vnd.github.v3+json"
	SEARCH_ALL   = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100&%s"
	URL_REPO     = "https://github.com/KaOS-Community-Packages/%s.git"
	URL_PKGBUILD = "https://raw.githubusercontent.com/KaOS-Community-Packages/%s/master/PKGBUILD"
	APP_ID       = "&client_id=11f5f3d9dab26c7fff24"
	SECRET_ID    = "&client_secret=bb456e9fa4e2d0fe2df9e194974c98c2f9133ff5"
	IDENT        = APP_ID + SECRET_ID
)

//Json keys of github API
const (
	NAME          = "name"
	DESCRIPTION   = "description"
	STARS         = "stargazers_count"
	ITEMS         = "items"
	MESSAGE       = "message"
	DOCUMENTATION = "documentation_url"
)

//Messages
const (
	MSG_NOT_FOUND   = "Package not found!"
	MSG_UNKNOWN     = "Unknown error!"
	MSG_PATH_EXISTS = "Dir %s already exists!"
	UNKNOWN_VERSION = "<unknown>"
)

var tr = gettext.Gettext

//Conversions
func o2p(o json.Object) (p *kcpdb.Package) {
	if s, e := o.GetString(NAME); e == nil {
		p = new(kcpdb.Package)
		p.Name = s
		p.Description, _ = o.GetString(DESCRIPTION)
		p.Stars, _ = o.GetInt64(STARS)
	}
	return
}
func o2e(o json.Object) error {
	msg, e1 := o.GetString(MESSAGE)
	doc, e2 := o.GetString(DOCUMENTATION)
	if e1 != nil || e2 != nil {
		return errors.New(tr(MSG_UNKNOWN))
	}
	return fmt.Errorf("%s\n%s\n", msg, doc)
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

func listPkg(search string, debug bool) (db kcpdb.Database, e error) {
	db = kcpdb.New()
	var wg sync.WaitGroup
	var mx = new(sync.Mutex)
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
		obj, err := json.ArrayObjectBytes(b)
		if err != nil {
			end = true
			o, _ := json.ObjectBytes(b)
			e = o2e(o)
			return
		}
		if len(obj) == 0 {
			end = true
			break
		}
		for _, o := range obj {
			go func(o json.Object) {
				wg.Add(1)
				defer wg.Done()
				p := o2p(o)
				if p != nil {
					p.LocalVersion = sysutil.InstalledVersion(p.Name)
					p.KcpVersion = kcpVersion(p.Name)
					mx.Lock()
					db.Add(p)
					mx.Unlock()
				}
			}(o)
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
	return UNKNOWN_VERSION
}

//Pkgbuild returns the PKGBUILD of the given repo.
func Pkgbuild(app string) ([]byte, error) {
	b, e := launchRequest(false, "", URL_PKGBUILD, app)
	if e == nil && string(b) == "Not Found" {
		e = errors.New(tr(MSG_NOT_FOUND))
	}
	return b, e
}

//List returns the complete list of repos in KCP.
func List(debug bool) (db kcpdb.Database, e error) { return listPkg(SEARCH_ALL, debug) }

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
