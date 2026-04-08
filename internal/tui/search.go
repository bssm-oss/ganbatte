package tui

import "github.com/sahilm/fuzzy"

// matchable wraps items for fuzzy matching.
type matchable []Item

func (m matchable) String(i int) string {
	return m[i].FilterValue()
}

func (m matchable) Len() int {
	return len(m)
}

// FuzzyFilter returns items matching the query, sorted by match score.
// If query is empty, returns all items.
func FuzzyFilter(items []Item, query string) []Item {
	if query == "" {
		return items
	}

	matches := fuzzy.FindFrom(query, matchable(items))
	result := make([]Item, len(matches))
	for i, m := range matches {
		result[i] = items[m.Index]
	}
	return result
}
