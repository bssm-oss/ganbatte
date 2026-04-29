package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import aliases and workflows from a file",
	Long: `Import aliases and workflows from a config file.
By default, existing items are skipped (merge mode).
Use --replace to overwrite existing items.
Example:
  gnb import backup.toml
  gnb import shared.yaml --replace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		replace, _ := cmd.Flags().GetBool("replace")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		strategy := "merge"
		if replace {
			strategy = "replace"
		}

		result, err := config.Import(cfg, srcPath, strategy)
		if err != nil {
			return err
		}

		if len(result.Added) > 0 {
			cmd.Printf("Added: %v\n", result.Added)
		}
		if len(result.Skipped) > 0 {
			cmd.Printf("Skipped (already exists): %v\n", result.Skipped)
		}
		if len(result.Replaced) > 0 {
			cmd.Printf("Replaced: %v\n", result.Replaced)
		}

		if len(result.Added) > 0 || len(result.Replaced) > 0 {
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			cmd.Println("Config saved")
		} else {
			cmd.Println("No changes made")
		}

		return nil
	},
}

func init() {
	importCmd.Flags().Bool("replace", false, "Overwrite existing items")
	RootCmd.AddCommand(importCmd)
}
