package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Action represents what the user wants to do after exiting the TUI.
type Action int

const (
	ActionNone   Action = iota
	ActionRun           // Enter: run selected item
	ActionEdit          // e: edit selected item
	ActionDelete        // d: delete selected item
)

type viewMode int

const (
	normalMode viewMode = iota
	searchMode
	helpMode
	confirmDeleteMode
)

// Model is the main bubbletea model for the TUI browser.
type Model struct {
	allItems      []Item
	filteredItems []Item
	cursor        int
	mode          viewMode
	searchInput   textinput.Model
	tagFilter     string
	width         int
	height        int
	keys          KeyMap

	// Set when user selects an item
	SelectedItem *Item
	SelectedAction Action
	Quitting       bool
}

// New creates a new TUI model with the given items.
func New(items []Item) Model {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type // aliases first
		}
		return items[i].Name < items[j].Name
	})

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100

	return Model{
		allItems:      items,
		filteredItems: items,
		keys:          DefaultKeyMap(),
		searchInput:   ti,
		width:         80,
		height:        24,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.mode == searchMode {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Search mode: handle input
	if m.mode == searchMode {
		switch msg.String() {
		case "esc":
			m.mode = normalMode
			m.searchInput.SetValue("")
			m.searchInput.Blur()
			m.applyFilter()
			return m, nil
		case "enter":
			m.mode = normalMode
			m.searchInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}
	}

	// Help mode
	if m.mode == helpMode {
		m.mode = normalMode
		return m, nil
	}

	// Confirm delete mode
	if m.mode == confirmDeleteMode {
		switch msg.String() {
		case "y", "Y":
			if len(m.filteredItems) > 0 {
				item := m.filteredItems[m.cursor]
				m.SelectedItem = &item
				m.SelectedAction = ActionDelete
				return m, tea.Quit
			}
		}
		m.mode = normalMode
		return m, nil
	}

	// Normal mode
	switch {
	case msg.String() == "q" || msg.String() == "ctrl+c":
		m.Quitting = true
		return m, tea.Quit
	case msg.String() == "up" || msg.String() == "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case msg.String() == "down" || msg.String() == "j":
		if m.cursor < len(m.filteredItems)-1 {
			m.cursor++
		}
	case msg.String() == "enter":
		if len(m.filteredItems) > 0 {
			item := m.filteredItems[m.cursor]
			m.SelectedItem = &item
			m.SelectedAction = ActionRun
			return m, tea.Quit
		}
	case msg.String() == "/":
		m.mode = searchMode
		return m, m.searchInput.Focus()
	case msg.String() == "e":
		if len(m.filteredItems) > 0 {
			item := m.filteredItems[m.cursor]
			m.SelectedItem = &item
			m.SelectedAction = ActionEdit
			return m, tea.Quit
		}
	case msg.String() == "d":
		if len(m.filteredItems) > 0 {
			m.mode = confirmDeleteMode
		}
	case msg.String() == "t":
		m.cycleTagFilter()
	case msg.String() == "?":
		m.mode = helpMode
	}

	return m, nil
}

func (m *Model) applyFilter() {
	items := m.allItems

	// Apply tag filter
	if m.tagFilter != "" {
		var filtered []Item
		for _, item := range items {
			for _, tag := range item.Tags {
				if tag == m.tagFilter {
					filtered = append(filtered, item)
					break
				}
			}
		}
		items = filtered
	}

	// Apply search filter
	query := m.searchInput.Value()
	m.filteredItems = FuzzyFilter(items, query)

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filteredItems) {
		m.cursor = max(0, len(m.filteredItems)-1)
	}
}

func (m *Model) cycleTagFilter() {
	// Collect all unique tags
	tagSet := make(map[string]bool)
	for _, item := range m.allItems {
		for _, tag := range item.Tags {
			tagSet[tag] = true
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	if len(tags) == 0 {
		return
	}

	// Cycle through tags, then back to "" (no filter)
	if m.tagFilter == "" {
		m.tagFilter = tags[0]
	} else {
		found := false
		for i, tag := range tags {
			if tag == m.tagFilter && i < len(tags)-1 {
				m.tagFilter = tags[i+1]
				found = true
				break
			}
		}
		if !found {
			m.tagFilter = ""
		}
	}
	m.applyFilter()
}

// View implements tea.Model.
func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	if m.mode == helpMode {
		return m.helpView()
	}

	// Layout: list on left, preview on right
	listWidth := m.width / 2
	previewWidth := m.width - listWidth - 3

	list := m.listView(listWidth)
	preview := m.previewView(previewWidth)

	content := lipgloss.JoinHorizontal(lipgloss.Top, list, " ", preview)

	// Status bar
	var status string
	if m.mode == confirmDeleteMode && len(m.filteredItems) > 0 {
		item := m.filteredItems[m.cursor]
		status = warnStyle.Render(fmt.Sprintf("Delete '%s'? (y/N)", item.Name))
	} else if m.mode == searchMode {
		status = searchStyle.Render("/ ") + m.searchInput.View()
	} else if m.tagFilter != "" {
		status = tagStyle.Render(fmt.Sprintf("Tag: %s", m.tagFilter))
	} else {
		status = helpStyle.Render("↑↓ navigate • enter run • e edit • d delete • / search • t tag • ? help • q quit")
	}

	return titleStyle.Render("ganbatte 頑張って") + "\n" + content + "\n" + statusStyle.Render(status)
}

func (m Model) listView(width int) string {
	var b strings.Builder
	visibleHeight := m.height - 5 // account for title, status, padding

	if len(m.filteredItems) == 0 {
		return dimStyle.Render("No items found")
	}

	// Calculate scroll window
	start := 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	end := start + visibleHeight
	if end > len(m.filteredItems) {
		end = len(m.filteredItems)
	}

	for i := start; i < end; i++ {
		item := m.filteredItems[i]
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = "> "
			style = selectedStyle
		}

		typeTag := dimStyle.Render(fmt.Sprintf("[%s]", item.TypeLabel()))
		line := fmt.Sprintf("%s%s %s", cursor, style.Render(item.Name), typeTag)

		if len(line) > width {
			line = line[:width-1] + "…"
		}
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m Model) previewView(width int) string {
	if len(m.filteredItems) == 0 || m.cursor >= len(m.filteredItems) {
		return previewStyle.Width(width).Render(dimStyle.Render("No selection"))
	}

	item := m.filteredItems[m.cursor]
	title := previewTitleStyle.Render(item.Name)
	body := item.Preview()

	return previewStyle.Width(width).Render(title + "\n" + body)
}

func (m Model) helpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Key Bindings") + "\n\n")

	keys := []struct{ key, desc string }{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"Enter", "Run selected item"},
		{"/", "Search"},
		{"t", "Cycle tag filter"},
		{"e", "Edit in $EDITOR"},
		{"d", "Delete (confirm)"},
		{"?", "Toggle help"},
		{"q/Ctrl+C", "Quit"},
		{"Esc", "Back / clear search"},
	}

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			selectedStyle.Render(fmt.Sprintf("%-12s", k.key)),
			k.desc))
	}

	b.WriteString("\n" + helpStyle.Render("Press any key to return"))
	return b.String()
}
