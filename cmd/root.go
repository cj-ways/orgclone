package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "orgclone",
	Short: "Clone entire GitHub orgs or GitLab groups with one command",
	Long: `orgclone fetches all repositories from a GitHub organization or GitLab group
and clones them locally. Running it again pulls updates on existing repos.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
