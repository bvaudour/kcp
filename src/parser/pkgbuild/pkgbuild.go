package pkgbuild

import (
	"io/ioutil"
	"strings"
)

func Parse(file string) (*Pkgbuild, error) {
	if b, e := ioutil.ReadFile(file); e != nil {
		return NewPkgbuild(), e
	} else {
		p := NewPkgbuild()
		p.Parse(strings.Split(string(b), "\n"))
		return p, e
	}
}

func Unparse(p *Pkgbuild, file string) error {
	b := make([]byte, 0)
	for _, l := range p.Lines() {
		b = append(b, []byte(l)...)
		b = append(b, '\n')
	}
	return ioutil.WriteFile(file, b, 0644)
}
