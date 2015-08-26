package flag

import (
	"errors"
	"fmt"
	"gettext"
)

//Templates for errors' messages.
const (
	invalidFlag         = "Invalid %s flag '%s': %s!"
	invalidValue        = "Invalid value '%v': Must be a %s!"
	mustBeAlternative   = "Flag '%s' cannot be used with '%s'"
	needRequirment      = "Flag '%s' needs one of this flags: %v"
	noMultipleAllowed   = "Flag '%s' doesn't allow multiple arguments!"
	notAllowed          = "%s not allowed!"
	notEnoughArg        = "Arg needed!"
	unexpectedArg       = "Unexpected arg: %s"
	unexpectedChoiceArg = "Arg '%s' not in %v!"
	unexpectedFlag      = "Flag '%s' is already set!"
	unknownProperty     = "Unknown property: %d"
	unsupportedFlag     = "Unsupported flag '%s'"
)

func newErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(gettext.Gettext(format), args...)
}
func newError(err string) error { return errors.New(gettext.Gettext(err)) }
