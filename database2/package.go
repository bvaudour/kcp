package database

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bvaudour/kcp/color"
	"github.com/bvaudour/kcp/common"
	"github.com/bvaudour/kcp/pkgbuild"
	"github.com/bvaudour/kcp/pkgbuild/standard"
)

// Package stores informations about a package.
type Package struct {
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         time.Time `json:"pushed_at"`
	RepoUrl          string    `json:"html_url"`
	CloneUrl         string    `json:"clone_url"`
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
	noChange         bool
}

// GetLocaleVersion searches the installed version of
// the package. If the package is not installed
// it returns an empty string.
func (p Package) GetLocaleVersion() string {
	return common.InstalledVersion(p.Name)
}

// GetPKGBUILD reads and parses the remote PKGBUILD
// from the github organization URL.
func (p Package) GetPKGBUID(debug ...bool) (pkg *pkgbuild.PKGBUILD, err error) {
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

func (p *Package) updateFromPKGBUILD(file *pkgbuild.PKGBUILD) {
	p.RepoVersion = file.GetFullVersion()
	p.HasInstallScript = false

	for n, v := range file.GetArrayValues() {
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

func (p *Package) updateFromPackage(p2 Package) {
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

// String returns the string representation of a package.
func (p Package) String() string {
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
			fmt.Fprint(&w, color.Cyan.Colorize(common.Tr(labelInstalled)))
		} else {
			fmt.Fprint(&w, color.Cyan.Format(common.Tr(labelInstalledVersion), p.LocalVersion))
		}
	}

	fmt.Fprintln(&w, color.Blue.Format(" (%d)", p.Stars))
	fmt.Fprint(&w, "\t", p.Description)

	return w.String()
}

// Detail returns detailled informations of the package.
func (p Package) Detail() string {
	labels, values := make([]string, 14), make([]string, 14)

	labels[0], values[0] = common.Tr(labelName), p.Name
	labels[1], values[1] = common.Tr(labelVersion), p.RepoVersion
	labels[2], values[2] = common.Tr(labelDescription), p.Description
	labels[3], values[3] = common.Tr(labelArch), strings.Join(p.Arch, " ")
	labels[4], values[4] = common.Tr(labelUrl), p.Url
	labels[5], values[5] = common.Tr(labelLicenses), strings.Join(p.Licenses, " ")
	labels[6], values[6] = common.Tr(labelProvides), strings.Join(p.Provides, " ")
	labels[7], values[7] = common.Tr(labelDepends), strings.Join(p.Depends, " ")
	labels[8], values[8] = common.Tr(labelMakeDepends), strings.Join(p.MakeDepends, " ")
	labels[9], values[9] = common.Tr(labelOptDepends), strings.Join(p.OptDepends, " ")
	labels[10], values[10] = common.Tr(labelConflicts), strings.Join(p.Conflicts, " ")
	labels[11], values[11] = common.Tr(labelReplaces), strings.Join(p.Replaces, " ")
	labels[12], values[12] = common.Tr(labelInstall), common.Tr(labelNo)

	if p.HasInstallScript {
		values[12] = common.Tr(labelYes)
	}
	labels[13], values[13] = common.Tr(labelValidatedBy), p.ValidatedBy

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

// Clone clone the git repo corresponding to the package
// on the given dir.
// if ssh it clones using ssh.
func (p Package) Clone(dir string, ssh bool) (fullDir string, err error) {
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

type FilterFunc func(Package) bool
type SorterFunc func(Package, Package) int

// NewFilter aggregates multiple filter funcs in one filter func.
func NewFilter(filters ...FilterFunc) FilterFunc {
	return func(p Package) bool {
		for _, f := range filters {
			if !f(p) {
				return false
			}
		}
		return true
	}
}

// NewSorter aggregates multiple sort funcs in one sort func.
func NewSorter(sorters ...SorterFunc) SorterFunc {
	return func(p1, p2 Package) int {
		for _, s := range sorters {
			if c := s(p1, p2); c != 0 {
				return c
			}
		}
		return 0
	}
}

// Packages is a list of packages.
type Packages []Package

// ToSet returns a set of the list.
func (pl Packages) ToSet() PackageSet {
	ps := make(PackageSet)

	for _, p := range pl {
		ps[p.Name] = p
	}

	return ps
}

// Push append the given entries to the list.
func (pl *Packages) Push(packages ...Package) {
	*pl = append(*pl, packages...)
}

// Remove removes the given entries from the list.
func (pl *Packages) Remove(packages ...Package) {
	ps := Packages(packages).ToSet()
	np := pl.Filter(func(p Package) bool { return !ps.Contains(p.Name) })
	*pl = np
}

// Filter returns a list which contains all packages
// matching the filters.
func (pl Packages) Filter(filters ...FilterFunc) (result Packages) {
	f := NewFilter(filters...)

	for _, p := range pl {
		if f(p) {
			result.Push(p)
		}
	}

	return result
}

// Sort sorts the list using the given criterias
// and returns it.
func (pl Packages) Sort(sorters ...SorterFunc) Packages {
	s := NewSorter(sorters...)
	less := func(i, j int) bool {
		return s(pl[i], pl[j]) <= 0
	}

	sort.Slice(pl, less)

	return pl
}

// Get returns the package with the given name.
// If no package found, ok is false.
func (pl Packages) Get(name string) (result Package, ok bool) {
	for _, p := range pl {
		if ok = p.Name == name; ok {
			return p, ok
		}
	}

	return
}

// SearchBroken returns packages which have at least
// one depend missing on the offical repo or on KCP.
func (pl Packages) SearchBroken() (broken []string) {
	available := make(map[string]bool)
	for _, p := range pl {
		available[p.Name] = true
	}

	buffer := make(chan string, len(pl))
	checkBroken := func(d string) {
		if available[d] {
			return
		}

		result, _ := common.GetOutputCommand("pacman", "-Si", d)
		if len(result) > 0 {
			buffer <- d
		}
	}

	go (func() {
		for {
			b, ok := <-buffer
			if !ok {
				return
			}
			broken = append(broken, b)
		}
	})()

	for _, p := range pl {
		for _, d := range p.Depends {
			go checkBroken(d)
		}
		for _, d := range p.MakeDepends {
			go (func(d string) {
				i := strings.Index(d, ":")
				if i >= 0 {
					d = d[:i]
				}
				checkBroken(d)
			})(d)
		}
		for _, d := range p.OptDepends {
			go checkBroken(d)
		}
	}

	close(buffer)

	return
}

// Names returns the list of the packages’ names.
func (pl Packages) Names() []string {
	names := make([]string, len(pl))

	for i, p := range pl {
		names[i] = p.Name
	}

	return names
}

// String returns the string representation of the list.
func (pl Packages) String() string {
	out := make([]string, len(pl))

	for i, p := range pl {
		out[i] = p.String()
	}

	return strings.Join(out, "\n")
}

// PackageSet is list of packages mapped by name
type PackageSet map[string]Package

// Contains checks if the set contains a package with the given name.
func (ps PackageSet) Contains(name string) (ok bool) {
	_, ok = ps[name]

	return
}

// Push adds all given packages to the set.
func (ps PackageSet) Push(packages ...Package) {
	for _, p := range packages {
		ps[p.Name] = p
	}
}

// Remove removes all packages with the given names.
func (ps PackageSet) Remove(names ...string) {
	for _, n := range names {
		delete(ps, n)
	}
}

// ToList converts the set to a list of packages.
func (ps PackageSet) ToList() (pl Packages) {
	for _, p := range ps {
		pl.Push(p)
	}

	return
}

// Filter returns a list which contains all packages
// matching the filters.
func (ps PackageSet) Filter(filters ...FilterFunc) (pl Packages) {
	f := NewFilter(filters...)

	for _, p := range ps {
		if f(p) {
			pl.Push(p)
		}
	}

	return
}

// Sort sorts the packages using the given criterias
// and returns them.
func (ps PackageSet) Sort(sorters ...SorterFunc) Packages {
	return ps.ToList().Sort(sorters...)
}

// Names returns the list of the packages’ names.
func (ps PackageSet) Names() []string {
	var names []string

	for n := range ps {
		names = append(names, n)
	}

	return names
}

func SortByName(p1, p2 Package) int {
	return strings.Compare(p1.Name, p2.Name)
}

func SortByStar(p1, p2 Package) int {
	c := p2.Stars - p1.Stars

	if c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}
	return c
}

func FilterInstalled(p Package) bool {
	return p.LocalVersion != ""
}

func FilterOutdated(p Package) bool {
	return FilterInstalled(p) && p.LocalVersion != p.RepoVersion
}

func FilterStarred(p Package) bool {
	return p.Stars > 0
}
