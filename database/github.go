package database

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"codeberg.org/bvaudour/kcp/common"
)

// GithubConnector implements the Connector interface for a Github instance.
type GithubConnector struct {
	organization string
	token        string
	auth         *common.BasicAuth
}

// githubOrg represents organization details from the Github API.
type githubOrg struct {
	PublicRepos int `json:"public_repos"`
}

// NewGithubConnector creates a new GithubConnector.
// The auth parameter is optional and can be a token (1 value)
// or a username and password (2 values) for basic authentication.
func NewGithubConnector(org string, auth ...string) *GithubConnector {
	gc := &GithubConnector{
		organization: org,
	}
	if len(auth) == 1 {
		gc.token = auth[0]
	} else if len(auth) == 2 {
		gc.auth = &common.BasicAuth{Username: auth[0], Password: auth[1]}
	}
	return gc
}

// doRequest handles the common logic for making requests to the Github API.
func (gc *GithubConnector) doRequest(method, path string, query url.Values) (io.Reader, http.Header, error) {
	requestURL := fmt.Sprintf("https://api.github.com%s", path)

	header := http.Header{}
	header.Set("Accept", "application/vnd.github.v3+json")

	ctx := common.Context{
		Method: method,
		Header: header,
		Query:  query,
	}

	if gc.token != "" {
		header.Set("Authorization", "token "+gc.token)
	} else if gc.auth != nil {
		ctx.BasicAuth = gc.auth
	}

	return common.Request(requestURL, ctx)
}

// CountPublcRepos counts the number of public repositories in the organization.
func (gc *GithubConnector) CountPublcRepos() (int, error) {
	path := fmt.Sprintf("/orgs/%s", gc.organization)
	responseBody, _, err := gc.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		return 0, err
	}

	var org githubOrg
	if err := json.Unmarshal(bodyBytes, &org); err != nil {
		return 0, err
	}

	return org.PublicRepos, nil
}

// GetPage retrieves a paginated list of packages from the organization.
func (gc *GithubConnector) GetPage(page, limit int) ([]Package, error) {
	path := fmt.Sprintf("/orgs/%s/repos", gc.organization)
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("per_page", strconv.Itoa(limit))
	query.Set("sort", "pushed")

	responseBody, _, err := gc.doRequest(http.MethodGet, path, query)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	var packages []Package
	if err := json.Unmarshal(bodyBytes, &packages); err != nil {
		return nil, err
	}

	for i := range packages {
		packages[i].PkgbuildUrl = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/PKGBUILD", gc.organization, packages[i].Name, packages[i].Branch)
	}

	return packages, nil
}
