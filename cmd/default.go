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
	Use:     "default <setting> <value>",
	Short:   "Set a default value permanently",
	GroupID: "config",
	Long: `Set a default value that persists across all future runs.
Settings are saved to ~/.orgclone.yml.`,
	Example: `  orgclone default platform gitlab    # use GitLab by default
  orgclone default platform github    # switch back to GitHub
  orgclone default dest ~/projects    # change default clone destination`,
	Args: cobra.ExactArgs(2),
	RunE: runDefault,
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}

func runDefault(cmd *cobra.Command, args []string) error {
	setting := strings.ToLower(args[0])
	value := args[1]

	switch setting {
	case "platform":
		value = strings.ToLower(value)
		if value != "github" && value != "gitlab" {
			return fmt.Errorf("platform must be 'github' or 'gitlab'")
		}
		return saveConfig("default_platform", value, fmt.Sprintf("Default platform set to: %s", value))

	case "dest":
		// Expand ~ so we store a clean path
		if strings.HasPrefix(value, "~/") {
			value = "~/" + value[2:] // keep tilde for portability
		}
		return saveConfig("default_dest", value, fmt.Sprintf("Default destination set to: %s", value))

	default:
		return fmt.Errorf("unknown setting %q — available: platform, dest", setting)
	}
}

func saveConfig(key, value, successMsg string) error {
	cfgPath := filepath.Join(homeDir(), ".orgclone.yml")

	raw := make(map[string]any)
	if data, err := os.ReadFile(cfgPath); err == nil {
		_ = yaml.Unmarshal(data, &raw)
	}

	raw[key] = value

	data, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return err
	}

	fmt.Println(successMsg)
	return nil
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
