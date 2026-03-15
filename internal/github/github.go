// Package github fetches repository listings from the GitHub API.
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
	// Try as org first, fall back to user if 404
	for _, endpoint := range []string{
		fmt.Sprintf("https://api.github.com/orgs/%s/repos", url.PathEscape(org)),
		fmt.Sprintf("https://api.github.com/users/%s/repos", url.PathEscape(org)),
	} {
		repos, err := get(endpoint, token, page)
		if err == nil {
			return repos, nil
		}
	}
	return nil, fmt.Errorf("could not find org or user %q on GitHub", org)
}

func get(endpoint, token string, page int) ([]Repo, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?per_page=100&page=%d", endpoint, page), nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API: HTTP %d", resp.StatusCode)
	}

	var repos []Repo
	return repos, json.NewDecoder(resp.Body).Decode(&repos)
}
