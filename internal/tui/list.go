package tui

import (
	"fmt"
	"strings"

	"github.com/bssm-oss/ganbatte/internal/config"
)

// ItemType represents the type of a list item.
type ItemType int

const (
	AliasItem ItemType = iota
	WorkflowItem
)

// Scope indicates where an item is defined.
type Scope string

const (
	GlobalScope  Scope = "global"
	ProjectScope Scope = "project"
)

// Item represents a single item in the TUI list.
type Item struct {
	Name        string
	Type        ItemType
	Command     string // for aliases
	Description string // for workflows
	Steps       []config.Step
	Params      []string
	Tags        []string
	Confirm     bool  // for aliases with confirm guard
	Scope       Scope // global or project
}

// Title returns the display title.
func (i Item) Title() string {
	if i.Scope == ProjectScope {
		return i.Name + " [project]"
	}
	return i.Name
}

// FilterValue returns the string used for fuzzy matching.
func (i Item) FilterValue() string {
	parts := []string{i.Name}
	if i.Command != "" {
		parts = append(parts, i.Command)
	}
	if i.Description != "" {
		parts = append(parts, i.Description)
	}
	parts = append(parts, i.Tags...)
	return strings.Join(parts, " ")
}

// TypeLabel returns "alias" or "workflow".
func (i Item) TypeLabel() string {
	if i.Type == AliasItem {
		return "alias"
	}
	return "workflow"
}

// Preview returns a detailed string for the preview panel.
func (i Item) Preview() string {
	var b strings.Builder

	if i.Type == AliasItem {
		fmt.Fprintf(&b, "Type: alias\n")
		fmt.Fprintf(&b, "Command: %s\n", i.Command)
		if i.Confirm {
			fmt.Fprintf(&b, "Confirm: yes\n")
		}
	} else {
		fmt.Fprintf(&b, "Type: workflow\n")
		if i.Description != "" {
			fmt.Fprintf(&b, "Description: %s\n", i.Description)
		}
		if len(i.Params) > 0 {
			fmt.Fprintf(&b, "Params: %s\n", strings.Join(i.Params, ", "))
		}
		if len(i.Steps) > 0 {
			fmt.Fprintf(&b, "\nSteps:\n")
			for j, s := range i.Steps {
				fmt.Fprintf(&b, "  %d. %s\n", j+1, s.Run)
				if s.OnFail != "" {
					fmt.Fprintf(&b, "     on_fail: %s\n", s.OnFail)
				}
				if s.Confirm {
					fmt.Fprintf(&b, "     confirm: true\n")
				}
			}
		}
	}

	if len(i.Tags) > 0 {
		fmt.Fprintf(&b, "\nTags: %s\n", strings.Join(i.Tags, ", "))
	}

	return b.String()
}

// ItemsFromConfig converts a config into a list of TUI items.
func ItemsFromConfig(cfg *config.Config) []Item {
	return itemsFromConfigWithScope(cfg, "")
}

// ItemsFromScopedConfig converts a scoped config into items with scope labels.
// Project items appear first.
func ItemsFromScopedConfig(scoped *config.ScopedConfig) []Item {
	var items []Item

	// Project items first for discoverability
	if scoped.Project != nil {
		items = append(items, itemsFromConfigWithScope(scoped.Project, ProjectScope)...)
	}

	if scoped.Global != nil {
		// Skip items that are overridden by project scope
		globalItems := itemsFromConfigWithScope(scoped.Global, GlobalScope)
		for _, gi := range globalItems {
			overridden := false
			if scoped.Project != nil {
				if gi.Type == AliasItem {
					_, overridden = scoped.Project.Aliases[gi.Name]
				} else {
					_, overridden = scoped.Project.Workflows[gi.Name]
				}
			}
			if !overridden {
				items = append(items, gi)
			}
		}
	}

	return items
}

func itemsFromConfigWithScope(cfg *config.Config, scope Scope) []Item {
	var items []Item

	for name, alias := range cfg.Aliases {
		items = append(items, Item{
			Name:    name,
			Type:    AliasItem,
			Command: alias.Cmd,
			Confirm: alias.Confirm,
			Scope:   scope,
		})
	}

	for name, wf := range cfg.Workflows {
		items = append(items, Item{
			Name:        name,
			Type:        WorkflowItem,
			Description: wf.Description,
			Steps:       wf.Steps,
			Params:      wf.Params,
			Tags:        wf.Tags,
			Scope:       scope,
		})
	}

	return items
}
