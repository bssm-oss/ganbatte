package tui

import (
	"fmt"
	"testing"

	"github.com/justn-hyeok/ganbatte/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestItemsFromConfig(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"gs": {Cmd: "git status -sb"},
			"ll": {Cmd: "ls -la"},
		},
		Workflows: map[string]config.Workflow{
			"deploy": {
				Description: "Deploy app",
				Params:      []string{"branch"},
				Steps:       []config.Step{{Run: "pnpm build"}},
				Tags:        []string{"ci"},
			},
		},
	}

	items := ItemsFromConfig(cfg)
	assert.Len(t, items, 3)

	aliases := 0
	workflows := 0
	for _, item := range items {
		switch item.Type {
		case AliasItem:
			aliases++
		case WorkflowItem:
			workflows++
		}
	}
	assert.Equal(t, 2, aliases)
	assert.Equal(t, 1, workflows)
}

func TestItemsFromConfig_Empty(t *testing.T) {
	cfg := &config.Config{
		Aliases:   map[string]config.Alias{},
		Workflows: map[string]config.Workflow{},
	}
	items := ItemsFromConfig(cfg)
	assert.Empty(t, items)
}

func TestItemPreview_Alias(t *testing.T) {
	alias := Item{
		Name:    "gs",
		Type:    AliasItem,
		Command: "git status -sb",
	}
	preview := alias.Preview()
	assert.Contains(t, preview, "alias")
	assert.Contains(t, preview, "git status -sb")
}

func TestItemPreview_Workflow(t *testing.T) {
	wf := Item{
		Name:        "deploy",
		Type:        WorkflowItem,
		Description: "Deploy app",
		Params:      []string{"env"},
		Steps:       []config.Step{{Run: "pnpm build", OnFail: "stop"}, {Run: "pnpm deploy", Confirm: true}},
		Tags:        []string{"ci"},
	}
	preview := wf.Preview()
	assert.Contains(t, preview, "workflow")
	assert.Contains(t, preview, "Deploy app")
	assert.Contains(t, preview, "pnpm build")
	assert.Contains(t, preview, "on_fail: stop")
	assert.Contains(t, preview, "confirm: true")
	assert.Contains(t, preview, "env")
	assert.Contains(t, preview, "ci")
}

func TestItemPreview_WorkflowMinimal(t *testing.T) {
	wf := Item{
		Name: "empty",
		Type: WorkflowItem,
	}
	preview := wf.Preview()
	assert.Contains(t, preview, "workflow")
	assert.NotContains(t, preview, "Description")
}

func TestItemTitle(t *testing.T) {
	item := Item{Name: "gs"}
	assert.Equal(t, "gs", item.Title())
}

func TestItemTypeLabel(t *testing.T) {
	assert.Equal(t, "alias", Item{Type: AliasItem}.TypeLabel())
	assert.Equal(t, "workflow", Item{Type: WorkflowItem}.TypeLabel())
}

func TestFuzzyFilter(t *testing.T) {
	items := []Item{
		{Name: "gs", Type: AliasItem, Command: "git status"},
		{Name: "deploy", Type: WorkflowItem, Description: "Deploy app"},
		{Name: "ll", Type: AliasItem, Command: "ls -la"},
	}

	result := FuzzyFilter(items, "")
	assert.Len(t, result, 3)

	result = FuzzyFilter(items, "git")
	assert.Len(t, result, 1)
	assert.Equal(t, "gs", result[0].Name)

	result = FuzzyFilter(items, "deploy")
	assert.Len(t, result, 1)

	result = FuzzyFilter(items, "zzzzz")
	assert.Empty(t, result)
}

func TestNewModel(t *testing.T) {
	items := []Item{
		{Name: "b", Type: WorkflowItem},
		{Name: "a", Type: AliasItem},
		{Name: "c", Type: AliasItem},
	}
	m := New(items)
	assert.Len(t, m.filteredItems, 3)
	assert.Equal(t, AliasItem, m.filteredItems[0].Type)
	assert.Equal(t, "a", m.filteredItems[0].Name)
	assert.Equal(t, "c", m.filteredItems[1].Name)
}

func TestNewModel_Empty(t *testing.T) {
	m := New(nil)
	assert.Empty(t, m.filteredItems)
	assert.Equal(t, 0, m.cursor)
}

func TestModelInit(t *testing.T) {
	m := New(nil)
	cmd := m.Init()
	assert.Nil(t, cmd)
}

func TestModelView_Empty(t *testing.T) {
	m := New(nil)
	view := m.View()
	assert.Contains(t, view, "ganbatte")
	assert.Contains(t, view, "No items found")
}

func TestModelView_WithItems(t *testing.T) {
	items := []Item{
		{Name: "gs", Type: AliasItem, Command: "git status"},
	}
	m := New(items)
	view := m.View()
	assert.Contains(t, view, "gs")
	assert.Contains(t, view, "ganbatte")
}

func TestModelView_Quitting(t *testing.T) {
	m := New(nil)
	m.Quitting = true
	assert.Equal(t, "", m.View())
}

func TestModelUpdate_WindowSize(t *testing.T) {
	m := New(nil)
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := newM.(Model)
	assert.Equal(t, 120, result.width)
	assert.Equal(t, 40, result.height)
}

