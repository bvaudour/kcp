package database

const (
	labelInstalled        = "[installed]"
	labelInstalledVersion = "[installed: %s]"
	labelName             = "Name"
	labelVersion          = "Version"
	labelDescription      = "Description"
	labelArch             = "Architecture"
	labelUrl              = "URL"
	labelLicenses         = "Licenses"
	labelProvides         = "Provides"
	labelDepends          = "Depends on"
	labelMakeDepends      = "Depends on (make)"
	labelOptDepends       = "Optional Deps"
	labelConflicts        = "Conflicts With"
	labelReplaces         = "Replaces"
	labelInstall          = "Install Script"
	labelValidatedBy      = "Validated By"
	labelYes              = "Yes"
	labelNo               = "No"

	errPathExists = "Dir %s already exists!"

	msgAdded   = "%d entries added!"
	msgDeleted = "%d entries deleted!"
	msgUpdated = "%d entries updated!"

	defaultLimit    = 100
	defaultRoutines = 150
)
