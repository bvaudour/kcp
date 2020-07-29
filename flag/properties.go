package flag

import (
	"encoding/json"
	"fmt"
)

//Property is a type of property.
type PropertyType int

const (
	Unknown PropertyType = iota
	Name
	Short
	Long
	Description
	LongDescription
	Synopsis
	Author
	Version
	AllowPreArs
	AllowPostArgs
	ValueName
	DefaultValue
	MultipleValues
	Hidden
)

var (
	// Excepted property's type following key (true if boolean, false if string)
	kType = map[PropertyType]bool{
		Name:            false,
		Short:           false,
		Long:            false,
		Description:     false,
		LongDescription: false,
		Synopsis:        false,
		Author:          false,
		Version:         false,
		AllowPreArs:     true,
		AllowPostArgs:   true,
		ValueName:       false,
		DefaultValue:    false,
		MultipleValues:  true,
		Hidden:          true,
	}

	kParser = []PropertyType{
		Name,
		Description,
		LongDescription,
		Synopsis,
		Author,
		Version,
		AllowPreArs,
		AllowPostArgs,
	}
	kFlag = []PropertyType{
		Short,
		Long,
		Description,
		ValueName,
		DefaultValue,
		MultipleValues,
		Hidden,
	}
)

//Property represents an internal property of a flag or a flags’ parser.
type Property struct {
	t     PropertyType
	vbool bool
	vstr  string
}

func newProperty(t PropertyType) *Property {
	return &Property{t: t}
}

//Type returns the property’s type.
func (p *Property) Type() PropertyType {
	return p.t
}

//IsBool returns true if the property’s value accept only booleans.
func (p *Property) IsBool() bool {
	return kType[p.t]
}

//IsString returns true if the property’s value accept only strings.
func (p *Property) IsString() bool {
	return !p.IsBool()
}

//ValueBool returns the boolean value of the property.
//It returns false if the property is not boolean.
func (p *Property) ValueBool() bool {
	return p.vbool
}

//ValueString returns the string value of the property.
//It returns an empty string if the property is not a string.
func (p *Property) ValueString() string {
	return p.vstr
}

//Value returns the value of the property which is either
//boolean or string depending of the type of the property.
func (p *Property) Value() interface{} {
	if p.IsBool() {
		return p.ValueBool()
	}
	return p.ValueString()
}

//String returns the string representation of the value’s property.
func (p *Property) String() string {
	if p.IsString() {
		return p.ValueString()
	}
	return fmt.Sprint(p.ValueBool())
}

//Set sets the property to the given value.
//It returns an error if the given value has not the good type.
func (p *Property) Set(v interface{}) (err error) {
	if p.IsBool() {
		p.vbool, err = boolOf(v)
	} else {
		p.vstr, err = stringOf(v)
	}
	return
}

//Properties is a set of properties
type Properties map[PropertyType]*Property

func newProperties(keys []PropertyType) Properties {
	p := make(Properties)
	for _, k := range keys {
		p[k] = newProperty(k)
	}
	return p
}

//ParserProps returns all properties supported by
//a flags’ parser.
func ParserProps() Properties {
	return newProperties(kParser)
}

//FlagProps returns all propertis supported by
//a flag.
func FlagProps() Properties {
	return newProperties(kFlag)
}

//Set set the given property to the given value.
//It returns an error if the type of value isn’t supported
//by the property’s type.
func (l Properties) Set(k PropertyType, v interface{}) error {
	if p, ok := l[k]; ok {
		return p.Set(v)
	}
	return NewError(errUnknownProperty, k)
}

//Value returns a string or a boolan of the given property
//depending of the property’s type.
//It returns nil if the property doesn’t exist.
func (l Properties) Value(k PropertyType) interface{} {
	if p, ok := l[k]; ok {
		return p.Value()
	}
	return nil
}

//ValueBool returns the boolan value of the given property.
//It returns false if the property doesn’t exist or is not a boolean.
func (l Properties) ValueBool(k PropertyType) bool {
	if p, ok := l[k]; ok {
		return p.ValueBool()
	}
	return false
}

//ValueString returns the string value of the given property.
//It returns an empty string if the property doesn’t exist or is not a string.
func (l Properties) ValueString(k PropertyType) string {
	if p, ok := l[k]; ok {
		return p.ValueString()
	}
	return ""
}

//String returns the string representation af the properties’ set.
func (l Properties) String() string {
	m := make(map[PropertyType]string)
	for k, p := range l {
		m[k] = p.String()
	}
	json, _ := json.MarshalIndent(m, "", "\t")
	return string(json)
}
