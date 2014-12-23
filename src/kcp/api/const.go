package api

// Needed keys for requests
const (
	NAME          = "name"
	DESCRIPTION   = "description"
	STARS         = "stargazers_count"
	ITEMS         = "items"
	MESSAGE       = "message"
	DOCUMENTATION = "documentation_url"
)

// Needed keys for database
const (
	DB_NAME         = "name"
	DB_DESCRIPTION  = "description"
	DB_STARS        = "stars"
	DB_LOCALVERSION = "localversion"
	DB_KCPVERSION   = "kcpversion"
)

// Needed URLs for requests
const (
	HEADER       = "application/vnd.github.v3+json"
	HEADERMATCH  = "application/vnd.github.v3.text-match+json"
	SEARCH_ALL   = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100&%s"
	SEARCH_ALLST = "https://api.github.com/orgs/KaOS-Community-Packages/repos?page=%d&per_page=100+stars:>=1&%s"
	SEARCH_APP   = "https://api.github.com/search/repositories?q=%v+user:KaOS-Community-Packages+fork:true&%s"
	URL_REPO     = "https://github.com/KaOS-Community-Packages/%v.git"
	URL_PKGBUILD = "https://raw.githubusercontent.com/KaOS-Community-Packages/%v/master/PKGBUILD"
	APP_ID       = "&client_id=11f5f3d9dab26c7fff24"
	SECRET_ID    = "&client_secret=bb456e9fa4e2d0fe2df9e194974c98c2f9133ff5"
	IDENT        = APP_ID + SECRET_ID
)

// Messages
const (
	MSG_NOPACKAGE       = "No package found"
	MSG_NOROOT          = "Don't launch this program as root!"
	MSG_DIREXISTS       = "Dir %s already exists!"
	MSG_ONLYONEINSTANCE = "Another instance of kcp is running!"
	MSG_INTERRUPT       = "Interrupt by user..."
	MSG_EDIT            = "Do you want to edit PKGBUILD?"
	MSG_UNKNOWN         = "Unknown error!"
	MSG_NOT_FOUND       = "Package not found!"
	MSG_ENTRIES_UPDATED = "%d entries updated!"
)

// Other constants
const (
	DEFAULT_EDITOR     = "vim"
	UNKNOWN_VERSION    = "<unknown>"
	INSTALLED_VERSION  = "[installed]"
	INSTALLED_VERSIONF = "[installed: %s]"
	KCP_LOCK           = "kcp.lock"
	KCP_DB             = ".kcp.json"
	LOCALE_DIR         = "/usr/share/locale"
)
