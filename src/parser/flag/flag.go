package flag

import (
	"strconv"
	"strings"
)

//Flag is a descriptor of a parsing element.
type Flag struct {
	p    properties
	f    func(string) error
	used string
}

//Internal checkers
func contains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}
func isString(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", newErrorf(invalidValue, v, "string")
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
		return newErrorf(invalidFlag, "short", n, "too short")
	case len(n) > 2:
		return newErrorf(invalidFlag, "short", n, "too long")
	case n[0] != '-' || n[1] == '-':
		return newErrorf(invalidFlag, "short", n, "must begin with '-'")
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
		return newErrorf(invalidFlag, "long", n, "too short")
	case n[0:2] != "--":
		return newErrorf(invalidFlag, "long", n, "must begin with '--'")
	default:
		return nil
	}
}

//Set modifies the given property with the given value.
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

//GetString returns the string representation of the needed property.
func (f *Flag) GetString(k int) string { return f.p.vstring(k) }

//GetBool returns the boolean representation of the needed property.
func (f *Flag) GetBool(k int) bool { return f.p.vbool(k) }

//Short returns the value of the short flag.
func (f *Flag) Short() string { return f.GetString(SHORT) }

//Long returns the value of the long flag.
func (f *Flag) Long() string { return f.GetString(LONG) }

//Description returns the description of the flag.
func (f *Flag) Description() string { return f.GetString(DESCRIPTION) }

//ValueName returns the value' name of the flag.
func (f *Flag) ValueName() string { return f.GetString(VALUENAME) }

//DefaultValue returns the default value of the flag.
func (f *Flag) DefaultValue() string { return f.GetString(DEFAULTVALUE) }

//AllowMultipleValues returns true if the flag accepts one or more values.
func (f *Flag) AllowMultipleValues() bool { return f.GetBool(MULTIPLEVALUES) }

//Hidden returns true if the flag shouldn't appear in the help.
func (f *Flag) Hidden() bool { return f.GetBool(HIDDEN) }

//Parse functions
func parseBool(v *bool) func(string) error {
	return func(s string) error {
		if s != "" {
			return newErrorf(unexpectedArg, s)
		}
		*v = true
		return nil
	}
}
func parseString(v *string) func(string) error {
	return func(s string) error {
		if s == "" {
			return newError(notEnoughArg)
		}
		*v = s
		return nil
	}
}
func parseChoice(v *string, c []string) func(string) error {
	return func(s string) error {
		switch {
		case s == "":
			return newError(notEnoughArg)
		case !contains(c, s):
			return newErrorf(unexpectedChoiceArg, s, c)
		default:
			*v = s
			return nil
		}
	}
}
func parseArray(v *[]string) func(string) error {
	return func(s string) error {
		if s != "" {
			*v = append(*v, s)
		}
		return nil
	}
}
func parseInt(v *int) func(string) error {
	return func(s string) error {
		var e error
		*v, e = strconv.Atoi(s)
		return e
	}
}

//Flag initialization
func initFlag(short, long, description string, parse func(string) error) (f *Flag, e error) {
	f = new(Flag)
	f.p = flagProps()
	if e = f.Set(SHORT, short); e != nil {
		return
	}
	if e = f.Set(LONG, long); e != nil {
		return
	}
	f.Set(DESCRIPTION, description)
	f.f = parse
	return
}

//NewBoolFlag returns a new flag which doesn't accept args and a pointer to its value.
func NewBoolFlag(short, long, description string) (f *Flag, v *bool, e error) {
	v = new(bool)
	f, e = initFlag(short, long, description, parseBool(v))
	return
}

//NewStringFlag returns a new flag which accepts a string arg and a pointer to its value.
func NewStringFlag(short, long, description, valueName, defaultValue string) (f *Flag, v *string, e error) {
	v = new(string)
	f, e = initFlag(short, long, description, parseString(v))
	if e != nil {
		return
	}
	if valueName == "" {
		valueName = "ARG"
	}
	f.Set(VALUENAME, valueName)
	f.Set(DEFAULTVALUE, defaultValue)
	return
}

//NewChoiceFlag returns a new flag which accepts arg among a list of choices, and a pointer to its value.
func NewChoiceFlag(short, long, description, defaultValue string, choices []string) (f *Flag, v *string, e error) {
	if defaultValue != "" && !contains(choices, defaultValue) {
		e = newErrorf(unexpectedChoiceArg, defaultValue, choices)
		return
	}
	v = new(string)
	f, e = initFlag(short, long, description, parseChoice(v, choices))
	if e != nil {
		return
	}
	f.Set(VALUENAME, "["+strings.Join(choices, "|")+"]")
	f.Set(DEFAULTVALUE, defaultValue)
	return
}

//NewArrayFlag returns a new flag which accepts multiple string args, and a pointer to its value.
func NewArrayFlag(short, long, description, valueName string) (f *Flag, v *[]string, e error) {
	*v = make([]string, 0, 1)
	f, e = initFlag(short, long, description, parseArray(v))
	if e != nil {
		return
	}
	if valueName == "" {
		valueName = "ARG..."
	}
	f.Set(VALUENAME, valueName)
	return
}

//NewIntFlag returns a new flag which accepts an int arg and a pointer to its value.
func NewIntFlag(short, long, description, valueName string, defaultValue int) (f *Flag, v *int, e error) {
	v = new(int)
	*v = defaultValue
	f, e = initFlag(short, long, description, parseInt(v))
	if e != nil {
		return
	}
	if valueName == "" {
		valueName = "ARG"
	}
	f.Set(VALUENAME, valueName)
	return
}

// Group represents a selected flag among a group of flags.
type Group struct{ selected string }

