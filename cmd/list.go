package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List aliases and workflows",
	Long: `List all aliases and workflows in the configuration.
Example:
  gnb list
  gnb list --tag deploy
  gnb list --scope project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tagFilter, _ := cmd.Flags().GetString("tag")
		scopeFilter, _ := cmd.Flags().GetString("scope")
		if scopeFilter != "" && scopeFilter != "global" && scopeFilter != "project" {
			return fmt.Errorf("invalid scope %q: must be global or project", scopeFilter)
		}

		scoped, err := config.LoadScoped()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		hasProject := scoped.Project != nil && scopeFilter == ""

		// Show conflicts if any
		if scopeFilter == "" && len(scoped.Conflicts) > 0 {
			cmd.Println("=== Conflicts ===")
			for _, c := range scoped.Conflicts {
				cmd.Printf("  %s '%s': global=%s, project=%s (project wins)\n", c.Type, c.Name, c.GlobalVal, c.ProjectVal)
			}
			cmd.Println()
		}

		cfg := configForListScope(scoped, scopeFilter)

		cmd.Println("=== Aliases ===")
		aliasCount := 0
		for name, alias := range cfg.Aliases {
			if tagFilter != "" && !containsTag(alias.Tags, tagFilter) {
				continue
			}
			scope := scopeLabel(name, scoped, "alias", hasProject, scopeFilter)
			cmd.Printf("- %s: %s%s\n", name, alias.Cmd, scope)
			aliasCount++
		}
		if aliasCount == 0 {
			cmd.Println("No aliases found")
		}

		cmd.Println("\n=== Workflows ===")
		wfCount := 0
		for name, workflow := range cfg.Workflows {
			if tagFilter != "" && !containsTag(workflow.Tags, tagFilter) {
				continue
			}
			scope := scopeLabel(name, scoped, "workflow", hasProject, scopeFilter)
			cmd.Printf("- %s: %s%s\n", name, workflow.Description, scope)
			if len(workflow.Tags) > 0 {
				cmd.Printf("  Tags: %v\n", workflow.Tags)
			}
			wfCount++
		}
		if wfCount == 0 {
			cmd.Println("No workflows found")
		}
		return nil
	},
}

func configForListScope(scoped *config.ScopedConfig, scopeFilter string) *config.Config {
	switch scopeFilter {
	case "global":
		return scoped.Global
	case "project":
		if scoped.Project != nil {
			return scoped.Project
		}
		return emptyListConfig()
	default:
		return scoped.Merged
	}
}

func emptyListConfig() *config.Config {
	return &config.Config{
		Aliases:   make(map[string]config.Alias),
		Workflows: make(map[string]config.Workflow),
	}
}

// scopeLabel returns " [global]" or " [project]" when both scopes exist.
func scopeLabel(name string, scoped *config.ScopedConfig, itemType string, hasProject bool, scopeFilter string) string {
	if scopeFilter != "" {
		return " [" + scopeFilter + "]"
	}
	if !hasProject {
		return ""
	}

	switch itemType {
	case "alias":
		if scoped.Project != nil {
			if _, ok := scoped.Project.Aliases[name]; ok {
				return " [project]"
			}
		}
		return " [global]"
	case "workflow":
		if scoped.Project != nil {
			if _, ok := scoped.Project.Workflows[name]; ok {
				return " [project]"
			}
		}
		return " [global]"
	}
	return ""
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func init() {
	listCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	listCmd.Flags().String("scope", "", "Filter by scope (global or project)")
	RootCmd.AddCommand(listCmd)
}
