package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Repo struct {
	Name        string `json:"path"`
	CloneHTTP   string `json:"http_url_to_repo"`
	CloneSSH    string `json:"ssh_url_to_repo"`
	Archived    bool   `json:"archived"`
	Description string `json:"description"`
}

type group struct {
	ID int `json:"id"`
}

func ListRepos(groupPath, token, baseURL string) ([]Repo, error) {
	baseURL = trimSlash(baseURL)

	// Resolve group ID
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/groups/%s", baseURL, url.PathEscape(groupPath)), nil)
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitLab API returned %d for group %q", resp.StatusCode, groupPath)
	}
	var g group
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, err
	}

	var all []Repo
	page := 1
	for {
		endpoint := fmt.Sprintf("%s/api/v4/groups/%d/projects?per_page=100&page=%d&include_subgroups=true&with_shared=false", baseURL, g.ID, page)
		req, _ := http.NewRequest("GET", endpoint, nil)
		if token != "" {
			req.Header.Set("PRIVATE-TOKEN", token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		var repos []Repo
		json.NewDecoder(resp.Body).Decode(&repos)
		resp.Body.Close()
		if len(repos) == 0 {
			break
		}
		all = append(all, repos...)
		page++
	}
	return all, nil
}

func trimSlash(s string) string {
	if len(s) > 0 && s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}
