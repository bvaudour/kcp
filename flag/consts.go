package flag

import (
	"github.com/bvaudour/kcp/common"
)

//Templates for errors' messages.
const (
	errInvalidFlag         = "Invalid %s flag '%s': %s!"
	errInvalidValue        = "Invalid value '%v': Must be a %s!"
	errMustBeAlternative   = "Flag '%s' cannot be used with '%s'"
	errNeedRequirment      = "Flag '%s' needs one of this flags: %v"
	errNoMultipleAllowed   = "Flag '%s' doesn't allow multiple arguments!"
	errNotAllowed          = "%s not allowed!"
	errNotEnoughArg        = "Arg needed!"
	errUnexpectedArg       = "Unexpected arg: %s"
	errUnexpectedChoiceArg = "Arg '%s' not in %v!"
	errUnexpectedFlag      = "Flag '%s' is already set!"
	errUnknownProperty     = "Unknown property: %d"
	errUnsupportedFlag     = "Unsupported flag '%s'"

	typeString = "string"
	typeBool   = "bool"

	flagShort     = "short"
	flagLong      = "long"
	flagMustBegin = "must begin with '%s'"
	flagTooShort  = "too short"
	flagTooLong   = "too long"

	usage = "Usage:"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func NewError(err string, args ...interface{}) error {
	return Error(common.Tr(err, args...))
}
