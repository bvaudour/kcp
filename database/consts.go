package database

const (
	cInstalled        = "[installed]"
	cInstalledVersion = "[installed: %s]"
	cName             = "Name"
	cVersion          = "Version"
	cDescription      = "Description"
	cArch             = "Architecture"
	cUrl              = "URL"
	cLicenses         = "Licenses"
	cProvides         = "Provides"
	cDepends          = "Depends on"
	cMakeDepends      = "Depends on (make)"
	cOptDepends       = "Optional Deps"
	cConflicts        = "Conflicts With"
	cReplaces         = "Replaces"
	cInstall          = "Install Script"
	cValidatedBy      = "Validated By"
	cYes              = "Yes"
	cNo               = "No"

	errPathExists = "Dir %s already exists!"

	msgAdded   = "%d entries added!"
	msgDeleted = "%d entries deleted!"
	msgUpdated = "%d entries updated!"
)
