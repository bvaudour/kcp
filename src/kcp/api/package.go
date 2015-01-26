package api

import (
	"fmt"
	"sort"
)

// Informations of a package
type Package struct {
	Name         string
	Description  string
	LocalVersion string
	KcpVersion   string
	Stars        int64
}

// Marshall to Json
func (p *Package) Map() map[string]interface{} {
	m := make(map[string]interface{})
	m[DB_NAME] = p.Name
	m[DB_DESCRIPTION] = p.Description
	m[DB_LOCALVERSION] = p.LocalVersion
	m[DB_KCPVERSION] = p.KcpVersion
	m[DB_STARS] = p.Stars
	return m
}

// String representation
func (p *Package) String() string {
	name := fmt.Sprintf("\033[1m%s\033[m", p.Name)
	kcpversion := fmt.Sprintf("\033[1;32m%s\033[m", p.KcpVersion)
	localversion := ""
	if p.LocalVersion != "" {
		if p.LocalVersion == p.KcpVersion {
			localversion = Translate(INSTALLED_VERSION)
		} else {
			localversion = fmt.Sprintf(Translate(INSTALLED_VERSIONF), p.LocalVersion)
		}
		localversion = fmt.Sprintf("\033[1;36m%s\033[m", localversion)
	}
	stars := fmt.Sprintf("\033[1;34m(%d)\033[m", p.Stars)
	description := ""
	if p.Description != "" {
		description = fmt.Sprintf("\n\t%s", p.Description)
	}
	return fmt.Sprintf("\033[1;35mkcp/\033[m%s %s %s %s %s", name, kcpversion, localversion, stars, description)
}

// Collection of packages
type PList []*Package
type PMap map[string]*Package
type PCollection interface {
	Add(p *Package)
	List() PList
	Map() PMap
}

// Needed methods to sort a PList
func (l PList) Len() int      { return len(l) }
func (l PList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l PList) Less(i, j int) bool {
	if l[i].Stars != l[j].Stars {
		return l[i].Stars > l[j].Stars
	}
	return l[i].Name <= l[j].Name
}
func (l PList) Sorted() PList {
	sort.Sort(l)
	return l
}

// Methods for collections
func EmptyPList() *PList {
	l := make(PList, 0)
	return &l
}
func (l *PList) Add(p *Package) { *l = append(*l, p) }
func (l *PList) List() PList    { return *l }
func (l *PList) Map() PMap {
	m := EmptyPMap()
	for _, p := range l.List() {
		m.Add(p)
	}
	return m
}

func EmptyPMap() PMap         { return make(PMap) }
func (m PMap) Add(p *Package) { m[p.Name] = p }
func (m PMap) Map() PMap      { return m }
func (m PMap) List() PList {
	l := EmptyPList()
	for _, p := range m {
		l.Add(p)
	}
	return l.List()
}
func (m PMap) Keys() []string {
	l := make([]string, len(m))
	i := 0
	for p, _ := range m {
		l[i] = p
		i++
	}
	return l
}

func EmptyPCollection(t_lst bool) PCollection {
	if t_lst {
		return EmptyPList()
	}
	return EmptyPMap()
}
