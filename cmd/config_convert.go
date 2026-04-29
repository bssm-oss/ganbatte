package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// configConvertCmd represents the "config convert" command
var configConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert config to another format",
	Long: `Convert the active config file to a different format (TOML, YAML, or JSON).
Example:
  gnb config convert --to yaml
  gnb config convert --to json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetFormat, _ := cmd.Flags().GetString("to")
		if targetFormat == "" {
			return fmt.Errorf("--to flag is required (toml, yaml, or json)")
		}

		_, meta, err := config.LoadWithMeta()
		if err != nil {
			return err
		}

		if meta == nil || meta.FilePath == "" {
			return fmt.Errorf("no config file found. Run 'gnb init' first")
		}

		if meta.Format == targetFormat {
			cmd.Printf("Config is already in %s format\n", targetFormat)
			return nil
		}

		destPath, err := config.Convert(meta.FilePath, targetFormat)
		if err != nil {
			return err
		}

		cmd.Printf("Converted %s → %s\n", meta.FilePath, destPath)
		return nil
	},
}

func init() {
	configConvertCmd.Flags().String("to", "", "Target format (toml, yaml, json)")
	configCmd.AddCommand(configConvertCmd)
}
