package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config parent command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long: `Manage ganbatte configuration files.
Subcommands:
  gnb config path      Show active config file path
  gnb config convert   Convert config to another format`,
}

func init() {
	RootCmd.AddCommand(configCmd)
}
