package pkgbuild

import (
	"regexp"
)

// Types of data
const (
	TD_UNKNOWN int = iota
	TD_VARIABLE
	TD_BLANK
	TD_COMMENT
	TD_FUNC
)

// Types of containers
const (
	TC_UNKNOWN int = iota
	TC_HEADER
	TC_VARIABLE
	TC_UVARIABLE
	TC_FUNCTION
	TC_UFUNCTION
	TC_SFUNCTION
	TC_BLANKCOMMENT
)

// Types of unparing
const (
	TU_SINGLEVAR int = iota
	TU_SINGLEVARQ
	TU_OPTIONAL
	TU_OPTIONALQ
	TU_MULTIPLEVAR
	TU_MULTIPLEVARQ
	TU_MULTIPLELINES
	TU_LINES
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

var U_VARIABLES = map[string]int{
	PKGBASE:      TU_SINGLEVAR,
	PKGNAME:      TU_OPTIONAL,
	PKGVER:       TU_SINGLEVAR,
	PKGREL:       TU_SINGLEVAR,
	EPOCH:        TU_SINGLEVAR,
	PKGDESC:      TU_SINGLEVARQ,
	ARCH:         TU_MULTIPLEVARQ,
	URL:          TU_SINGLEVARQ,
	LICENSE:      TU_MULTIPLEVARQ,
	GROUPS:       TU_MULTIPLEVARQ,
	DEPENDS:      TU_MULTIPLEVARQ,
	MAKEDEPENDS:  TU_MULTIPLEVARQ,
	CHECKDEPENDS: TU_MULTIPLEVARQ,
	OPTDEPENDS:   TU_MULTIPLELINES,
	PROVIDES:     TU_MULTIPLEVARQ,
	CONFLICTS:    TU_MULTIPLEVARQ,
	REPLACES:     TU_MULTIPLEVARQ,
	BACKUP:       TU_MULTIPLEVARQ,
	OPTIONS:      TU_MULTIPLEVAR,
	INSTALL:      TU_SINGLEVARQ,
	CHANGELOG:    TU_SINGLEVARQ,
	SOURCE:       TU_MULTIPLELINES,
	NOEXTRACT:    TU_MULTIPLEVARQ,
	MD5SUMS:      TU_MULTIPLELINES,
	SHA1SUMS:     TU_MULTIPLELINES,
	SHA256SUMS:   TU_MULTIPLELINES,
}

// List of common functions
const (
	PREPARE = "prepare"
	BUILD   = "build"
	CHECK   = "check"
	PACKAGE = "package"
)

var L_FUNCTIONS = []string{
	PREPARE,
	BUILD,
	CHECK,
	PACKAGE,
}

// Other
const (
	HEADER  = "<header>"
	BLANK   = "<blank>"
	UNKNOWN = "<unknown>"
)

// Regexp
var R_BLANK = regexp.MustCompile(`^\s*$`)
var R_COMMENT = regexp.MustCompile(`^\s*(#.*)$`)
var R_MVAR1 = regexp.MustCompile(`^\s*(\S+)=(\(.*\))\s*$`)
var R_MVAR2 = regexp.MustCompile(`^\s*(\S+)=(\(.*)\s*$`)
var R_MVAR3 = regexp.MustCompile(`^\s*(\S+)=(.*)\s*$`)
var R_FUNCTION = regexp.MustCompile(`^\s*(\S+)\s*\(\s*\).*$`)

//var R_FUNCTION = regexp.MustCompile(`^\s*(\S+)\s*\(\s*\)\s*\{\s*$`)
