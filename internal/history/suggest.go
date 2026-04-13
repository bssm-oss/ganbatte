package history

import (
	"fmt"
	"sort"
	"strings"
)

// Suggestion represents a recommended alias or workflow.
type Suggestion struct {
	Type    string   // "alias" or "workflow"
	Name    string   // suggested name
	Command string   // for alias suggestions
	Steps   []string // for workflow suggestions
	Reason  string   // human-readable explanation
}

// SuggestOptions configures the suggestion engine.
type SuggestOptions struct {
	// MinFrequency is the minimum number of occurrences to suggest an alias.
	MinFrequency int
	// MinSequence is the minimum number of occurrences of a command sequence
	// to suggest a workflow.
	MinSequence int
	// MaxSuggestions limits the number of suggestions returned.
	MaxSuggestions int
}

// DefaultSuggestOptions returns sensible defaults.
func DefaultSuggestOptions() SuggestOptions {
	return SuggestOptions{
		MinFrequency:   5,
		MinSequence:    3,
		MaxSuggestions: 10,
	}
}

// Suggest analyzes history entries and returns alias/workflow suggestions.
func Suggest(entries []Entry, existingAliases map[string]string, opts SuggestOptions) []Suggestion {
	if opts.MinFrequency <= 0 {
		opts.MinFrequency = 5
	}
	if opts.MinSequence <= 0 {
		opts.MinSequence = 3
	}
	if opts.MaxSuggestions <= 0 {
		opts.MaxSuggestions = 10
	}

	var suggestions []Suggestion

	// 1. Frequency-based alias suggestions
	suggestions = append(suggestions, suggestAliases(entries, existingAliases, opts)...)

	// 2. Sequence-based workflow suggestions
	suggestions = append(suggestions, suggestWorkflows(entries, opts)...)

	// Limit results
	if len(suggestions) > opts.MaxSuggestions {
		suggestions = suggestions[:opts.MaxSuggestions]
	}

	return suggestions
}

// suggestAliases finds frequently used commands and suggests aliases for them.
func suggestAliases(entries []Entry, existing map[string]string, opts SuggestOptions) []Suggestion {
	// Count command frequencies
	freq := make(map[string]int)
	for _, e := range entries {
		cmd := strings.TrimSpace(e.Command)
		if cmd != "" {
			freq[cmd]++
		}
	}

	// Build reverse lookup of existing aliases
	existingCmds := make(map[string]bool)
	for _, cmd := range existing {
		existingCmds[cmd] = true
	}

	// Sort by frequency (descending)
	type cmdCount struct {
		cmd   string
		count int
	}
	var sorted []cmdCount
	for cmd, count := range freq {
		if count >= opts.MinFrequency && !existingCmds[cmd] {
			sorted = append(sorted, cmdCount{cmd, count})
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var suggestions []Suggestion
	for _, cc := range sorted {
		name := generateAliasName(cc.cmd)
		suggestions = append(suggestions, Suggestion{
			Type:    "alias",
			Name:    name,
			Command: cc.cmd,
			Reason:  fmt.Sprintf("Used %d times", cc.count),
		})
	}

	return suggestions
}

// suggestWorkflows finds repeated command sequences and suggests workflows.
func suggestWorkflows(entries []Entry, opts SuggestOptions) []Suggestion {
	if len(entries) < 2 {
		return nil
	}

	// Count 2-command and 3-command sequences
	pairFreq := make(map[string]int)
	tripleFreq := make(map[string]int)

	for i := 0; i < len(entries)-1; i++ {
		a := strings.TrimSpace(entries[i].Command)
		b := strings.TrimSpace(entries[i+1].Command)
		if a == "" || b == "" || a == b {
			continue
		}
		key := a + " && " + b
		pairFreq[key]++

		if i < len(entries)-2 {
			c := strings.TrimSpace(entries[i+2].Command)
			if c != "" && c != b {
				triKey := a + " && " + b + " && " + c
				tripleFreq[triKey]++
			}
		}
	}

	var suggestions []Suggestion

	// Prefer triples over pairs
	type seqCount struct {
		seq   string
		count int
	}
	var triples []seqCount
	for seq, count := range tripleFreq {
		if count >= opts.MinSequence {
			triples = append(triples, seqCount{seq, count})
		}
	}
	sort.Slice(triples, func(i, j int) bool {
		return triples[i].count > triples[j].count
	})

	for _, tc := range triples {
		steps := strings.Split(tc.seq, " && ")
		suggestions = append(suggestions, Suggestion{
			Type:   "workflow",
			Name:   "wf-" + generateAliasName(steps[0]),
			Steps:  steps,
			Reason: fmt.Sprintf("Sequence appeared %d times", tc.count),
		})
	}

	// Add pairs if we don't have enough suggestions
	var pairs []seqCount
	for seq, count := range pairFreq {
		if count >= opts.MinSequence {
			pairs = append(pairs, seqCount{seq, count})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	for _, pc := range pairs {
		steps := strings.Split(pc.seq, " && ")
		suggestions = append(suggestions, Suggestion{
			Type:   "workflow",
			Name:   "wf-" + generateAliasName(steps[0]),
			Steps:  steps,
			Reason: fmt.Sprintf("Pair appeared %d times", pc.count),
		})
	}

	return suggestions
}

// generateAliasName creates a short alias name from a command.
func generateAliasName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "cmd"
	}

	// Use initials of first 2-3 words
	var name strings.Builder
	limit := 3
	if len(parts) < limit {
		limit = len(parts)
	}
	for i := 0; i < limit; i++ {
		word := strings.TrimLeft(parts[i], "-./")
		if word != "" {
			name.WriteByte(word[0])
		}
	}

	result := strings.ToLower(name.String())
	if result == "" {
		return "cmd"
	}
	return result
}
