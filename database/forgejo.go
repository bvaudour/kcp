package database

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"codeberg.org/bvaudour/kcp/common"
)

// ForgejoConnector implements the Connector interface for a Forgejo instance.
type ForgejoConnector struct {
	host         string
	organization string
	token        string
	auth         *common.BasicAuth
}

// NewForgejoConnector creates a new ForgejoConnector.
// The token is optional.
func NewForgejoConnector(domain, org string, auth ...string) *ForgejoConnector {
	fc := &ForgejoConnector{
		host:         "https://" + domain,
		organization: org,
	}
	if len(auth) == 1 {
		fc.token = auth[0]
	} else if len(auth) == 2 {
		fc.auth = &common.BasicAuth{Username: auth[0], Password: auth[1]}
	}
	return fc
}

// forgejoRepo represents a repository as returned by the Forgejo API.
type forgejoRepo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CloneURL    string    `json:"clone_url"`
	HTMLURL     string    `json:"html_url"`
	SSHURL      string    `json:"ssh_url"`
	Stars       int       `json:"stars_count"`
	Branch      string    `json:"default_branch"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// doRequest handles the common logic for making requests to the Forgejo API.
func (fc *ForgejoConnector) doRequest(method string, query url.Values) (io.Reader, http.Header, error) {
	requestURL := fmt.Sprintf("%s/api/v1/orgs/%s/repos", fc.host, fc.organization)

	header := http.Header{}
	header.Set("accept", "application/json")

	if fc.token != "" {
		header.Set("Authorization", "token "+fc.token)
	}

	return common.Request(requestURL, common.Context{
		Method: method,
		Header: header,
		Query:  query,
	})
}

// CountPublcRepos counts the number of public repositories in the organization.
func (fc *ForgejoConnector) CountPublcRepos() (int, error) {
	query := url.Values{}
	query.Set("limit", "1")

	_, responseHeader, err := fc.doRequest(http.MethodHead, query)
	if err != nil {
		return 0, err
	}

	totalCountStr := responseHeader.Get("X-Total-Count")
	if totalCountStr == "" {
		return 0, fmt.Errorf("X-Total-Count header not found in response")
	}

	totalCount, err := strconv.Atoi(totalCountStr)
	if err != nil {
		return 0, fmt.Errorf("could not parse X-Total-Count header: %w", err)
	}

	return totalCount, nil
}

// GetPage retrieves a paginated list of packages from the organization.
func (fc *ForgejoConnector) GetPage(page, limit int) ([]Package, error) {
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(limit))

	responseBody, _, err := fc.doRequest(http.MethodGet, query)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	var repos []forgejoRepo
	if err := json.Unmarshal(bodyBytes, &repos); err != nil {
		return nil, err
	}

	packages := make([]Package, len(repos))
	for i, repo := range repos {
		packages[i] = Package{
			Name:        repo.Name,
			Description: repo.Description,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
			PushedAt:    repo.UpdatedAt, // PushedAt is not available in Forgejo, using UpdatedAt
			RepoUrl:     repo.HTMLURL,
			CloneUrl:    repo.CloneURL,
			SshUrl:      repo.SSHURL,
			PkgbuildUrl: fmt.Sprintf("%s/raw/branch/%s/PKGBUILD", repo.HTMLURL, repo.Branch),
			Stars:       repo.Stars,
			Branch:      repo.Branch,
			RepoVersion: "", // Version is not available from this endpoint.
		}
	}

	return packages, nil
}
