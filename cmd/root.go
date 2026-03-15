package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:   "orgclone",
	Short: "Clone entire GitHub orgs or GitLab groups with one command",
	Long: `Clone entire GitHub organizations or GitLab groups with one command.
Runs again on existing folders to pull updates — no re-cloning.`,
	Example: `  orgclone clone my-org
  orgclone clone my-group --gitlab
  orgclone clone my-org --pick
  orgclone default platform gitlab
  orgclone default dest ~/projects`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Group commands for cleaner help output
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Commands:"},
		&cobra.Group{ID: "config", Title: "Configuration:"},
	)

	// Hide completion and help subcommands — they clutter the output
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Remove the empty "Additional Commands" section cobra adds by default
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}
