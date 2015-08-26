package flag

import (
	"fmt"
)

// List of properties' keys
const (
	NAME int = iota
	SHORT
	LONG
	DESCRIPTION
	LONGDESCRIPTION
	SYNOPSIS
	AUTHOR
	VERSION
	ALLOWPREARGS
	ALLOWPOSTARGS
	VALUENAME
	DEFAULTVALUE
	MULTIPLEVALUES
	HIDDEN
)

// Excepted property's type following key (true if boolean, false if string)
var ktype = map[int]bool{
	NAME:            false,
	SHORT:           false,
	LONG:            false,
	DESCRIPTION:     false,
	LONGDESCRIPTION: false,
	SYNOPSIS:        false,
	AUTHOR:          false,
	VERSION:         false,
	ALLOWPREARGS:    true,
	ALLOWPOSTARGS:   true,
	VALUENAME:       false,
	DEFAULTVALUE:    false,
	MULTIPLEVALUES:  true,
	HIDDEN:          true,
}

var kparser = []int{
	NAME,
	DESCRIPTION,
	LONGDESCRIPTION,
	SYNOPSIS,
	AUTHOR,
	VERSION,
	ALLOWPREARGS,
	ALLOWPOSTARGS,
}
var kflag = []int{
	SHORT,
	LONG,
	DESCRIPTION,
	VALUENAME,
	DEFAULTVALUE,
	MULTIPLEVALUES,
	HIDDEN,
}

type property struct {
	isBool bool
	value  interface{}
}

func newProperty(t bool) *property {
	if t {
		return &property{t, false}
	}
	return &property{t, ""}
}
func (p *property) vbool() (v bool) {
	if p.isBool {
		v = p.value.(bool)
	}
	return
}
func (p *property) vstring() (v string) {
	if !p.isBool {
		v = p.value.(string)
	}
	return
}
func (p *property) set(v interface{}) error {
	switch v.(type) {
	case bool:
		if p.isBool {
			p.value = v
			return nil
		}
	case string:
		if !p.isBool {
			p.value = v
			return nil
		}
	}
	if p.isBool {
		return newErrorf(invalidValue, v, "boolean")
	}
	return newErrorf(invalidValue, v, "string")
}
func (p *property) String() string { return fmt.Sprintf("%v", p.value) }

type properties map[int]*property

func newProperties(keys []int) properties {
	p := make(properties)
	for _, k := range keys {
		p[k] = newProperty(ktype[k])
	}
	return p
}
func parserProps() properties { return newProperties(kparser) }
func flagProps() properties   { return newProperties(kflag) }
func (p properties) set(k int, v interface{}) error {
	prop, ok := p[k]
	if !ok {
		return newErrorf(unknownProperty, k)
	}
	return prop.set(v)
}
func (p properties) vbool(k int) (v bool) {
	if prop, ok := p[k]; ok {
		v = prop.vbool()
	}
	return
}
func (p properties) vstring(k int) (v string) {
	if prop, ok := p[k]; ok {
		v = prop.vstring()
	}
	return
}
func (p properties) String() string {
	out := "{\n"
	for k, v := range p {
		out += fmt.Sprintf("\t%d: %s\n", k, v.String())
	}
	out += "}"
	return out
}
