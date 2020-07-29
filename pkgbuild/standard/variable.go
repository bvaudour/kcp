package standard

const (
	PKGBASE      = "pkgbase"
	PKGNAME      = "pkgname"
	PKGVER       = "pkgver"
	PKGREL       = "pkgrel"
	EPOCH        = "epoch"
	PKGDESC      = "pkgdesc"
	ARCH         = "arch"
	URL          = "url"
	LICENSE      = "license"
	GROUPS       = "groups"
	DEPENDS      = "depends"
	MAKEDEPENDS  = "makedepends"
	CHECKDEPENDS = "checkdepends"
	OPTDEPENDS   = "optdepends"
	PROVIDES     = "provides"
	CONFLICTS    = "conflicts"
	REPLACES     = "replaces"
	BACKUP       = "backup"
	OPTIONS      = "options"
	INSTALL      = "install"
	CHANGELOG    = "changelog"
	SOURCE       = "source"
	NOEXTRACT    = "noextract"
	MD5SUMS      = "md5sums"
	SHA1SUMS     = "sha1sums"
	SHA256SUMS   = "sha256sums"
)

var (
	vlist = []string{
		PKGBASE,
		PKGNAME,
		PKGVER,
		PKGREL,
		EPOCH,
		PKGDESC,
		ARCH,
		URL,
		LICENSE,
		GROUPS,
		DEPENDS,
		MAKEDEPENDS,
		CHECKDEPENDS,
		OPTDEPENDS,
		PROVIDES,
		CONFLICTS,
		REPLACES,
		BACKUP,
		OPTIONS,
		INSTALL,
		CHANGELOG,
		SOURCE,
		NOEXTRACT,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
	}

	vset      = listToSet(vlist...)
	vrequired = listToSet(
		PKGNAME,
		PKGVER,
		PKGREL,
		PKGDESC,
		ARCH,
		URL,
		LICENSE,
	)
	varray = listToSet(
		PKGBASE,
		ARCH,
		LICENSE,
		GROUPS,
		DEPENDS,
		MAKEDEPENDS,
		CHECKDEPENDS,
		OPTDEPENDS,
		PROVIDES,
		CONFLICTS,
		REPLACES,
		BACKUP,
		OPTIONS,
		SOURCE,
		NOEXTRACT,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
	)
	vquoted = listToSet(
		PKGDESC,
		ARCH,
		URL,
		LICENSE,
		GROUPS,
		DEPENDS,
		MAKEDEPENDS,
		CHECKDEPENDS,
		OPTDEPENDS,
		PROVIDES,
		CONFLICTS,
		REPLACES,
		BACKUP,
		INSTALL,
		CHANGELOG,
		SOURCE,
		NOEXTRACT,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
	)
)

//GetVariables returns the list of the well-known variables.
func GetVariables() []string {
	out := make([]string, len(vlist))
	copy(out, vlist)
	return out
}

//IsStandardVariable returns true if the name is a well-known variable.
func IsStandardVariable(name string) bool {
	return vset[name]
}

//IsRequiredVariable returns true if the name is a required variable.
func IsRequiredVariable(name string) bool {
	return vrequired[name]
}

//IsRequiredVariable returns true if the name is a variable of type array.
func IsArrayVariable(name string) bool {
	return varray[name]
}

//IsQuotedVariable returns true if the value(s) of named variable should be quoted.
func IsQuotedVariable(name string) bool {
	return vquoted[name] || !vset[name]
}