func TestModelHelpView(t *testing.T) {
	m := New(nil)
	m.mode = helpMode
	view := m.View()
	assert.Contains(t, view, "Key Bindings")
	assert.Contains(t, view, "Move up")
	assert.Contains(t, view, "Quit")
}

func TestModelListView_Scroll(t *testing.T) {
	var items []Item
	for i := 0; i < 50; i++ {
		items = append(items, Item{Name: fmt.Sprintf("item%d", i), Type: AliasItem})
	}
	m := New(items)
	m.height = 10
	m.cursor = 40
	view := m.listView(40)
	assert.NotEmpty(t, view)
}

func TestModelPreviewView_Empty(t *testing.T) {
	m := New(nil)
	view := m.previewView(40)
	assert.Contains(t, view, "No selection")
}

func TestModelPreviewView_WithItem(t *testing.T) {
	items := []Item{
		{Name: "gs", Type: AliasItem, Command: "git status -sb"},
	}
	m := New(items)
	view := m.previewView(40)
	assert.Contains(t, view, "gs")
}

func TestModelApplyFilter_Tag(t *testing.T) {
	items := []Item{
		{Name: "deploy", Type: WorkflowItem, Tags: []string{"ci"}},
		{Name: "gs", Type: AliasItem},
	}
	m := New(items)
	m.tagFilter = "ci"
	m.applyFilter()
	assert.Len(t, m.filteredItems, 1)
	assert.Equal(t, "deploy", m.filteredItems[0].Name)
}

func TestModelApplyFilter_Search(t *testing.T) {
	items := []Item{
		{Name: "deploy", Type: WorkflowItem, Description: "Deploy app"},
		{Name: "gs", Type: AliasItem, Command: "git status"},
	}
	m := New(items)
	m.searchInput.SetValue("git")
	m.applyFilter()
	assert.Len(t, m.filteredItems, 1)
	assert.Equal(t, "gs", m.filteredItems[0].Name)
}

func TestModelApplyFilter_CursorReset(t *testing.T) {
	items := []Item{
		{Name: "a", Type: AliasItem},
		{Name: "b", Type: AliasItem},
		{Name: "c", Type: AliasItem},
	}
	m := New(items)
	m.cursor = 2
	m.searchInput.SetValue("a")
	m.applyFilter()
	assert.Equal(t, 0, m.cursor)
}

func TestModelCycleTagFilter(t *testing.T) {
	items := []Item{
		{Name: "a", Tags: []string{"ci"}},
		{Name: "b", Tags: []string{"dev"}},
		{Name: "c"},
	}
	m := New(items)

	assert.Equal(t, "", m.tagFilter)
	m.cycleTagFilter()
	assert.Equal(t, "ci", m.tagFilter)
	m.cycleTagFilter()
	assert.Equal(t, "dev", m.tagFilter)
	m.cycleTagFilter()
	assert.Equal(t, "", m.tagFilter) // cycles back
}

func TestModelCycleTagFilter_NoTags(t *testing.T) {
	items := []Item{{Name: "a"}, {Name: "b"}}
	m := New(items)
	m.cycleTagFilter()
	assert.Equal(t, "", m.tagFilter) // no change
}

func TestGroupByTag(t *testing.T) {
	items := []Item{
		{Name: "deploy", Tags: []string{"ci", "deploy"}},
		{Name: "test", Tags: []string{"ci"}},
		{Name: "gs", Tags: nil},
		{Name: "lint", Tags: []string{"dev"}},
		{Name: "ll", Tags: nil},
	}

	categories := GroupByTag(items)
	assert.Len(t, categories, 3)
	assert.Equal(t, "ci", categories[0].Tag)
	assert.Len(t, categories[0].Items, 2)
	assert.Equal(t, "dev", categories[1].Tag)
	assert.Len(t, categories[1].Items, 1)
	assert.Equal(t, "Untagged", categories[2].Tag)
	assert.Len(t, categories[2].Items, 2)
}

func TestGroupByTag_AllTagged(t *testing.T) {
	items := []Item{
		{Name: "a", Tags: []string{"x"}},
		{Name: "b", Tags: []string{"y"}},
	}
	categories := GroupByTag(items)
	assert.Len(t, categories, 2)
	for _, c := range categories {
		assert.NotEqual(t, "Untagged", c.Tag)
	}
}

func TestGroupByTag_Empty(t *testing.T) {
	categories := GroupByTag(nil)
	assert.Empty(t, categories)
}

func TestItemFilterValue(t *testing.T) {
	item := Item{
		Name:        "deploy",
		Description: "Deploy to prod",
		Tags:        []string{"ci", "deploy"},
	}
	fv := item.FilterValue()
	assert.Contains(t, fv, "deploy")
	assert.Contains(t, fv, "Deploy to prod")
	assert.Contains(t, fv, "ci")
}

func TestModelConfirmDeleteView(t *testing.T) {
	items := []Item{{Name: "gs", Type: AliasItem}}
	m := New(items)
	m.mode = confirmDeleteMode
	view := m.View()
	assert.Contains(t, view, "Delete 'gs'?")
}

// suppress unused import
var _ = fmt.Sprintf
