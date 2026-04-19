package cmd

import (
	"fmt"
	"os"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [names...]",
	Short: "Export aliases and workflows to a file",
	Long: `Export selected aliases and workflows to a config file.
If no names are specified, exports everything.
Example:
  gnb export --output backup.toml
  gnb export gs deploy --format yaml --output selected.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		aliasesOnly, _ := cmd.Flags().GetBool("aliases-only")

		if output == "" {
			return fmt.Errorf("--output flag is required")
		}

		if format == "" {
			format = "toml"
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		data, err := config.Export(cfg, config.ExportOptions{
			Names:       args,
			Format:      format,
			AliasesOnly: aliasesOnly,
		})
		if err != nil {
			return err
		}

		if err := os.WriteFile(output, data, 0o644); err != nil {
			return fmt.Errorf("writing export file: %w", err)
		}

		count := len(cfg.Aliases) + len(cfg.Workflows)
		if aliasesOnly {
			count = len(cfg.Aliases)
		}
		if len(args) > 0 {
			count = len(args)
		}
		cmd.Printf("Exported %d item(s) to %s\n", count, output)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringP("output", "o", "", "Output file path (required)")
	exportCmd.Flags().StringP("format", "f", "toml", "Output format (toml, yaml, json)")
	exportCmd.Flags().Bool("aliases-only", false, "Export only aliases, skip workflows")
	RootCmd.AddCommand(exportCmd)
}
