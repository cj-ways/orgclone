// Package gitlab fetches repository listings from the GitLab API.
package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type Repo struct {
	Name        string `json:"path"`
	CloneHTTP   string `json:"http_url_to_repo"`
	CloneSSH    string `json:"ssh_url_to_repo"`
	Archived    bool   `json:"archived"`
	Description string `json:"description"`
}

// ListRepos returns all projects in a GitLab group, including subgroups.
func ListRepos(group, token, baseURL string) ([]Repo, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	groupID, err := resolveGroupID(group, token, baseURL)
	if err != nil {
		return nil, err
	}

	var all []Repo
	for page := 1; ; page++ {
		batch, err := fetchPage(groupID, token, baseURL, page)
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

func resolveGroupID(group, token, baseURL string) (int, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/groups/%s", baseURL, url.PathEscape(group)), nil)
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return 0, fmt.Errorf("authentication failed — check your token")
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GitLab API: HTTP %d for group %q", resp.StatusCode, group)
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("parsing group response: %w", err)
	}
	return result.ID, nil
}

func fetchPage(groupID int, token, baseURL string, page int) ([]Repo, error) {
	endpoint := fmt.Sprintf(
		"%s/api/v4/groups/%d/projects?per_page=100&page=%d&include_subgroups=true&with_shared=false",
		baseURL, groupID, page,
	)

	req, _ := http.NewRequest("GET", endpoint, nil)
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed — check your token")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitLab API: HTTP %d", resp.StatusCode)
	}

	var repos []Repo
	return repos, json.NewDecoder(resp.Body).Decode(&repos)
}
