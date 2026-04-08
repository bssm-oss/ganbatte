package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bssm-oss/ganbatte/internal/shell"
	"github.com/spf13/cobra"
)

var defaultConfigs = map[string]string{
	"toml": `# ganbatte configuration file
version = "0.1.0"
global_scope = true

# [alias.gs]
# cmd = "git status -sb"

# [workflow.deploy]
# description = "Lint, test, build, push"
# params = ["branch"]
# steps = [
#   { run = "pnpm lint" },
#   { run = "pnpm test", on_fail = "stop" },
#   { run = "pnpm build" },
#   { run = "git push origin {branch}", confirm = true },
# ]
# tags = ["deploy", "ci"]
`,
	"yaml": `# ganbatte configuration file
version: "0.1.0"
global_scope: true

# alias:
#   gs:
#     cmd: "git status -sb"

# workflow:
#   deploy:
#     description: "Lint, test, build, push"
#     params: [branch]
#     steps:
#       - run: pnpm lint
#       - run: pnpm test
#         on_fail: stop
#       - run: pnpm build
#       - run: "git push origin {branch}"
#         confirm: true
#     tags: [deploy, ci]
`,
	"json": `{
  "version": "0.1.0",
  "global_scope": true,
  "alias": {},
  "workflow": {}
}
`,
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ganbatte configuration",
	Long: `Initialize ganbatte configuration by detecting shell,
choosing format, and creating example config.
Example:
  gnb init
  gnb init --format yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		formatFlag, _ := cmd.Flags().GetString("format")

		// Shell detection
		sh := shell.Detect()
		cmd.Printf("Detected shell: %s\n", sh)

		// Format selection
		format := formatFlag
		if format == "" {
			// Interactive prompt
			cmd.Print("Config format (toml/yaml/json) [toml]: ")
			reader := bufio.NewReader(cmd.InOrStdin())
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "" {
				format = "toml"
			} else {
				format = input
			}
		}

		// Validate format
		switch format {
		case "toml", "yaml", "json":
			// ok
		default:
			return fmt.Errorf("unsupported format '%s' (use toml, yaml, or json)", format)
		}

		// Create config directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}

		configDir := filepath.Join(home, ".config", "ganbatte")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}

		// Write config file
		configFile := filepath.Join(configDir, "config."+format)
		content := defaultConfigs[format]
		if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		cmd.Printf("Created %s config at %s\n", format, configFile)
		cmd.Println("Try 'gnb add gs \"git status -sb\"' to add your first alias")
		return nil
	},
}

func init() {
	initCmd.Flags().String("format", "", "Config format (toml, yaml, json)")
	RootCmd.AddCommand(initCmd)
}
