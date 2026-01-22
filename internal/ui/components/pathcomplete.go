package components

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PathCompleter provides path autocompletion
type PathCompleter struct {
	suggestions []string
	index       int
}

// NewPathCompleter creates a new path completer
func NewPathCompleter() *PathCompleter {
	return &PathCompleter{}
}

// Complete returns completions for the given path
func (p *PathCompleter) Complete(input string) []string {
	if input == "" {
		input = "."
	}

	// Expand ~
	expanded := expandTilde(input)

	// Get directory and prefix
	dir := expanded
	prefix := ""

	if !strings.HasSuffix(expanded, string(filepath.Separator)) {
		dir = filepath.Dir(expanded)
		prefix = filepath.Base(expanded)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless prefix starts with .
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		// Only directories
		if !entry.IsDir() {
			continue
		}

		// Match prefix
		if prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}

		// Build full path
		fullPath := filepath.Join(dir, name)

		// Convert back to use ~ if it was used
		if strings.HasPrefix(input, "~") {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(fullPath, home) {
				fullPath = "~" + fullPath[len(home):]
			}
		}

		suggestions = append(suggestions, fullPath+string(filepath.Separator))
	}

	sort.Strings(suggestions)
	return suggestions
}

// GetNext returns the next suggestion
func (p *PathCompleter) GetNext(input string) string {
	if len(p.suggestions) == 0 || !p.matchesCurrentInput(input) {
		p.suggestions = p.Complete(input)
		p.index = 0
	}

	if len(p.suggestions) == 0 {
		return input
	}

	suggestion := p.suggestions[p.index]
	p.index = (p.index + 1) % len(p.suggestions)
	return suggestion
}

// matchesCurrentInput checks if suggestions are still valid
func (p *PathCompleter) matchesCurrentInput(input string) bool {
	if len(p.suggestions) == 0 {
		return false
	}

	// Check if input is one of our suggestions or a prefix of them
	for _, s := range p.suggestions {
		if strings.HasPrefix(s, input) || input == s {
			return true
		}
	}
	return false
}

// Reset clears the suggestions
func (p *PathCompleter) Reset() {
	p.suggestions = nil
	p.index = 0
}

// GetSuggestions returns current suggestions
func (p *PathCompleter) GetSuggestions(input string, max int) []string {
	suggestions := p.Complete(input)
	if len(suggestions) > max {
		return suggestions[:max]
	}
	return suggestions
}

// expandTilde expands ~ to home directory
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// CommonPrefix returns the longest common prefix of suggestions
func CommonPrefix(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}
	if len(suggestions) == 1 {
		return suggestions[0]
	}

	prefix := suggestions[0]
	for _, s := range suggestions[1:] {
		for len(prefix) > 0 && !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
