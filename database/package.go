package database

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bvaudour/kcp/color"
	"github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/pkgbuild"
	"github.com/bvaudour/kcp/pkgbuild/standard"
)

//Package stores informations about a package.
type Package struct {
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         time.Time `json:"pushed_at"`
	RepoUrl          string    `json:"html_url"`
	CloneUrl         string    `json:"git_url"`
	SshUrl           string    `json:"ssh_url"`
	PkgbuildUrl      string    `json:"pkgbuild_url"`
	Stars            int       `json:"stargazers_count"`
	Branch           string    `json:"default_branch"`
	LocalVersion     string    `json:"local_version"`
	RepoVersion      string    `json:"remote_version"`
	Arch             []string  `json:"architectures"`
	Url              string    `json:"upstream_url"`
	Provides         []string  `json:"provides"`
	Depends          []string  `json:"depends"`
	OptDepends       []string  `json:"opt_depends"`
	MakeDepends      []string  `json:"make_depends"`
	Conflicts        []string  `json:"conflicts"`
	Replaces         []string  `json:"replaces"`
	Licenses         []string  `json:"licenses"`
	ValidatedBy      string    `json:"validated_by"`
	HasInstallScript bool      `json:"has_install_script"`
}

func (p *Package) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":             p.Name,
		"description":      p.Description,
		"createdAt":        p.CreatedAt,
		"updatedAt":        p.UpdatedAt,
		"PushedAt":         p.PushedAt,
		"repoUrl":          p.RepoUrl,
		"cloneUrl":         p.CloneUrl,
		"sshUrl":           p.SshUrl,
		"pkgbuildUrl":      p.PkgbuildUrl,
		"stars":            p.Stars,
		"localVersion":     p.LocalVersion,
		"repoVersion":      p.RepoVersion,
		"arch":             p.Arch,
		"url":              p.Url,
		"provides":         p.Provides,
		"depends":          p.Depends,
		"optDepends":       p.OptDepends,
		"makeDepends":      p.MakeDepends,
		"conflicts":        p.Conflicts,
		"replaces":         p.Replaces,
		"licenses":         p.Licenses,
		"validatedBy":      p.ValidatedBy,
		"hasInstallScript": p.HasInstallScript,
	}
}

func s2m(l []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range l {
		m[s] = true
	}
	return m
}

func eqSlices(l1, l2 []string) bool {
	m1, m2 := s2m(l1), s2m(l2)
	if len(m1) != len(m2) {
		return false
	}
	for s := range m1 {
		if !m2[s] {
			return false
		}
	}
	return true
}

//Eq checks if the both packages contains similar informations.
func (p1 *Package) Eq(p2 *Package) bool {
	m1, m2 := p1.toMap(), p2.toMap()
	for k, v1 := range m1 {
		v2 := m2[k]
		switch v1.(type) {
		case []string:
			if !eqSlices(v1.([]string), v2.([]string)) {
				return false
			}
		case time.Time:
			if !v1.(time.Time).Equal(v2.(time.Time)) {
				return false
			}
		default:
			if v1 != v2 {
				return false
			}
		}
	}
	return true
}

//GetLocaleVersion searches the installed version of
//the package. If the package is not installed
//it returns an empty string.
func (p *Package) GetLocaleVersion() string {
	return common.InstalledVersion(p.Name)
}

//GetPKGBUILD reads and parses the remote PKGBUILD
//from the github organization URL.
func (p *Package) GetPKGBUID(debug ...bool) (pkg *pkgbuild.PKGBUILD, err error) {
	url := p.PkgbuildUrl
	printDebug := len(debug) > 0 && debug[0]
	var buf io.Reader
	if buf, err = execRequest(url, ctx{}); err != nil {
		if printDebug {
			fmt.Fprintf(os.Stderr, "%s %s\n", color.Red.Format("[Error: %s]", err), url)
		}
		return
	}
	return pkgbuild.DecodeVars(buf)
}

func (p *Package) udpateFromPKGBUILD(pkg *pkgbuild.PKGBUILD) {
	p.RepoVersion = pkg.GetFullVersion()
	p.HasInstallScript = false
	for n, v := range pkg.GetArrayValues() {
		switch n {
		case standard.ARCH:
			p.Arch = v
		case standard.URL:
			p.Url = ""
			if len(v) > 0 {
				p.Url = v[0]
			}
		case standard.PROVIDES:
			p.Provides = v
		case standard.DEPENDS:
			p.Depends = v
		case standard.OPTDEPENDS:
			p.OptDepends = v
		case standard.MAKEDEPENDS:
			p.MakeDepends = v
		case standard.CONFLICTS:
			p.Conflicts = v
		case standard.REPLACES:
			p.Replaces = v
		case standard.MD5SUMS:
			p.ValidatedBy = "MD5"
		case standard.SHA1SUMS:
			p.ValidatedBy = "SHA-1"
		case standard.SHA256SUMS:
			p.ValidatedBy = "SHA-256"
		case standard.LICENSE:
			p.Licenses = v
		case standard.INSTALL:
			p.HasInstallScript = true
		}
	}
}

