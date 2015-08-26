package pkgbuild

//List of standard variables
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

var L_VARIABLES = []string{
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

var L_NEEDED = []string{
	PKGNAME,
	PKGVER,
	PKGREL,
	PKGDESC,
	ARCH,
	URL,
	LICENSE,
}

//List of common functions
const (
	PREPARE = "prepare"
	BUILD   = "build"
	CHECK   = "check"
	PACKAGE = "package"
)

var L_FUNCTIONS = []string{
	PKGVER,
	PREPARE,
	BUILD,
	CHECK,
	PACKAGE,
}

//List of unparsing types
const (
	uSingleVar = iota
	uSingleVarQ
	uOptional
	uOptionalQ
	uMultipleVar
	uMultipleVarQ
	uMultipleLines
	uLines
)

var uVariables = map[string]int{
	PKGBASE:      uSingleVar,
	PKGNAME:      uOptional,
	PKGVER:       uSingleVar,
	PKGREL:       uSingleVar,
	EPOCH:        uSingleVar,
	PKGDESC:      uSingleVarQ,
	ARCH:         uMultipleVarQ,
	URL:          uSingleVarQ,
	LICENSE:      uMultipleVarQ,
	GROUPS:       uMultipleVarQ,
	DEPENDS:      uMultipleVarQ,
	MAKEDEPENDS:  uMultipleVarQ,
	CHECKDEPENDS: uMultipleVarQ,
	OPTDEPENDS:   uMultipleVarQ,
	PROVIDES:     uMultipleVarQ,
	CONFLICTS:    uMultipleVarQ,
	REPLACES:     uMultipleVarQ,
	BACKUP:       uMultipleVarQ,
	OPTIONS:      uMultipleVarQ,
	INSTALL:      uSingleVarQ,
	CHANGELOG:    uSingleVarQ,
	SOURCE:       uMultipleLines,
	NOEXTRACT:    uMultipleVarQ,
	MD5SUMS:      uMultipleLines,
	SHA1SUMS:     uMultipleLines,
	SHA256SUMS:   uMultipleLines,
}
