package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justn-hyeok/ganbatte/internal/shell"
	"github.com/spf13/cobra"
)

var projectConfigs = map[string]string{
	"toml": `# ganbatte project configuration
version = "1.0.0"

# [workflow.setup]
# description = "Project initial setup"
# steps = [
#   { run = "npm install" },
#   { run = "cp .env.example .env" },
# ]
# tags = ["onboarding"]

# [workflow.dev]
# description = "Start dev server"
# steps = [
#   { run = "docker compose up -d" },
#   { run = "npm run dev" },
# ]
# tags = ["dev"]
`,
	"yaml": `# ganbatte project configuration
version: "1.0.0"

# workflow:
#   setup:
#     description: "Project initial setup"
#     steps:
#       - run: npm install
#       - run: cp .env.example .env
#     tags: [onboarding]
#
#   dev:
#     description: "Start dev server"
#     steps:
#       - run: docker compose up -d
#       - run: npm run dev
#     tags: [dev]
`,
	"json": `{
  "version": "1.0.0",
  "alias": {},
  "workflow": {}
}
`,
}

var defaultConfigs = map[string]string{
	"toml": `# ganbatte configuration file
version = "1.0.0"
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
version: "1.0.0"
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
  "version": "1.0.0",
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
  gnb init --format yaml
  gnb init --project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		formatFlag, _ := cmd.Flags().GetString("format")
		projectFlag, _ := cmd.Flags().GetBool("project")

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

		if projectFlag {
			return initProject(cmd, format)
		}

		return initGlobal(cmd, format)
	},
}

func initGlobal(cmd *cobra.Command, format string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "ganbatte")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config."+format)
	content := defaultConfigs[format]
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	cmd.Printf("Created %s config at %s\n", format, configFile)
	cmd.Println("Try 'gnb add gs \"git status -sb\"' to add your first alias")
	return nil
}

func initProject(cmd *cobra.Command, format string) error {
	configFile := ".ganbatte." + format
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("project config already exists: %s", configFile)
	}

	content := projectConfigs[format]
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing project config: %w", err)
	}

	cmd.Printf("Created project config at %s\n", configFile)
	cmd.Println("Commit this file to share workflows with your team")
	return nil
}

func init() {
	initCmd.Flags().String("format", "", "Config format (toml, yaml, json)")
	initCmd.Flags().Bool("project", false, "Create project-scoped config (.ganbatte.*) in current directory")
	RootCmd.AddCommand(initCmd)
}
