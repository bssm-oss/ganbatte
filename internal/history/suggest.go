package history

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Suggestion represents a recommended alias or workflow.
type Suggestion struct {
	Type            string   // "alias", "param-alias", or "workflow"
	Name            string   // suggested name
	Command         string   // for alias suggestions (may contain {param} placeholders)
	Params          []string // for param-alias suggestions
	Steps           []string // for workflow suggestions
	Reason          string   // human-readable explanation
	SavedKeystrokes int      // estimated keystrokes saved based on history
	Confirm         bool     // true for destructive commands
}

// SuggestOptions configures the suggestion engine.
type SuggestOptions struct {
	MinFrequency   int
	MinSequence    int
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

// universalAliases contains names that the developer ecosystem has already claimed.
// Suggesting these as names would create silent, dangerous collisions.
var universalAliases = map[string]bool{
	"gs": true, "gc": true, "gco": true, "gcb": true, "gcm": true,
	"gd": true, "gp": true, "gpu": true, "gpl": true, "gl": true,
	"ga": true, "gaa": true, "gst": true, "ll": true, "la": true,
	"k": true, "vi": true, "vim": true, "py": true, "rb": true,
}

// destructivePrefixes are command prefixes where confirm should be set automatically.
var destructivePrefixes = []string{
	"rm ", "rm\t", "kill ", "pkill ", "killall ",
	"dd ", "mkfs", "truncate ", "shred ",
	"git reset", "git clean", "git push --force",
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
	suggestions = append(suggestions, suggestAliases(entries, existingAliases, opts)...)
	suggestions = append(suggestions, detectParamPatterns(entries, existingAliases, opts.MinFrequency)...)
	suggestions = append(suggestions, suggestWorkflows(entries, opts)...)

	if len(suggestions) > opts.MaxSuggestions {
		suggestions = suggestions[:opts.MaxSuggestions]
	}
	return suggestions
}

func suggestAliases(entries []Entry, existing map[string]string, opts SuggestOptions) []Suggestion {
	freq := make(map[string]int)
	for _, e := range entries {
		cmd := strings.TrimSpace(e.Command)
		if cmd != "" && !isNoise(cmd) {
			freq[cmd]++
		}
	}

	existingCmds := make(map[string]bool)
	for _, cmd := range existing {
		existingCmds[cmd] = true
	}

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
		si := keystrokesSaved(sorted[i].cmd, sorted[i].count)
		sj := keystrokesSaved(sorted[j].cmd, sorted[j].count)
		if si != sj {
			return si > sj
		}
		return sorted[i].count > sorted[j].count
	})

	var suggestions []Suggestion
	usedNames := make(map[string]bool)
	for _, name := range existing {
		_ = name
	}
	for name := range existing {
		usedNames[name] = true
	}

	for _, cc := range sorted {
		name := safeAliasName(cc.cmd, usedNames)
		if name == "" {
			continue
		}
		usedNames[name] = true
		saved := keystrokesSaved(cc.cmd, cc.count)
		suggestions = append(suggestions, Suggestion{
			Type:            "alias",
			Name:            name,
			Command:         cc.cmd,
			Reason:          fmt.Sprintf("Used %d times · saves ~%d keystrokes", cc.count, saved),
			SavedKeystrokes: saved,
			Confirm:         isDestructive(cc.cmd),
		})
	}
	return suggestions
}

