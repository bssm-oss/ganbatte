package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ganbatte configuration",
	Long: `Initialize ganbatte configuration by detecting shell, 
choosing format, and creating example workflows.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing ganbatte...")

		// 홈 디렉토리에서 설정 파일 경로 결정
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".config", "ganbatte")

		// 설정 디렉토리 생성
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		// 설정 파일 경로 (기본값: TOML)
		configFile := filepath.Join(configDir, "config.toml")

		// 기본 설정 내용
		configContent := `# ganbatte configuration file
# 지원 포맷: TOML, YAML, JSON
version = "0.1.0"
global_scope = true

# 예시 alias
# [alias.gs]
# cmd = "git status -sb"

# 예시 workflow
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
`

		// 설정 파일 쓰기
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created default config at %s\n", configFile)

		fmt.Println("Initialization complete!")
		fmt.Println("Try 'gnb add gs \"git status -sb\"' to add your first alias")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
