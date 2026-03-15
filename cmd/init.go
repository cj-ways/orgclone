package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Create a sample ~/.orgclone.yml config file",
	GroupID: "config",
	Example: `  orgclone init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		cfgPath := filepath.Join(home, ".orgclone.yml")

		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Printf("Config already exists: %s\n", cfgPath)
			return nil
		}

		content := `# orgclone configuration (~/.orgclone.yml)

default_dest: ~/Desktop

github:
  token: ""   # or set GITHUB_TOKEN env var

gitlab:
  token: ""   # or set GITLAB_TOKEN env var
  url: https://gitlab.com  # change for self-hosted instances

orgs:
  my-org:
    exclude:
      - old-repo
      - legacy-stuff
`
		if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
			return err
		}
		fmt.Printf("Created config: %s\n", cfgPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
