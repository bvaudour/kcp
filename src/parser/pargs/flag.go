package pargs

import (
	"strconv"
)

type Flag struct {
	p    properties
	f    func(string) error
	used string
}

// Internal checkers
func contains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}
func isString(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		return v.(string), nil
	default:
		return "", invalidValue(v, "string")
	}
}
func isShortFlag(v interface{}) error {
	n, e := isString(v)
	if e != nil {
		return e
	}
	switch {
	case len(n) == 0:
		return nil
	case len(n) == 1:
		return invalidShortFlag(n, "too short")
	case len(n) > 2:
		return invalidShortFlag(n, "too long")
	case n[0] != '-' || n[1] == '-':
		return invalidShortFlag(n, "must begin with '-'")
	default:
		return nil
	}
}
func isLongFlag(v interface{}) error {
	n, e := isString(v)
	if e != nil {
		return e
	}
	switch {
	case len(n) == 0:
		return nil
	case len(n) < 3:
		return invalidLongFlag(n, "too short")
	case n[0:2] != "--":
		return invalidLongFlag(n, "must begin with '--'")
	default:
		return nil
	}
}

// Flag properties
func (f *Flag) Set(k int, v interface{}) error {
	switch k {
	case SHORT:
		if e := isShortFlag(v); e != nil {
			return e
		}
	case LONG:
		if e := isLongFlag(v); e != nil {
			return e
		}
	}
	return f.p.set(k, v)
}
func (f *Flag) GetString(k int) string    { return f.p.vstring(k) }
func (f *Flag) GetBool(k int) bool        { return f.p.vbool(k) }
func (f *Flag) Short() string             { return f.GetString(SHORT) }
func (f *Flag) Long() string              { return f.GetString(LONG) }
func (f *Flag) Description() string       { return f.GetString(DESCRIPTION) }
func (f *Flag) ValueName() string         { return f.GetString(VALUENAME) }
func (f *Flag) DefaultValue() string      { return f.GetString(DEFAULTVALUE) }
func (f *Flag) AllowMultipleValues() bool { return f.GetBool(MULTIPLEVALUES) }
func (f *Flag) Hidden() bool              { return f.GetBool(HIDDEN) }

// Process creation
func fBool(v *bool) func(string) error {
	return func(s string) error {
		if s != "" {
			return unexpectedArg(s)
		}
		*v = true
		return nil
	}
}
func fString(v *string) func(string) error {
	return func(s string) error {
		if s == "" {
			return notEnoughArg()
		}
		*v = s
		return nil
	}
}
func fChoice(v *string, c []string) func(string) error {
	return func(s string) error {
		switch {
		case s == "":
			return notEnoughArg()
		case !contains(c, s):
			return unexpectedChoiceArg(s, c)
		default:
			*v = s
			return nil
		}
	}
}
func fArray(v *[]string) func(string) error {
	return func(s string) error {
		if s != "" {
			*v = append(*v, s)
		}
		return nil
	}
}
func fInt(v *int) func(string) error {
	return func(s string) error {
		var e error
		*v, e = strconv.Atoi(s)
		return e
	}
}

// Flag creation
func initf(sf, lf, d string, pr func(string) error) (f *Flag, e error) {
	f = new(Flag)
	f.p = flagProps()
	if e = f.Set(SHORT, sf); e != nil {
		return
	}
	if e = f.Set(LONG, lf); e != nil {
		return
	}
	f.Set(DESCRIPTION, d)
	f.f = pr
	return
}
func NewBoolFlag(sf, lf, d string) (f *Flag, v *bool, e error) {
	v = new(bool)
	f, e = initf(sf, lf, d, fBool(v))
	return
}
func NewStringFlag(sf, lf, d, vn, dv string) (f *Flag, v *string, e error) {
	v = new(string)
	f, e = initf(sf, lf, d, fString(v))
	if e == nil {
		if vn == "" {
			vn = "ARG"
		}
		f.Set(VALUENAME, vn)
		f.Set(DEFAULTVALUE, dv)
	}
	return
}
func NewChoiceFlag(sf, lf, d, dv string, c []string) (f *Flag, v *string, e error) {
	if dv != "" && !contains(c, dv) {
		e = unexpectedChoiceArg(dv, c)
		return
	}
	v = new(string)
	f, e = initf(sf, lf, d, fChoice(v, c))
	if e == nil {
		vn := "["
		for i, v := range c {
			if i > 0 {
				vn += "|"
			}
			vn += v
		}
		vn += "]"
		f.Set(VALUENAME, vn)
		f.Set(DEFAULTVALUE, dv)
	}
	return
}
func NewArrayFlag(sf, lf, d, vn string) (f *Flag, v *[]string, e error) {
	*v = make([]string, 0, 1)
	f, e = initf(sf, lf, d, fArray(v))
	if e == nil {
		if vn == "" {
			vn = "ARG..."
		}
		f.Set(VALUENAME, vn)
	}
	return
}
func NewIntFlag(sf, lf, d, vn string, dv int) (f *Flag, v *int, e error) {
	v = new(int)
	*v = dv
	f, e = initf(sf, lf, d, fInt(v))
	if e == nil {
		if vn == "" {
			vn = "ARG"
		}
		f.Set(VALUENAME, vn)
	}
	return
}

// Groups
type Group struct {
	selected string
}