func (p *Package) updateFromPackage(p2 *Package) {
	p.RepoVersion = p2.RepoVersion
	p.Arch = p2.Arch
	p.Url = p2.Url
	p.Provides = p2.Provides
	p.Depends = p2.Depends
	p.OptDepends = p2.OptDepends
	p.MakeDepends = p2.MakeDepends
	p.Conflicts = p2.Conflicts
	p.Replaces = p2.Replaces
	p.ValidatedBy = p2.ValidatedBy
	p.HasInstallScript = p2.HasInstallScript
	p.Licenses = p2.Licenses
}

//String returns the string representation of a package.
func (p *Package) String() string {
	var w strings.Builder
	fmt.Fprint(
		&w,
		color.Magenta.Colorize("kcp/"),
		color.NoColor.Colorize(p.Name),
		" ",
		color.Green.Colorize(p.RepoVersion),
	)
	if p.LocalVersion != "" {
		fmt.Fprint(&w, " ")
		if p.LocalVersion == p.RepoVersion {
			fmt.Fprint(&w, color.Cyan.Colorize(cInstalled))
		} else {
			fmt.Fprint(&w, color.Cyan.Format(cInstalledVersion, p.LocalVersion))
		}
	}
	fmt.Fprintln(&w, color.Blue.Format(" (%d)", p.Stars))
	fmt.Fprint(&w, "\t", p.Description)
	return w.String()
}

//Detail returns detailled informations of the package.
func (p *Package) Detail() string {
	labels, values := make([]string, 14), make([]string, 14)

	labels[0], values[0] = common.Tr(cName), p.Name
	labels[1], values[1] = common.Tr(cVersion), p.RepoVersion
	labels[2], values[2] = common.Tr(cDescription), p.Description
	labels[3], values[3] = common.Tr(cArch), strings.Join(p.Arch, " ")
	labels[4], values[4] = common.Tr(cUrl), p.Url
	labels[5], values[5] = common.Tr(cLicenses), strings.Join(p.Licenses, " ")
	labels[6], values[6] = common.Tr(cProvides), strings.Join(p.Provides, " ")
	labels[7], values[7] = common.Tr(cDepends), strings.Join(p.Depends, " ")
	labels[8], values[8] = common.Tr(cMakeDepends), strings.Join(p.MakeDepends, " ")
	labels[9], values[9] = common.Tr(cOptDepends), strings.Join(p.OptDepends, " ")
	labels[10], values[10] = common.Tr(cConflicts), strings.Join(p.Conflicts, " ")
	labels[11], values[11] = common.Tr(cReplaces), strings.Join(p.Replaces, " ")
	labels[12], values[12] = common.Tr(cInstall), common.Tr(cNo)
	if p.HasInstallScript {
		values[12] = common.Tr(cYes)
	}
	labels[13], values[13] = common.Tr(cValidatedBy), p.ValidatedBy
	s := 0
	for _, l := range labels {
		sl := utf8.RuneCountInString(l)
		if sl > s {
			s = sl
		}
	}
	result := make([]string, len(labels))
	for i, l := range labels {
		v := values[i]
		if v == "" {
			v = "--"
		}
		sep := strings.Repeat(" ", s-utf8.RuneCountInString(l))
		result[i] = fmt.Sprintf("%s%s : %s", l, sep, v)
	}
	return strings.Join(result, "\n")
}

//Clone clone the git repo corresponding to the package
//on the given dir.
//if ssh it clones using ssh.
func (p *Package) Clone(dir string, ssh bool) (fullDir string, err error) {
	fullDir = path.Join(dir, p.Name)
	if common.FileExists(fullDir) {
		err = errors.New(common.Tr(errPathExists, fullDir))
		return
	}
	if err = os.Chdir(dir); err != nil {
		return
	}
	url := p.CloneUrl
	if ssh {
		url = p.SshUrl
	}
	err = common.LaunchCommand("git", "clone", url)
	return
}

type PackageFunc func(*Package)
type FilterFunc func(*Package) bool
type SorterFunc func(*Package, *Package) int

//NewFilter aggregates multiple filter funcs in one filter func.
func NewFilter(filters ...FilterFunc) FilterFunc {
	return func(p *Package) bool {
		for _, f := range filters {
			if !f(p) {
				return false
			}
		}
		return true
	}
}

//NewSorter aggregates multiple sort funcs in one sort func.
func NewSorter(sorters ...SorterFunc) SorterFunc {
	return func(p1, p2 *Package) int {
		for _, s := range sorters {
			if c := s(p1, p2); c != 0 {
				return c
			}
		}
		return 0
	}
}

//Packages is a list of packages.
type Packages []*Package

//Iterator returns an object to loop at the list.
func (pl Packages) Iterator() <-chan *Package {
	ch := make(chan *Package)
	go (func() {
		defer close(ch)
		for _, p := range pl {
			ch <- p
		}
	})()
	return ch
}

