package tui

import "sort"

// Category represents a group of items with the same tag.
type Category struct {
	Tag   string
	Items []Item
}

// GroupByTag organizes items into categories by their first tag.
// Items without tags go into an "Untagged" group at the end.
func GroupByTag(items []Item) []Category {
	groups := make(map[string][]Item)
	var untagged []Item

	for _, item := range items {
		if len(item.Tags) > 0 {
			tag := item.Tags[0] // group by first tag
			groups[tag] = append(groups[tag], item)
		} else {
			untagged = append(untagged, item)
		}
	}

	// Sort tags alphabetically
	var tags []string
	for tag := range groups {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	var categories []Category
	for _, tag := range tags {
		categories = append(categories, Category{
			Tag:   tag,
			Items: groups[tag],
		})
	}

	// Untagged at the end
	if len(untagged) > 0 {
		categories = append(categories, Category{
			Tag:   "Untagged",
			Items: untagged,
		})
	}

	return categories
}
