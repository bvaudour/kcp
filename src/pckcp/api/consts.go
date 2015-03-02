package main

const (
	E int = iota
	W
	I
	N
)

const (
	E_NOPKGBUILD   = "Current folder doesn't contain PKGBUILD!"
	E_SAVEPKGBUILD = "PKGBUILD cannot be saved!"
	I_HEADER       = "Header is clean."
	W_HEADER       = "Header was found. Do not use names of maintainers or contributors in PKGBUILD, anyone can contribute, keep the header clean from this."
	Q_HEADER       = "Remove header?"
	I_PKGREL       = "pkgrel is clean."
	W_PKGREL       = "pkgrel is different from 1. It should be the case only if build instructions are edited but not pkgver."
	Q_PKGREL       = "Reset pkgrel to 1?"
	I_ARCH         = "arch is clean."
	W_ARCH         = "arch is different from 'x86_64'. Since KaOS only supports this architecture, no other arch would be added here."
	Q_ARCH         = "Reset arch to x86_64?"
	I_URL          = "url is clean."
	W_URL          = "No url specified."
	Q_URL          = "Add url?"
	W_EMPTYVAR     = "var '%s' is empty."
	Q_EMPTYVAR     = "Remove var '%s'?"
)
