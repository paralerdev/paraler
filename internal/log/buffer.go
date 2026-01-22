package log

import (
	"strings"
	"sync"

	"github.com/paralerdev/paraler/internal/config"
)

const (
	// DefaultBufferSize is the default number of entries per service
	DefaultBufferSize = 1000
)

// Buffer is a ring buffer for storing log entries per service
type Buffer struct {
	mu      sync.RWMutex
	entries map[string][]Entry // key: ServiceID.String()
	maxSize int
}

// NewBuffer creates a new log buffer
func NewBuffer(maxSize int) *Buffer {
	if maxSize <= 0 {
		maxSize = DefaultBufferSize
	}
	return &Buffer{
		entries: make(map[string][]Entry),
		maxSize: maxSize,
	}
}

// Add adds an entry to the buffer
func (b *Buffer) Add(entry Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := entry.ServiceID.String()
	entries := b.entries[key]

	// Add entry
	entries = append(entries, entry)

	// Trim if over capacity
	if len(entries) > b.maxSize {
		entries = entries[len(entries)-b.maxSize:]
	}

	b.entries[key] = entries
}

// Get returns all entries for a service
func (b *Buffer) Get(id config.ServiceID) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entries := b.entries[id.String()]
	result := make([]Entry, len(entries))
	copy(result, entries)
	return result
}

// GetAll returns all entries across all services
func (b *Buffer) GetAll() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var all []Entry
	for _, entries := range b.entries {
		all = append(all, entries...)
	}
	return all
}

// GetFiltered returns entries matching a filter string
func (b *Buffer) GetFiltered(id config.ServiceID, filter string) []Entry {
	entries := b.Get(id)

	if filter == "" {
		return entries
	}

	filter = strings.ToLower(filter)
	var filtered []Entry
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Line), filter) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Clear removes all entries for a service
func (b *Buffer) Clear(id config.ServiceID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, id.String())
}

// ClearAll removes all entries
func (b *Buffer) ClearAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = make(map[string][]Entry)
}

// Count returns the number of entries for a service
func (b *Buffer) Count(id config.ServiceID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries[id.String()])
}

// TotalCount returns the total number of entries
func (b *Buffer) TotalCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, entries := range b.entries {
		count += len(entries)
	}
	return count
}

// GetLines returns log entries as formatted strings
func (b *Buffer) GetLines(id config.ServiceID, filter string, showTimestamp bool) []string {
	entries := b.GetFiltered(id, filter)
	lines := make([]string, len(entries))

	for i, entry := range entries {
		if showTimestamp {
			lines[i] = entry.Timestamp.Format("15:04:05") + " " + entry.Line
		} else {
			lines[i] = entry.Line
		}
	}

	return lines
}

// ErrorCount returns the number of stderr entries for a service
func (b *Buffer) ErrorCount(id config.ServiceID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entries := b.entries[id.String()]
	count := 0
	for _, entry := range entries {
		if entry.IsStderr {
			count++
		}
	}
	return count
}
