package tui

import (
	"fmt"
	"strings"

	"github.com/bssm-oss/ganbatte/internal/config"
)

// ItemType represents the type of a list item.
type ItemType int

const (
	AliasItem    ItemType = iota
	WorkflowItem
)

// Item represents a single item in the TUI list.
type Item struct {
	Name        string
	Type        ItemType
	Command     string   // for aliases
	Description string   // for workflows
	Steps       []config.Step
	Params      []string
	Tags        []string
}

// Title returns the display title.
func (i Item) Title() string {
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
	var items []Item

	for name, alias := range cfg.Aliases {
		items = append(items, Item{
			Name:    name,
			Type:    AliasItem,
			Command: alias.Cmd,
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
		})
	}

	return items
}