// detectParamPatterns finds commands that share a common N-token prefix but vary in one argument.
// Applies three quality gates: broken-quote filter, subcommand-vs-argument discrimination, name safety.
func detectParamPatterns(entries []Entry, existing map[string]string, minCount int) []Suggestion {
	freq := make(map[string]int)
	for _, e := range entries {
		cmd := strings.TrimSpace(e.Command)
		if cmd != "" && !isNoise(cmd) {
			freq[cmd]++
		}
	}

	existingNames := make(map[string]bool)
	for name := range existing {
		existingNames[name] = true
	}

	// Group by N-token prefix (N=2,3)
	prefixGroups := make(map[string]map[string]int)
	for cmd, count := range freq {
		tokens := strings.Fields(cmd)
		if len(tokens) < 3 {
			continue
		}
		for prefixLen := 2; prefixLen <= min(3, len(tokens)-1); prefixLen++ {
			prefix := strings.Join(tokens[:prefixLen], " ")
			if hasUnmatchedQuote(prefix) {
				continue
			}
			if prefixGroups[prefix] == nil {
				prefixGroups[prefix] = make(map[string]int)
			}
			prefixGroups[prefix][cmd] += count
		}
	}

	var suggestions []Suggestion
	usedNames := make(map[string]bool)
	for name := range existingNames {
		usedNames[name] = true
	}

	for prefix, cmds := range prefixGroups {
		if len(cmds) < 3 {
			continue
		}

		total := 0
		prefixTokenCount := len(strings.Fields(prefix))
		nextTokens := make(map[string]int)
		tailTokenCounts := make([]int, 0, len(cmds)) // tokens after the varying one

		for cmd, count := range cmds {
			total += count
			tokens := strings.Fields(cmd)
			if len(tokens) > prefixTokenCount {
				varying := tokens[prefixTokenCount]
				nextTokens[varying] += count
				tailTokenCounts = append(tailTokenCounts, len(tokens)-prefixTokenCount-1)
			}
		}

		if total < minCount || len(nextTokens) < 3 {
			continue
		}

		// Gate 1: reject if varying token looks like a subcommand position.
		// If >50% of commands have tokens AFTER the varying one, it's a subcommand.
		tailCount := 0
		for _, tc := range tailTokenCounts {
			if tc > 0 {
				tailCount++
			}
		}
		if tailCount > len(tailTokenCounts)/2 {
			continue
		}

		// Gate 2: reject if most "varying tokens" look like flags or subcommands
		// (short words with no path/url signals = likely subcommands)
		argLikeCount := 0
		for t := range nextTokens {
			if isArgLike(t) {
				argLikeCount++
			}
		}
		if argLikeCount < len(nextTokens)/2 {
			continue
		}

		// Gate 3: skip if varying tokens are mostly flags
		flagCount := 0
		for t := range nextTokens {
			if strings.HasPrefix(t, "-") {
				flagCount++
			}
		}
		if flagCount > len(nextTokens)/2 {
			continue
		}

		name := safeAliasName(prefix, usedNames)
		if name == "" {
			continue
		}
		usedNames[name] = true

		paramName := paramNameForPrefix(prefix)
		saved := (len(prefix) + 2) * total / 10
		suggestions = append(suggestions, Suggestion{
			Type:            "param-alias",
			Name:            name,
			Command:         prefix + " {" + paramName + "}",
			Params:          []string{paramName},
			Reason:          fmt.Sprintf("Pattern '%s <...>' used %d times with %d variants", prefix, total, len(nextTokens)),
			SavedKeystrokes: saved,
			Confirm:         isDestructive(prefix),
		})
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].SavedKeystrokes > suggestions[j].SavedKeystrokes
	})
	return suggestions
}

