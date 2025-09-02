package database

import (
	"codeberg.org/bvaudour/kcp/common"
)

// Connector is an interface which defines
// the way to implement a git API server to
// get the repositories' list of an organization.
type Connector interface {
	CountPublcRepos() (int, error)
	GetPage(page, limit int) ([]Package, error)
}

// NewConnector returns the connector according to the configuration.
func NewConnector() Connector {
	auth := common.GetAuthParameters()
	if common.GitDomain == "github.com" || common.GitDomain == "" {
		return NewGithubConnector(common.Organization, auth...)
	}
	return NewForgejoConnector(common.GitDomain, common.Organization, auth...)
}
