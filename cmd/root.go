package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gnb",
	Short: "ganbatte - for lazy developers | 頑張って !",
	Long:  `워크플로우/단축어 관리 CLI. lazyasf의 정신적 후속작으로, 단순 alias 관리를 넘어 명령 시퀀스를 워크플로우로 묶고, shell history에서 패턴을 자동 발굴해 추천하는 도구.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TUI 브라우저 실행 (향후 구현)
		fmt.Println("TUI 브라우저 실행 준비 중... (v0.2 기능)")
	},
}