//Iterate applies the given callback to all
//entries of the list.
func (pl Packages) Iterate(cb PackageFunc) {
	for p := range pl.Iterator() {
		cb(p)
	}
}

//ToSet returns a set of the list.
func (pl Packages) ToSet() *PackageSet {
	ps := NewPackageSet()
	pl.Iterate(func(p *Package) { ps.packages[p.Name] = p })
	return ps
}

//Push append the given entries to the list.
func (pl *Packages) Push(packages ...*Package) {
	*pl = append(*pl, packages...)
}

//Remove removes the given entries from the list.
func (pl *Packages) Remove(packages ...*Package) {
	ps := Packages(packages).ToSet()
	np := pl.Filter(func(p *Package) bool { return !ps.Contains(p.Name) })
	*pl = np
}

//Filter returns a list which contains all packages
//matching the filters.
func (pl Packages) Filter(filters ...FilterFunc) Packages {
	var result Packages
	f := NewFilter(filters...)
	pl.Iterate(func(p *Package) {
		if f(p) {
			result = append(result, p)
		}
	})
	return result
}

//Sort sorts the list using the given criterias
//and returns it.
func (pl Packages) Sort(sorters ...SorterFunc) Packages {
	s := NewSorter(sorters...)
	less := func(i, j int) bool {
		return s(pl[i], pl[j]) <= 0
	}
	sort.Slice(pl, less)
	return pl
}

//Get returns the package with the given name.
//If no package found, ok is false.
func (pl Packages) Get(name string) (p *Package, ok bool) {
	for pp := range pl.Iterator() {
		if ok = pp.Name == name; ok {
			p = pp
			return
		}
	}
	return
}

//Names returns the list of the packages’ names.
func (pl Packages) Names() []string {
	names := make([]string, 0, len(pl))
	pl.Iterate(func(p *Package) { names = append(names, p.Name) })
	return names
}

//String returns the string representation of the list.
func (pl Packages) String() string {
	out := make([]string, len(pl))
	for i, p := range pl {
		out[i] = p.String()
	}
	return strings.Join(out, "\n")
}

//PackageSet is a safe-thread
//to manipulating packages informations.
type PackageSet struct {
	sync.RWMutex
	packages map[string]*Package
}

//NewPackageSet returns an empty package set.
func NewPackageSet() *PackageSet {
	return &PackageSet{
		packages: make(map[string]*Package),
	}
}

//Iterator returns an object to loop all entries of the set.
func (ps *PackageSet) Iterator() <-chan *Package {
	ch := make(chan *Package)
	go (func() {
		defer close(ch)
		for _, p := range ps.packages {
			ch <- p
		}
	})()
	return ch
}

//Iterate applies the given callback to all
//entries of the set.
func (ps *PackageSet) Iterate(cb func(*Package)) {
	ps.Lock()
	defer ps.Unlock()
	for p := range ps.Iterator() {
		cb(p)
	}
}

//Get returns the package with the given name.
//If no package found, ok is false.
func (ps *PackageSet) Get(name string) (p *Package, ok bool) {
	ps.Lock()
	defer ps.Unlock()
	p, ok = ps.packages[name]
	return
}

//Contains checks if the set contains a package with the given name.
func (ps *PackageSet) Contains(name string) bool {
	_, ok := ps.Get(name)
	return ok
}

//Add adds all given packages to the set.
func (ps *PackageSet) Add(packages ...*Package) {
	ps.Lock()
	defer ps.Unlock()
	for _, p := range packages {
		ps.packages[p.Name] = p
	}
}

//Remove removes all packages with the given names.
func (ps *PackageSet) Remove(names ...string) {
	ps.Lock()
	defer ps.Unlock()
	for _, n := range names {
		delete(ps.packages, n)
	}
}

//ToList converts the set to a list of packages.
func (ps *PackageSet) ToList() Packages {
	var pl Packages
	ps.Iterate(func(p *Package) { pl = append(pl, p) })
	return pl
}

//Filter returns a list which contains all packages
//matching the filters.
func (ps *PackageSet) Filter(filters ...FilterFunc) Packages {
	return ps.ToList().Filter(filters...)
}

//Sort sorts the packages using the given criterias
//and returns them.
func (ps *PackageSet) Sort(sorters ...SorterFunc) Packages {
	return ps.ToList().Sort(sorters...)
}

//Names returns the list of the packages’ names.
func (ps *PackageSet) Names() []string {
	var names []string
	ps.Lock()
	defer ps.Unlock()
	for n := range ps.packages {
		names = append(names, n)
	}
	return names
}

func SortByName(p1, p2 *Package) int {
	return strings.Compare(p1.Name, p2.Name)
}

func SortByStar(p1, p2 *Package) int {
	c := p2.Stars - p1.Stars
	if c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}
	return c
}

func FilterInstalled(p *Package) bool {
	return p.LocalVersion != ""
}

func FilterOutdated(p *Package) bool {
	return FilterInstalled(p) && p.LocalVersion != p.RepoVersion
}

func FilterStarred(p *Package) bool {
	return p.Stars > 0
}
