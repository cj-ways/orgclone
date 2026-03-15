package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var defaultCmd = &cobra.Command{
	Use:     "default <platform>",
	Short:   "Set the default platform (github or gitlab)",
	GroupID: "config",
	Example: `  orgclone default gitlab   # use GitLab by default
  orgclone default github   # switch back to GitHub`,
	Args: cobra.ExactArgs(1),
	RunE: runDefault,
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}

func runDefault(cmd *cobra.Command, args []string) error {
	platform := strings.ToLower(args[0])
	if platform != "github" && platform != "gitlab" {
		return fmt.Errorf("platform must be 'github' or 'gitlab'")
	}

	cfgPath := filepath.Join(homeDir(), ".orgclone.yml")

	// Load existing config as raw map to preserve unknown fields
	raw := make(map[string]any)
	if data, err := os.ReadFile(cfgPath); err == nil {
		_ = yaml.Unmarshal(data, &raw)
	}

	raw["default_platform"] = platform

	data, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return err
	}

	fmt.Printf("Default platform set to: %s\n", platform)
	return nil
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
