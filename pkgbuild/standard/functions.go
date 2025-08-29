package standard

import (
	"git.kaosx.ovh/benjamin/collection"
)

const (
	PREPARE = "prepare"
	BUILD   = "build"
	CHECK   = "check"
	PACKAGE = "package"
)

var (
	flist = []string{
		PREPARE,
		BUILD,
		CHECK,
		PACKAGE,
	}
	lfrequired = []string{
		PACKAGE,
	}
	fset      = collection.NewSet(flist...)
	frequired = collection.NewSet(lfrequired...)
)

// GetFunctions returns the list of standard function names.
func GetFunctions() []string {
	out := make([]string, len(flist))
	copy(out, flist)
	return out
}

// GetRequiredFunctions returns the list of the required functions.
func GetRequiredFunctions() []string {
	out := make([]string, len(lfrequired))
	copy(out, lfrequired)

	return out
}

// IsStandardFunction returns true if the name is a well-known function.
func IsStandardFunction(name string) bool {
	return fset.Contains(name)
}

// IsRequiredFunction returns true if the name is a required function.
func IsRequiredFunction(name string) bool {
	return frequired.Contains(name)
}
