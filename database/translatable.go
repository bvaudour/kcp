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

	errPathExists                     = "Dir %s already exists!"
	errFailedGetPKGBUILDForNewPackage = "Failed to get PKGBUILD for new package %s: %v"
	errFailedGetPKGBUILD              = "Failed to get PKGBUILD for %s: %v"
	errCountHeader                    = "Could not parse X-Total-Count header: %v"
	errCountHeaderNotFound            = "X-Total-Count not found in response"

	msgAdded   = "%d entries added!"
	msgDeleted = "%d entries deleted!"
	msgUpdated = "%d entries updated!"
)
