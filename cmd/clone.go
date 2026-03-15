package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/cj-ways/orgclone/internal/config"
	"github.com/cj-ways/orgclone/internal/git"
	gh "github.com/cj-ways/orgclone/internal/github"
	gl "github.com/cj-ways/orgclone/internal/gitlab"
	"github.com/spf13/cobra"
)

var (
	flagToken    string
	flagDest     string
	flagExclude  string
	flagSkipArch bool
	flagDryRun   bool
	flagPick     bool
	flagGitLab   bool
	flagGitLabURL string
)

var cloneCmd = &cobra.Command{
	Use:     "clone <name>",
	Short:   "Clone all repos from an org or group",
	GroupID: "core",
	Long: `Clone all repos from a GitHub org or GitLab group.

Defaults to GitHub. Use --gitlab to target GitLab instead,
or change the default permanently with: orgclone default gitlab.

Running it again on an already-cloned folder pulls updates on all repos.`,
	Example: `  orgclone clone my-org
  orgclone clone my-group --gitlab
  orgclone clone my-org --pick
  orgclone clone my-org --exclude old-repo,scratch
  orgclone clone my-org --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVarP(&flagToken, "token", "t", "", "API token (or set GITHUB_TOKEN / GITLAB_TOKEN env var)")
	cloneCmd.Flags().StringVarP(&flagDest, "dest", "d", "", "Destination folder (default: ~/Desktop/<name>)")
	cloneCmd.Flags().StringVarP(&flagExclude, "exclude", "e", "", "Comma-separated repo names to skip")
	cloneCmd.Flags().BoolVar(&flagSkipArch, "skip-archived", false, "Skip archived repositories")
	cloneCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "List repos without cloning")
	cloneCmd.Flags().BoolVar(&flagPick, "pick", false, "Interactively select which repos to clone")
	cloneCmd.Flags().BoolVar(&flagGitLab, "gitlab", false, "Use GitLab instead of GitHub")
	cloneCmd.Flags().StringVar(&flagGitLabURL, "gitlab-url", "", "Self-hosted GitLab URL")
}

type repoInfo struct {
	Name        string
	CloneURL    string
	Archived    bool
	Description string
}

func runClone(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Resolve platform: --gitlab flag > config default > "github"
	platform := config.GetDefaultPlatform()
	if flagGitLab {
		platform = "gitlab"
	}

	// Resolve token (CLI flag > env var > config file)
	token := flagToken
	if token == "" {
		token = config.GetToken(platform)
	}

	// Resolve destination
	dest := flagDest
	if dest == "" {
		dest = filepath.Join(config.GetDefaultDest(), name)
	} else if strings.HasPrefix(dest, "~/") {
		home, _ := os.UserHomeDir()
		dest = filepath.Join(home, dest[2:])
	}

	// Resolve exclusions (config file + --exclude flag)
	exclude := make(map[string]bool)
	for _, e := range config.GetExclusions(name) {
		exclude[e] = true
	}
	if flagExclude != "" {
		for _, e := range strings.Split(flagExclude, ",") {
			exclude[strings.TrimSpace(e)] = true
		}
	}

	// Fetch repo list
	fmt.Printf("\nFetching repos from %s: %s\n\n", platform, name)

	var repos []repoInfo

	if platform == "github" {
		raw, err := gh.ListRepos(name, token)
		if err != nil {
			return fmt.Errorf("failed to fetch repos: %w", err)
		}
		for _, r := range raw {
			repos = append(repos, repoInfo{r.Name, r.CloneURL, r.Archived, r.Description})
		}
	} else {
		glURL := flagGitLabURL
		if glURL == "" {
			glURL = config.GetGitLabURL()
		}
		raw, err := gl.ListRepos(name, token, glURL)
		if err != nil {
			return fmt.Errorf("failed to fetch repos: %w", err)
		}
		for _, r := range raw {
			repos = append(repos, repoInfo{r.Name, r.CloneHTTP, r.Archived, r.Description})
		}
	}

	// Apply exclusions and --skip-archived
	var filtered []repoInfo
	for _, r := range repos {
		if exclude[r.Name] || (flagSkipArch && r.Archived) {
			continue
		}
		filtered = append(filtered, r)
	}

	if len(filtered) == 0 {
		fmt.Println("No repos found (or all filtered out).")
		return nil
	}

	// --pick: interactive checkbox selection
	if flagPick {
		selected, err := pickRepos(filtered)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			fmt.Println("Nothing selected.")
			return nil
		}
		filtered = selected
	}

	// --dry-run: just list, don't clone
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

	// Clone or pull each repo
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	fmt.Printf("Destination: %s\n\n", dest)

	cloneURL := func(r repoInfo) string {
		if token != "" && strings.HasPrefix(r.CloneURL, "https://") {
			return strings.Replace(r.CloneURL, "https://", "https://oauth2:"+token+"@", 1)
		}
		return r.CloneURL
	}

	counts := map[git.Status]int{}
	for _, r := range filtered {
		result := git.CloneOrPull(r.Name, cloneURL(r), dest)
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

	return nil
}

// pickRepos shows an interactive checkbox list and returns selected repos.
func pickRepos(repos []repoInfo) ([]repoInfo, error) {
	// Build label list — show archived status in the label
	labels := make([]string, len(repos))
	for i, r := range repos {
		label := r.Name
		if r.Archived {
			label += " (archived)"
		}
		if r.Description != "" {
			desc := r.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			label += " — " + desc
		}
		labels[i] = label
	}

	// Default: all selected
	var chosen []string
	prompt := &survey.MultiSelect{
		Message:  "Select repos to clone:",
		Options:  labels,
		Default:  labels,
		PageSize: 20,
	}

	if err := survey.AskOne(prompt, &chosen); err != nil {
		return nil, err
	}

	// Map chosen labels back to repos
	chosenSet := make(map[string]bool, len(chosen))
	for _, c := range chosen {
		chosenSet[c] = true
	}

	var selected []repoInfo
	for i, label := range labels {
		if chosenSet[label] {
			selected = append(selected, repos[i])
		}
	}
	return selected, nil
}
