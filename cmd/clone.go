package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/000Janela000/orgclone/internal/config"
	"github.com/000Janela000/orgclone/internal/git"
	gh "github.com/000Janela000/orgclone/internal/github"
	gl "github.com/000Janela000/orgclone/internal/gitlab"
	"github.com/spf13/cobra"
)

var (
	flagToken      string
	flagDest       string
	flagExclude    string
	flagSkipArch   bool
	flagSSH        bool
	flagGitLabURL  string
	flagDryRun     bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <platform> <name>",
	Short: "Clone all repos from a GitHub org or GitLab group",
	Args:  cobra.ExactArgs(2),
	RunE:  runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVarP(&flagToken, "token", "t", "", "API token (or set GITHUB_TOKEN / GITLAB_TOKEN env var)")
	cloneCmd.Flags().StringVarP(&flagDest, "dest", "d", "", "Destination folder (default: ~/Desktop/<name>)")
	cloneCmd.Flags().StringVarP(&flagExclude, "exclude", "e", "", "Comma-separated repo names to exclude")
	cloneCmd.Flags().BoolVar(&flagSkipArch, "skip-archived", false, "Skip archived repositories")
	cloneCmd.Flags().BoolVar(&flagSSH, "ssh", false, "Force SSH URLs (requires SSH key set up)")
	cloneCmd.Flags().StringVar(&flagGitLabURL, "gitlab-url", "", "Self-hosted GitLab URL")
	cloneCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "List repos without cloning")
}

type repoInfo struct {
	Name        string
	CloneURL    string
	SSHURL      string
	Archived    bool
	Description string
}

func runClone(cmd *cobra.Command, args []string) error {
	platform := strings.ToLower(args[0])
	name := args[1]

	if platform != "github" && platform != "gitlab" {
		return fmt.Errorf("platform must be 'github' or 'gitlab'")
	}

	// Resolve token
	token := flagToken
	if token == "" {
		token = config.GetToken(platform)
	}

	// Resolve dest
	dest := flagDest
	if dest == "" {
		dest = filepath.Join(config.GetDefaultDest(), name)
	} else if len(dest) >= 2 && dest[:2] == "~/" {
		home, _ := os.UserHomeDir()
		dest = filepath.Join(home, dest[2:])
	}

	// Resolve exclusions
	exclude := make(map[string]bool)
	for _, e := range config.GetExclusions(name) {
		exclude[e] = true
	}
	if flagExclude != "" {
		for _, e := range strings.Split(flagExclude, ",") {
			exclude[strings.TrimSpace(e)] = true
		}
	}

	// Fetch repos
	fmt.Printf("\nFetching repos from %s: %s\n\n", platform, name)

	var repos []repoInfo
	var err error

	if platform == "github" {
		raw, e := gh.ListRepos(name, token)
		if e != nil {
			return fmt.Errorf("failed to fetch repos: %w", e)
		}
		for _, r := range raw {
			repos = append(repos, repoInfo{r.Name, r.CloneURL, r.SSHURL, r.Archived, r.Description})
		}
	} else {
		glURL := flagGitLabURL
		if glURL == "" {
			glURL = config.GetGitLabURL()
		}
		raw, e := gl.ListRepos(name, token, glURL)
		if e != nil {
			return fmt.Errorf("failed to fetch repos: %w", e)
		}
		for _, r := range raw {
			repos = append(repos, repoInfo{r.Name, r.CloneHTTP, r.CloneSSH, r.Archived, r.Description})
		}
	}

	// Filter
	var filtered []repoInfo
	for _, r := range repos {
		if exclude[r.Name] {
			continue
		}
		if flagSkipArch && r.Archived {
			continue
		}
		filtered = append(filtered, r)
	}

	if len(filtered) == 0 {
		fmt.Println("No repos found (or all filtered out).")
		return nil
	}

	if flagDryRun {
		fmt.Printf("%-40s  %-8s  %s\n", "REPO", "ARCHIVED", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range filtered {
			arch := ""
			if r.Archived {
				arch = "yes"
			}
			desc := r.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Printf("%-40s  %-8s  %s\n", r.Name, arch, desc)
		}
		fmt.Printf("\n%d repos\n", len(filtered))
		return nil
	}

	// Clone
	if err = os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	fmt.Printf("Destination: %s\n\n", dest)

	counts := map[git.Status]int{}
	for _, r := range filtered {
		cloneURL := r.CloneURL
		if flagSSH {
			cloneURL = r.SSHURL
		} else if token != "" && strings.HasPrefix(cloneURL, "https://") {
			cloneURL = strings.Replace(cloneURL, "https://", "https://oauth2:"+token+"@", 1)
		}

		result := git.CloneOrPull(r.Name, cloneURL, dest)
		counts[result.Status]++

		switch result.Status {
		case git.Cloned:
			fmt.Printf("  + %-35s cloned\n", r.Name)
		case git.Pulled:
			fmt.Printf("  ^ %-35s updated\n", r.Name)
		case git.UpToDate:
			fmt.Printf("  - %-35s up to date\n", r.Name)
		case git.Failed:
			fmt.Printf("  x %-35s FAILED: %s\n", r.Name, result.Message)
		}
	}

	fmt.Printf("\nDone.  %d cloned  %d updated  %d up-to-date  %d failed\n",
		counts[git.Cloned], counts[git.Pulled], counts[git.UpToDate], counts[git.Failed])

	_ = err
	return nil
}
