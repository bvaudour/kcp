package standard

const (
	PREPARE = "prepare"
	BUILD   = "build"
	CHECK   = "check"
	PACKAGE = "package"
)

func listToSet(l ...string) map[string]bool {
	m := make(map[string]bool)
	for _, e := range l {
		m[e] = true
	}
	return m
}

var (
	flist = []string{
		PREPARE,
		BUILD,
		CHECK,
		PACKAGE,
	}
	fset      = listToSet(flist...)
	frequired = listToSet(
		PACKAGE,
	)
)

//GetFunctions returns the list of standard function names.
func GetFunctions() []string {
	out := make([]string, len(flist))
	copy(out, flist)
	return out
}

//IsStandardFunction returns true if the name is a well-known function.
func IsStandardFunction(name string) bool {
	return fset[name]
}

//IsRequiredFunction returns true if the name is a required function.
func IsRequiredFunction(name string) bool {
	return frequired[name]
}
