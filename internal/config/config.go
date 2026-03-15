package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OrgConfig struct {
	Exclude []string `yaml:"exclude"`
}

type Config struct {
	DefaultDest string               `yaml:"default_dest"`
	GitHub      struct{ Token string `yaml:"token"` } `yaml:"github"`
	GitLab      struct {
		Token string `yaml:"token"`
		URL   string `yaml:"url"`
	} `yaml:"gitlab"`
	Orgs map[string]OrgConfig `yaml:"orgs"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".orgclone.yml")
}

func Load() Config {
	var cfg Config
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

func GetToken(platform string) string {
	cfg := Load()
	switch platform {
	case "github":
		if t := os.Getenv("GITHUB_TOKEN"); t != "" {
			return t
		}
		return cfg.GitHub.Token
	case "gitlab":
		if t := os.Getenv("GITLAB_TOKEN"); t != "" {
			return t
		}
		return cfg.GitLab.Token
	}
	return ""
}

func GetGitLabURL() string {
	cfg := Load()
	if cfg.GitLab.URL != "" {
		return cfg.GitLab.URL
	}
	return "https://gitlab.com"
}

func GetDefaultDest() string {
	cfg := Load()
	if cfg.DefaultDest != "" {
		home, _ := os.UserHomeDir()
		if len(cfg.DefaultDest) >= 2 && cfg.DefaultDest[:2] == "~/" {
			return filepath.Join(home, cfg.DefaultDest[2:])
		}
		return cfg.DefaultDest
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Desktop")
}

func GetExclusions(name string) []string {
	cfg := Load()
	if org, ok := cfg.Orgs[name]; ok {
		return org.Exclude
	}
	return nil
}
