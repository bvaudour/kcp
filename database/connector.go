package database

import (
	"codeberg.org/bvaudour/kcp/common"
)

type Connector interface {
	CountPublcRepos() (int, error)
	GetPage(page, limit int) ([]Package, error)
}

func NewConnector() Connector {
	auth := common.GetAuthParameters()
	if common.GitDomain == "github.com" || common.GitDomain == "" {
		return NewGithubConnector(common.Organization, auth...)
	}
	return NewForgejoConnector(common.GitDomain, common.Organization, auth...)
}
