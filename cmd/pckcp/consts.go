package main

//Flags informations
const (
	appLongDescription = `Provides a tool to check the validity of a PKGBUILD according to the KCP standards.

If flag -e is used, the common errors can be checked and a (potentially) valid PKGBUILD.new is created.`
	appDescription    = "Tool in command-line to manage common PKGBUILD errors"
	synopsis          = "[-h|-e|-v|-g[-c]]"
	help              = "Print this help"
	version           = "Print version"
	interactiveEdit   = "Interactive edition"
	generatePrototype = "Generate a prototype of PKGBUILD"
	cleanUseless      = "Removes the useless comments and blanks of the prototype"
)

//Messages’ templates
const (
	typeError   = "Error"
	typeWarning = "Warning"
	typeInfo    = "Info"

	errFileNotExist = "File %s does not exist."
	errMissingVar   = "Variable '%s' is missing."
	errMissingFunc  = "Function '%s' is missing."

	warnSaved            = "Modifications saved in %s!"
	warnHeader           = "Header was found. Do not use names of maintainers or contributors in PKGBUILD, anyone can contribute, keep the header clean from this."
	warnDuplicate        = "Some duplicates found:"
	warnBadType          = "Bad type declaration: '%s' is %s but it should be %s."
	warnEmpty            = "Variable '%s' is empty."
	warnPkgrel           = "pkgrel is different from 1. It should be the case only if build instructions are edited but not pkgver."
	warnArch             = "arch is different from 'x86_64'. Since KaOS only supports this architecture, no other arch would be added here."
	warnInstall          = "install: file '%s' doesn’t exist."
	warnDepends          = "Variable '%s' contains bad packages."
	warnPackageIsName    = "'%s' is the name of the package. It is useless."
	warnPackageNotInRepo = "'%s' isn't in repo neither in kcp."
	warnMissingDepends   = "Variables 'depends' and 'makedepends' are empty. You should manually check if it is not a missing."

	infoHeader      = "Header is clean."
	infoDuplicate   = "There aren’t duplicates."
	infoMissingVar  = "There aren’t missing variables."
	infoMissingFunc = "There aren’t missing functions."
	infoBadType     = "Declarations have the good type."
	infoEmpty       = "There aren’t empty variables."
	infoVarClean    = "Variable '%s' is clean."

	questionHeader      = "Remove header?"
	questionDuplicate   = "Remove duplicates?"
	questionMissingVar  = "Add variable '%s'?"
	questionAddValue    = "Set variable '%s' with (leave blank to ignore):"
	questionRemoveEmpty = "Remove variable '%s'?"
	questionPkgrel      = "Reset pkgrel to 1?"
	questionArch        = "Reset arch to x86_64?"
	questionInstall     = "Modify name of '%s' file?"
	questionInstall2    = "Type the new name (leave blank to remove install variable):"
	questionDepend      = "Modify '%s'?"
	questionTypeDepend  = "Type the new value (leave blank to remove it):"
	questionFormat      = "Format the PKGBUILD?"

	commentAddManually = "You should add it manually."
	commentFunction    = "a function"
	commentVariable    = "a variable"
	commentStringVar   = "a string variable"
	commentArrayVar    = "an array variable"
	commentInstall     = "Note that hooks provide similar functionnalities and are more powerful. For more informations: %s"
)
