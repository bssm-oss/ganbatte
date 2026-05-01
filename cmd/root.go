package cmd

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/bssm-oss/ganbatte/internal/tui"
	"github.com/bssm-oss/ganbatte/internal/workflow"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "gnb",
	Short:        "ganbatte - for lazy developers | 頑張って !",
	Long:         `워크플로우/단축어 관리 CLI. lazyasf의 정신적 후속작으로, 단순 alias 관리를 넘어 명령 시퀀스를 워크플로우로 묶고, shell history에서 패턴을 자동 발굴해 추천하는 도구.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		scoped, err := config.LoadScoped()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		cfg := scoped.Merged

		items := tui.ItemsFromScopedConfig(scoped)
		if len(items) == 0 {
			cmd.Println("No aliases or workflows configured.")
			cmd.Println("Run 'gnb init' to get started, then 'gnb add <name> <command>' to add aliases.")
			return nil
		}

		if !isInteractiveTerminal(os.Stdin) {
			return fmt.Errorf("TUI requires an interactive terminal; use 'gnb list' or 'gnb run <name>' in scripts")
		}

		m := tui.New(items)
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}

		result := finalModel.(tui.Model)
		if result.SelectedItem == nil {
			return nil
		}

		item := result.SelectedItem

		switch result.SelectedAction {
		case tui.ActionRun:
			return handleRun(scoped, item)
		case tui.ActionEdit:
			return handleEdit(cfg, item)
		case tui.ActionDelete:
			return handleDelete(cfg, item)
		}

		return nil
	},
}

func isInteractiveTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func handleRun(scoped *config.ScopedConfig, item *tui.Item) error {
	switch item.Type {
	case tui.AliasItem:
		if projectOverrides(scoped, item.Name, "alias") {
			if !confirmPrompt(os.Stdout, fmt.Sprintf("Project alias '%s' overrides a global alias. Continue? [y/N] ", item.Name)) {
				return nil
			}
		}
		if item.Confirm {
			if !confirmPrompt(os.Stdout, fmt.Sprintf("⚠ Run %q? [y/N] ", item.Command)) {
				return nil
			}
		}
		fmt.Fprintf(os.Stdout, "Running: %s\n", item.Command)
		ex := &workflow.RealExecutor{}
		return ex.Execute(item.Command)
	case tui.WorkflowItem:
		if projectOverrides(scoped, item.Name, "workflow") {
			if !confirmPrompt(os.Stdout, fmt.Sprintf("Project workflow '%s' overrides a global workflow. Continue? [y/N] ", item.Name)) {
				return nil
			}
		}
		fmt.Fprintf(os.Stdout, "Running workflow: %s\n", item.Description)
		wf := workflow.Workflow{
			Description: item.Description,
			Params:      item.Params,
		}
		for _, s := range item.Steps {
			wf.Steps = append(wf.Steps, workflow.Step{
				Run:     s.Run,
				OnFail:  s.OnFail,
				Confirm: s.Confirm,
			})
		}
		return workflow.Run(wf, nil, &workflow.RealExecutor{}, workflow.RunOptions{
			Writer: os.Stdout,
		})
	}
	return nil
}

func handleEdit(cfg *config.Config, item *tui.Item) error {
	// Find config file path to open in editor
	_, meta, err := config.LoadWithMeta()
	if err != nil {
		return fmt.Errorf("loading config meta: %w", err)
	}

	if meta == nil || meta.FilePath == "" {
		return fmt.Errorf("no config file found")
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	fmt.Fprintf(os.Stdout, "Opening %s in %s (item: %s)\n", meta.FilePath, editor, item.Name)
	c := exec.Command(editor, meta.FilePath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func handleDelete(cfg *config.Config, item *tui.Item) error {
	switch item.Type {
	case tui.AliasItem:
		delete(cfg.Aliases, item.Name)
	case tui.WorkflowItem:
		delete(cfg.Workflows, item.Name)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Deleted %s '%s'\n", item.TypeLabel(), item.Name)
	return nil
}
