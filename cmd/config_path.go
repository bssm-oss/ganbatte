package cmd

import (
	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// configPathCmd represents the "config path" command
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show active config file path",
	Long: `Print the path of the currently active configuration file.
Useful for syncingsh integration.
Example:
  gnb config path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, meta, err := config.LoadWithMeta()
		if err != nil {
			return err
		}

		if meta == nil || meta.FilePath == "" {
			cmd.Println("No config file found. Run 'gnb init' to create one.")
			return nil
		}

		cmd.Println(meta.FilePath)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
}
