package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <name> <command>",
	Short: "Edit an existing alias",
	Long: `Edit an existing alias in the configuration.
Example:
  gnb edit gs "git status --short --branch"
  gnb edit --global gs "git status"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := args[1]
		global, _ := cmd.Flags().GetBool("global")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if _, exists := cfg.Aliases[name]; !exists {
			return fmt.Errorf("alias '%s' not found. Use 'gnb add %s <command>' to create it", name, name)
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
		cmd.Printf("Updated alias '%s' = %s%s\n", name, command, scope)
		return nil
	},
}

func init() {
	editCmd.Flags().Bool("global", false, "Save to global config")
	RootCmd.AddCommand(editCmd)
}
