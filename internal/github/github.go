// Package github fetches repository listings from the GitHub API.
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type Repo struct {
	Name        string `json:"name"`
	CloneURL    string `json:"clone_url"` // HTTPS
	SSHURL      string `json:"ssh_url"`
	Archived    bool   `json:"archived"`
	Description string `json:"description"`
}

// ListRepos returns all repos for a GitHub org or user (handles pagination).
func ListRepos(org, token string) ([]Repo, error) {
	var all []Repo

	for page := 1; ; page++ {
		batch, err := fetchPage(org, token, page)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
	}

	return all, nil
}

func fetchPage(org, token string, page int) ([]Repo, error) {
	// Try as org first
	orgEndpoint := fmt.Sprintf("https://api.github.com/orgs/%s/repos", url.PathEscape(org))
	repos, err := get(orgEndpoint, token, page)
	if err == nil {
		return repos, nil
	}

	// Only fall back to user endpoint if the org endpoint returned 404
	if !isNotFound(err) {
		return nil, err // rate limit, auth error, etc. — don't mask it
	}

	userEndpoint := fmt.Sprintf("https://api.github.com/users/%s/repos", url.PathEscape(org))
	repos, err = get(userEndpoint, token, page)
	if err != nil {
		return nil, fmt.Errorf("could not find org or user %q on GitHub: %w", org, err)
	}
	return repos, nil
}

func isNotFound(err error) bool {
	return err != nil && err.Error() == "not found"
}

func get(endpoint, token string, page int) ([]Repo, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?per_page=100&page=%d", endpoint, page), nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed — check your token")
	}
	if resp.StatusCode == 403 && resp.Header.Get("X-Ratelimit-Remaining") == "0" {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Resets at: %s", resp.Header.Get("X-Ratelimit-Reset"))
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API: HTTP %d", resp.StatusCode)
	}

	var repos []Repo
	return repos, json.NewDecoder(resp.Body).Decode(&repos)
}
