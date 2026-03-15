// Package config loads settings from ~/.orgclone.yml and environment variables.
// CLI flags always take priority over config file values.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultDest     string               `yaml:"default_dest"`
	DefaultPlatform string               `yaml:"default_platform"`
	GitHub          platformConfig       `yaml:"github"`
	GitLab          gitlabConfig         `yaml:"gitlab"`
	Orgs            map[string]orgConfig `yaml:"orgs"`
}

type platformConfig struct {
	Token string `yaml:"token"`
}

type gitlabConfig struct {
	Token string `yaml:"token"`
	URL   string `yaml:"url"`
}

type orgConfig struct {
	Exclude []string `yaml:"exclude"`
}

func load() Config {
	var cfg Config
	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".orgclone.yml"))
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

// GetToken returns the API token for a platform.
// Priority: env var → config file.
func GetToken(platform string) string {
	envVars := map[string]string{"github": "GITHUB_TOKEN", "gitlab": "GITLAB_TOKEN"}
	if t := os.Getenv(envVars[platform]); t != "" {
		return t
	}
	cfg := load()
	if platform == "github" {
		return cfg.GitHub.Token
	}
	return cfg.GitLab.Token
}

// GetGitLabURL returns the GitLab base URL (defaults to gitlab.com).
func GetGitLabURL() string {
	if u := load().GitLab.URL; u != "" {
		return u
	}
	return "https://gitlab.com"
}

// GetDefaultDest returns the base destination folder (defaults to ~/Desktop).
func GetDefaultDest() string {
	if d := load().DefaultDest; d != "" {
		return expandHome(d)
	}
	return filepath.Join(homeDir(), "Desktop")
}

// GetDefaultPlatform returns the default platform (github unless changed via `orgclone default`).
func GetDefaultPlatform() string {
	if p := load().DefaultPlatform; p != "" {
		return p
	}
	return "github"
}

// GetExclusions returns repo names to skip for a given org/group.
func GetExclusions(name string) []string {
	return load().Orgs[name].Exclude
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

func expandHome(path string) string { //nolint
	if len(path) >= 2 && path[:2] == "~/" {
		return filepath.Join(homeDir(), path[2:])
	}
	return path
}
