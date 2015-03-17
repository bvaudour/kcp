package pkgbuild

// Type of data
type DataType int

const (
	DT_UNKNOWN DataType = iota
	DT_BLANK
	DT_COMMENT
	DT_VARIABLE
	DT_FUNCTION
)

// Type of container
type BlockType int

const (
	BT_NONE BlockType = iota
	BT_UNKNOWN
	BT_HEADER
	BT_VARIABLE
	BT_UVARIABLE
	BT_FUNCTION
	BT_UFUNCTION
)

// List of standard variables
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

// List of common functions
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

// Other
const (
	HEADER  = "<header>"
	UNKNOWN = "<unknown>"
)

// Types of unparsing
type UnparseType int

const (
	UT_SINGLEVAR UnparseType = iota
	UT_SINGLEVARQ
	UT_OPTIONAL
	UT_OPTIONALQ
	UT_MULTIPLEVAR
	UT_MULTIPLEVARQ
	UT_MULTIPLELINES
	UT_LINES
)

var U_VARIABLES = map[string]UnparseType{
	PKGBASE:      UT_SINGLEVAR,
	PKGNAME:      UT_OPTIONAL,
	PKGVER:       UT_SINGLEVAR,
	PKGREL:       UT_SINGLEVAR,
	EPOCH:        UT_SINGLEVAR,
	PKGDESC:      UT_SINGLEVARQ,
	ARCH:         UT_MULTIPLEVARQ,
	URL:          UT_SINGLEVARQ,
	LICENSE:      UT_MULTIPLEVARQ,
	GROUPS:       UT_MULTIPLEVARQ,
	DEPENDS:      UT_MULTIPLEVARQ,
	MAKEDEPENDS:  UT_MULTIPLEVARQ,
	CHECKDEPENDS: UT_MULTIPLEVARQ,
	OPTDEPENDS:   UT_MULTIPLELINES,
	PROVIDES:     UT_MULTIPLEVARQ,
	CONFLICTS:    UT_MULTIPLEVARQ,
	REPLACES:     UT_MULTIPLEVARQ,
	BACKUP:       UT_MULTIPLEVARQ,
	OPTIONS:      UT_MULTIPLEVAR,
	INSTALL:      UT_SINGLEVARQ,
	CHANGELOG:    UT_SINGLEVARQ,
	SOURCE:       UT_MULTIPLELINES,
	NOEXTRACT:    UT_MULTIPLEVARQ,
	MD5SUMS:      UT_MULTIPLELINES,
	SHA1SUMS:     UT_MULTIPLELINES,
	SHA256SUMS:   UT_MULTIPLELINES,
}
