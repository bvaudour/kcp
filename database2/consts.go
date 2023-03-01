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

	baseUrl             = "https://api.github.com/orgs"
	baseRawURL          = "https://raw.githubusercontent.com/%s/%s/%s/PKGBUILD"
	baseOrganizationURL = baseUrl + "/%s"
	baseReposURL        = baseOrganizationURL + "/repos?page=%d&per_page=%d"
	acceptHeader        = "application/vnd.github.v3+json"

	defaultLimit    = 100
	defaultRoutines = 50
)
