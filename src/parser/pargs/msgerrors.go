package pargs

import (
	"errors"
	"fmt"
)

func ferror(form string, e ...interface{}) error {
	return errors.New(fmt.Sprintf(form, e...))
}

func invalidValue(v interface{}, t string) error {
	return ferror("Invalid value '%v': Must be a %v!", v, t)
}

func invalidParse(v interface{}, t string) error {
	return ferror("Value '%v' cannot be parsed in %v!", v, t)
}

func unknownProperty(p int) error {
	return ferror("Unknown property: %d", p)
}

func invalidFlag(f, t, explain string) error {
	return ferror("Invalid %s flag '%s': %s!", t, f, explain)
}

func invalidShortFlag(f, explain string) error {
	return invalidFlag(f, "short", explain)
}

func invalidLongFlag(f, explain string) error {
	return invalidFlag(f, "long", explain)
}

func unexpectedFlag(f string) error {
	return ferror("Flag '%s' is already set!", f)
}

func unsupportedFlag(f string) error {
	return ferror("Unsupported flag '%s'", f)
}

func unexpectedArg(a string) error {
	return ferror("Unexpected arg: %s", a)
}

func notAllowed(a string) error {
	return ferror("%s not allowed!", a)
}

func notEnoughArg() error {
	return errors.New("Arg needed!")
}

func unexpectedChoiceArg(a string, c []string) error {
	return ferror("Arg '%s' not in %v!", a, c)
}

func mustBeAlternative(f1, f2 string) error {
	return ferror("Flag '%s' cannot be used with '%s'", f1, f2)
}

func noMultipleAllowed(f string) error {
	return ferror("Flag '%s' doesn't allow multiple arguments!", f)
}

func needRequirment(f string, req []string) error {
	return ferror("Flag '%s' needs one of this flags: %v", f, req)
}
