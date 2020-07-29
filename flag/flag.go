package flag

import (
	"fmt"
	"strconv"
	"strings"
)

//Flag is a descriptor of a parsing element.
type Flag struct {
	p    Properties
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
func stringOf(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", NewError(errInvalidValue, v, typeString)
}
func boolOf(v interface{}) (bool, error) {
	if b, ok := v.(bool); ok {
		return b, nil
	}
	return false, NewError(errInvalidValue, v, typeBool)
}
func isShortFlag(v interface{}) error {
	n, e := stringOf(v)
	if e != nil {
		return e
	}
	switch {
	case len(n) == 0:
		return nil
	case len(n) == 1:
		return NewError(errInvalidFlag, flagShort, n, flagTooShort)
	case len(n) > 2:
		return NewError(errInvalidFlag, flagShort, n, flagTooLong)
	case n[0] != '-' || n[1] == '-':
		return NewError(errInvalidFlag, flagShort, n, fmt.Sprintf(flagMustBegin, "-"))
	}
	return nil
}
func isLongFlag(v interface{}) error {
	n, e := stringOf(v)
	if e != nil {
		return e
	}
	switch {
	case len(n) == 0:
		return nil
	case len(n) < 3:
		return NewError(errInvalidFlag, flagLong, n, flagTooShort)
	case n[0:2] != "--":
		return NewError(errInvalidFlag, flagLong, n, fmt.Sprintf(flagMustBegin, "--"))
	default:
		return nil
	}
}

//Set modifies the given property with the given value.
func (f *Flag) Set(k PropertyType, v interface{}) error {
	switch k {
	case Short:
		if e := isShortFlag(v); e != nil {
			return e
		}
	case Long:
		if e := isLongFlag(v); e != nil {
			return e
		}
	}
	return f.p.Set(k, v)
}

//Get returns the value of the needed property.
func (f *Flag) Get(k PropertyType) interface{} {
	return f.p.Value(k)
}

//GetString returns the string representation of the needed property.
func (f *Flag) GetString(k PropertyType) string {
	return f.p.ValueString(k)
}

//GetBool returns the boolean representation of the needed property.
func (f *Flag) GetBool(k PropertyType) bool {
	return f.p.ValueBool(k)
}

//Short returns the value of the short flag.
func (f *Flag) Short() string {
	return f.GetString(Short)
}

//Long returns the value of the long flag.
func (f *Flag) Long() string {
	return f.GetString(Long)
}

//Description returns the description of the flag.
func (f *Flag) Description() string {
	return f.GetString(Description)
}

//ValueName returns the value' name of the flag.
func (f *Flag) ValueName() string {
	return f.GetString(ValueName)
}

//DefaultValue returns the default value of the flag.
func (f *Flag) DefaultValue() string {
	return f.GetString(DefaultValue)
}

//AllowMultipleValues returns true if the flag accepts one or more values.
func (f *Flag) AllowMultipleValues() bool {
	return f.GetBool(MultipleValues)
}

//Hidden returns true if the flag shouldn't appear in the help.
func (f *Flag) Hidden() bool {
	return f.GetBool(Hidden)
}

//Parse functions
func parseBool(v *bool) func(string) error {
	return func(s string) error {
		if s != "" {
			return NewError(errUnexpectedArg, s)
		}
		*v = true
		return nil
	}
}
func parseString(v *string) func(string) error {
	return func(s string) error {
		if s == "" {
			return NewError(errNotEnoughArg)
		}
		*v = s
		return nil
	}
}
func parseChoice(v *string, c []string) func(string) error {
	return func(s string) error {
		switch {
		case s == "":
			return NewError(errNotEnoughArg)
		case !contains(c, s):
			return NewError(errUnexpectedChoiceArg, s, c)
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
func initFlag(s, l, desc string, cb func(string) error) (f *Flag, e error) {
	f = new(Flag)
	f.p = FlagProps()
	if e = f.Set(Short, s); e != nil {
		return
	}
	if e = f.Set(Long, l); e != nil {
		return
	}
	f.Set(Description, desc)
	f.f = cb
	return
}

//NewBoolFlag returns a new flag which doesn't accept args and a pointer to its value.
func NewBoolFlag(s, l, desc string) (f *Flag, v *bool, e error) {
	v = new(bool)
	f, e = initFlag(s, l, desc, parseBool(v))
	return
}

//NewStringFlag returns a new flag which accepts a string arg and a pointer to its value.
func NewStringFlag(s, l, desc, vn, df string) (f *Flag, v *string, e error) {
	v = new(string)
	f, e = initFlag(s, l, desc, parseString(v))
	if e != nil {
		return
	}
	if vn == "" {
		vn = "ARG"
	}
	f.Set(ValueName, vn)
	f.Set(DefaultValue, df)
	return
}

//NewChoiceFlag returns a new flag which accepts arg among a list of choices, and a pointer to its value.
func NewChoiceFlag(s, l, desc, df string, choices []string) (f *Flag, v *string, e error) {
	if df != "" && !contains(choices, df) {
		e = NewError(errUnexpectedChoiceArg, df, choices)
		return
	}
	v = new(string)
	f, e = initFlag(s, l, desc, parseChoice(v, choices))
	if e != nil {
		return
	}
	f.Set(ValueName, "["+strings.Join(choices, "|")+"]")
	f.Set(DefaultValue, df)
	return
}

//NewArrayFlag returns a new flag which accepts multiple string args, and a pointer to its value.
func NewArrayFlag(s, l, desc, vn string) (f *Flag, v *[]string, e error) {
	*v = make([]string, 0, 1)
	f, e = initFlag(s, l, desc, parseArray(v))
	if e != nil {
		return
	}
	if vn == "" {
		vn = "ARG..."
	}
	f.Set(ValueName, vn)
	return
}

//NewIntFlag returns a new flag which accepts an int arg and a pointer to its value.
func NewIntFlag(s, l, desc, vn string, df int) (f *Flag, v *int, e error) {
	v = new(int)
	*v = df
	f, e = initFlag(s, l, desc, parseInt(v))
	if e != nil {
		return
	}
	if vn == "" {
		vn = "ARG"
	}
	f.Set(ValueName, vn)
	return
}

// Group represents a selected flag among a group of flags.
type Group struct{ selected string }
