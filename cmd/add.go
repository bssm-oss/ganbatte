package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a new alias",
	Long: `Add a new alias to the configuration.
Example:
  gnb add gs "git status -sb"
  gnb add ll "ls -la"
  gnb add --global gs "git status -sb"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := args[1]
		global, _ := cmd.Flags().GetBool("global")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.Aliases == nil {
			cfg.Aliases = make(map[string]config.Alias)
		}

		if _, exists := cfg.Aliases[name]; exists {
			return fmt.Errorf("alias '%s' already exists. Use 'gnb edit %s <command>' to modify it", name, name)
		}

		cfg.Aliases[name] = config.Alias{
			Cmd: command,
		}

		if global {
			if err := cfg.SaveGlobal(); err != nil {
				return fmt.Errorf("saving global config: %w", err)
			}
		} else {
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
		}

		scope := ""
		if global {
			scope = " (global)"
		}
		cmd.Printf("Added alias '%s' = %s%s\n", name, command, scope)
		return nil
	},
}

func init() {
	addCmd.Flags().Bool("global", false, "Save to global config")
	RootCmd.AddCommand(addCmd)
}