// isArgLike returns true when a token looks like a real argument (path, URL, identifier with dots/slashes)
// rather than a subcommand keyword.
func isArgLike(t string) bool {
	if strings.ContainsAny(t, "/.:@") {
		return true
	}
	if len(t) > 10 {
		return true
	}
	// Mixed case or digits embedded suggest an identifier/hash
	hasUpper := false
	hasDigit := false
	for _, r := range t {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasUpper || hasDigit
}

const seqMaxGap = 30 * 60 // 30 minutes in seconds — commands further apart aren't a workflow

func suggestWorkflows(entries []Entry, opts SuggestOptions) []Suggestion {
	if len(entries) < 2 {
		return nil
	}

	pairFreq := make(map[string]int)
	tripleFreq := make(map[string]int)

	for i := 0; i < len(entries)-1; i++ {
		a := strings.TrimSpace(entries[i].Command)
		b := strings.TrimSpace(entries[i+1].Command)
		if a == "" || b == "" || a == b || isNoise(a) || isNoise(b) {
			continue
		}
		// Skip pairs where timestamps indicate a session boundary.
		if !entries[i].Timestamp.IsZero() && !entries[i+1].Timestamp.IsZero() {
			gap := entries[i+1].Timestamp.Unix() - entries[i].Timestamp.Unix()
			if gap < 0 {
				gap = -gap
			}
			if gap > seqMaxGap {
				continue
			}
		}
		key := a + " && " + b
		pairFreq[key]++

		if i < len(entries)-2 {
			c := strings.TrimSpace(entries[i+2].Command)
			if c == "" || c == b || isNoise(c) {
				continue
			}
			// Same gap check for the third entry.
			if !entries[i+1].Timestamp.IsZero() && !entries[i+2].Timestamp.IsZero() {
				gap := entries[i+2].Timestamp.Unix() - entries[i+1].Timestamp.Unix()
				if gap < 0 {
					gap = -gap
				}
				if gap > seqMaxGap {
					continue
				}
			}
			triKey := a + " && " + b + " && " + c
			tripleFreq[triKey]++
		}
	}

	var suggestions []Suggestion

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
	sort.Slice(triples, func(i, j int) bool { return triples[i].count > triples[j].count })

	usedWfNames := make(map[string]bool)
	for _, tc := range triples {
		steps := strings.Split(tc.seq, " && ")
		name := workflowName(steps, usedWfNames)
		usedWfNames[name] = true
		suggestions = append(suggestions, Suggestion{
			Type:   "workflow",
			Name:   name,
			Steps:  steps,
			Reason: fmt.Sprintf("Sequence appeared %d times", tc.count),
		})
	}

	var pairs []seqCount
	for seq, count := range pairFreq {
		if count >= opts.MinSequence {
			pairs = append(pairs, seqCount{seq, count})
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].count > pairs[j].count })

	for _, pc := range pairs {
		steps := strings.Split(pc.seq, " && ")
		// skip pairs already covered by a triple with the same first two steps
		covered := false
		for _, s := range suggestions {
			if len(s.Steps) >= 2 && s.Steps[0] == steps[0] && s.Steps[1] == steps[1] {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		name := workflowName(steps, usedWfNames)
		usedWfNames[name] = true
		suggestions = append(suggestions, Suggestion{
			Type:   "workflow",
			Name:   name,
			Steps:  steps,
			Reason: fmt.Sprintf("Pair appeared %d times", pc.count),
		})
	}

	return suggestions
}

// safeAliasName generates a conflict-free alias name, returning "" if no safe name is found.
func safeAliasName(cmd string, used map[string]bool) string {
	base := generateAliasName(cmd)
	candidates := []string{base}

	// If base conflicts, try longer forms
	tokens := strings.Fields(cmd)
	if len(tokens) >= 2 {
		// First char of first word + first 2 chars of second
		w0 := strings.TrimLeft(tokens[0], "-./")
		w1 := strings.TrimLeft(tokens[1], "-./")
		if len(w0) > 0 && len(w1) > 0 {
			if len(w1) >= 2 {
				candidates = append(candidates, string(w0[0])+w1[:2])
			}
			if len(w1) >= 3 {
				candidates = append(candidates, string(w0[0])+w1[:3])
			}
		}
	}

	for _, name := range candidates {
		if name == "" {
			continue
		}
		if universalAliases[name] {
			continue
		}
		if used[name] {
			continue
		}
		return name
	}
	return ""
}

// isDestructive returns true when the command prefix matches known destructive patterns.
func isDestructive(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, prefix := range destructivePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// hasUnmatchedQuote returns true if the string has an odd number of " or ' characters.
func hasUnmatchedQuote(s string) bool {
	dq := strings.Count(s, `"`)
	sq := strings.Count(s, `'`)
	return dq%2 != 0 || sq%2 != 0
}

// isNoise returns true for commands that shouldn't be aliased.
func isNoise(cmd string) bool {
	if len(cmd) < 4 {
		return true
	}
	if strings.HasPrefix(cmd, "#") {
		return true
	}
	if strings.HasSuffix(cmd, `\`) {
		return true
	}
	allPunct := true
	for _, r := range cmd {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			allPunct = false
			break
		}
	}
	return allPunct
}

// keystrokesSaved estimates keystrokes saved over the command's history frequency.
func keystrokesSaved(cmd string, freq int) int {
	name := generateAliasName(cmd)
	saved := len(cmd) - len(name)
	if saved <= 0 {
		return 0
	}
	return saved * freq
}

// generateAliasName creates a short alias name from a command using initials.
func generateAliasName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "cmd"
	}

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

// workflowName generates a readable name from workflow steps.
// Tries "verb-noun" from the first command, falls back to initials.
// Ensures uniqueness within the used set.
func workflowName(steps []string, used map[string]bool) string {
	base := workflowBaseName(steps)
	name := base
	for i := 2; used[name]; i++ {
		name = fmt.Sprintf("%s-%d", base, i)
	}
	return name
}

func workflowBaseName(steps []string) string {
	if len(steps) == 0 {
		return "workflow"
	}
	tokens := strings.Fields(steps[0])
	if len(tokens) == 0 {
		return "workflow"
	}

	// Known verb→short mappings
	verbMap := map[string]string{
		"git": "git", "npm": "npm", "pnpm": "pnpm", "yarn": "yarn",
		"docker": "docker", "kubectl": "k8s", "make": "make",
		"go": "go", "cargo": "cargo", "python3": "py", "python": "py",
	}

	verb := strings.ToLower(tokens[0])
	prefix, ok := verbMap[verb]
	if !ok {
		// use first word directly if short enough
		if len(verb) <= 6 {
			prefix = verb
		} else {
			prefix = verb[:4]
		}
	}

	// Pick a noun from the subcommand or second step
	noun := ""
	if len(tokens) >= 2 {
		sub := strings.ToLower(strings.TrimLeft(tokens[1], "-"))
		// only use if it looks like a subcommand (short, alphabetic)
		if len(sub) >= 2 && len(sub) <= 8 && isAlpha(sub) {
			noun = sub
		}
	}
	// if no noun from first step, try key verb from second step
	if noun == "" && len(steps) >= 2 {
		t2 := strings.Fields(steps[1])
		if len(t2) >= 2 {
			sub := strings.ToLower(strings.TrimLeft(t2[1], "-"))
			if len(sub) >= 2 && len(sub) <= 8 && isAlpha(sub) {
				noun = sub
			}
		}
	}

	if noun != "" {
		return prefix + "-" + noun
	}
	return prefix + "-flow"
}

func isAlpha(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

// paramNameForPrefix returns the best parameter name for a given command prefix.
func paramNameForPrefix(prefix string) string {
	lower := strings.ToLower(prefix)

	// Exact prefix matches (checked longest first)
	prefixToParam := []struct {
		prefix string
		param  string
	}{
		{"git push origin", "branch"},
		{"git checkout", "branch"},
		{"git merge", "branch"},
		{"git rebase", "branch"},
		{"git pull origin", "branch"},
		{"git add", "file"},
		{"git clone", "repo"},
		{"git diff", "ref"},
		{"npm run", "script"},
		{"npm install", "package"},
		{"npm i -g", "package"},
		{"npm i", "package"},
		{"brew install --cask", "cask"},
		{"brew install", "package"},
		{"brew uninstall", "package"},
		{"pip install", "package"},
		{"pip uninstall", "package"},
		{"cargo add", "crate"},
		{"go get", "module"},
		{"rm -rf", "path"},
		{"rm -r", "path"},
		{"mkdir -p", "dir"},
		{"mkdir", "dir"},
		{"cd", "dir"},
		{"head -", "file"},
		{"tail -", "file"},
		{"cat", "file"},
		{"less", "file"},
		{"kubectl get", "resource"},
		{"kubectl delete", "resource"},
		{"kubectl describe", "resource"},
		{"kubectl apply", "file"},
		{"docker run", "image"},
		{"docker exec", "container"},
		{"docker stop", "container"},
		{"docker rm", "container"},
		{"docker rmi", "image"},
		{"docker pull", "image"},
		{"ssh", "host"},
		{"scp", "target"},
		{"curl", "url"},
		{"wget", "url"},
		{"open", "path"},
		{"code", "path"},
		{"kill", "pid"},
		{"pkill", "process"},
	}

	for _, entry := range prefixToParam {
		if strings.HasPrefix(lower, entry.prefix) {
			return entry.param
		}
	}

	// Fallback: last-token context hints
	tokens := strings.Fields(prefix)
	lastTokenHints := map[string]string{
		"origin":    "branch",
		"push":      "branch",
		"checkout":  "branch",
		"merge":     "branch",
		"rebase":    "branch",
		"run":       "script",
		"-n":        "namespace",
		"-f":        "file",
		"apply":     "file",
		"create":    "name",
		"delete":    "name",
		"get":       "name",
		"describe":  "name",
		"install":   "package",
		"uninstall": "package",
		"add":       "target",
		"remove":    "target",
	}
	for i := len(tokens) - 1; i >= max(0, len(tokens)-2); i-- {
		t := strings.ToLower(strings.TrimLeft(tokens[i], "-"))
		if hint, ok := lastTokenHints[t]; ok {
			return hint
		}
	}
	return "arg"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
