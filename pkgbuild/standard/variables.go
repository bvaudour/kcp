package standard

import (
	"git.kaosx.ovh/benjamin/collection"
)

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
	CKSUMS       = "cksums"
	MD5SUMS      = "md5sums"
	SHA1SUMS     = "sha1sums"
	SHA256SUMS   = "sha256sums"
	B2SUMS       = "b2sums"
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
		CKSUMS,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
		B2SUMS,
	}
	lrequired = []string{
		PKGNAME,
		PKGVER,
		PKGREL,
		PKGDESC,
		ARCH,
		URL,
		LICENSE,
	}
	lsums = []string{
		CKSUMS,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
		B2SUMS,
	}

	vset      = collection.NewSet(vlist...)
	vrequired = collection.NewSet(lrequired...)
	varray    = collection.NewSet(
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
		CKSUMS,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
		B2SUMS,
	)
	vquoted = collection.NewSet(
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
		CKSUMS,
		MD5SUMS,
		SHA1SUMS,
		SHA256SUMS,
		B2SUMS,
	)
	vsums = collection.NewSet(lsums...)
)

// GetVariables returns the list of the well-known variables.
func GetVariables() []string {
	out := make([]string, len(vlist))
	copy(out, vlist)
	return out
}

// GetRequiredVariables returns the list of the required variables.
func GetRequiredVariables() []string {
	out := make([]string, len(lrequired))
	copy(out, lrequired)
	return out
}

// GetChecksumsVariables returns the list of the variables that can be used for the check sum.
func GetChecksumsVariables() []string {
	out := make([]string, len(lsums))
	copy(out, lsums)
	return out
}

// IsStandardVariable returns true if the name is a well-known variable.
func IsStandardVariable(name string) bool { return vset.Contains(name) }

// IsRequiredVariable returns true if the name is a required variable.
func IsRequiredVariable(name string) bool { return vrequired.Contains(name) }

// IsRequiredVariable returns true if the name is a variable of type array.
func IsArrayVariable(name string) bool { return varray.Contains(name) }

// IsQuotedVariable returns true if the value(s) of named variable should be quoted.
func IsQuotedVariable(name string) bool { return vquoted.Contains(name) || !vset.Contains(name) }

// IsChecksumsVariable returns true if the name is a check sum.
func IsChecksumsVariable(name string) bool { return vsums.Contains(name) }
