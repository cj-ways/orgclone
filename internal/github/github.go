package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Repo struct {
	Name        string `json:"name"`
	CloneURL    string `json:"clone_url"`
	SSHURL      string `json:"ssh_url"`
	Archived    bool   `json:"archived"`
	Description string `json:"description"`
}

func ListRepos(org, token string) ([]Repo, error) {
	var all []Repo
	page := 1

	for {
		endpoint := fmt.Sprintf("https://api.github.com/orgs/%s/repos", url.PathEscape(org))
		repos, err := fetchPage(endpoint, token, page)
		if err != nil {
			// Try as user
			endpoint = fmt.Sprintf("https://api.github.com/users/%s/repos", url.PathEscape(org))
			repos, err = fetchPage(endpoint, token, page)
			if err != nil {
				return nil, err
			}
		}
		if len(repos) == 0 {
			break
		}
		all = append(all, repos...)
		page++
	}
	return all, nil
}

func fetchPage(endpoint, token string, page int) ([]Repo, error) {
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
		return nil, fmt.Errorf("not found: %s", endpoint)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	return repos, nil
}
