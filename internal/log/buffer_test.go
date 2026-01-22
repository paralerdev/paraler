package log

import (
	"testing"
	"time"

	"github.com/paralerdev/paraler/internal/config"
)

func TestBuffer_Add(t *testing.T) {
	buf := NewBuffer(10)

	id := config.ServiceID{Project: "test", Service: "backend"}
	entry := Entry{
		ServiceID: id,
		Line:      "test line",
		IsStderr:  false,
		Timestamp: time.Now(),
	}

	buf.Add(entry)

	entries := buf.Get(id)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Line != "test line" {
		t.Errorf("expected 'test line', got %q", entries[0].Line)
	}
}

func TestBuffer_RingBuffer(t *testing.T) {
	buf := NewBuffer(5)

	id := config.ServiceID{Project: "test", Service: "backend"}

	// Add 10 entries, buffer should only keep last 5
	for i := 0; i < 10; i++ {
		buf.Add(Entry{
			ServiceID: id,
			Line:      string(rune('a' + i)),
			Timestamp: time.Now(),
		})
	}

	entries := buf.Get(id)
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	// Should have entries f, g, h, i, j (last 5)
	if entries[0].Line != "f" {
		t.Errorf("expected first entry 'f', got %q", entries[0].Line)
	}
	if entries[4].Line != "j" {
		t.Errorf("expected last entry 'j', got %q", entries[4].Line)
	}
}

func TestBuffer_GetFiltered(t *testing.T) {
	buf := NewBuffer(100)

	id := config.ServiceID{Project: "test", Service: "backend"}

	buf.Add(Entry{ServiceID: id, Line: "error: something failed", Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id, Line: "info: all good", Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id, Line: "ERROR: another error", Timestamp: time.Now()})

	// Filter for "error" (case-insensitive)
	filtered := buf.GetFiltered(id, "error")
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered entries, got %d", len(filtered))
	}

	// No filter
	all := buf.GetFiltered(id, "")
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
}

func TestBuffer_ErrorCount(t *testing.T) {
	buf := NewBuffer(100)

	id := config.ServiceID{Project: "test", Service: "backend"}

	buf.Add(Entry{ServiceID: id, Line: "stdout line", IsStderr: false, Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id, Line: "stderr line", IsStderr: true, Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id, Line: "another stdout", IsStderr: false, Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id, Line: "another stderr", IsStderr: true, Timestamp: time.Now()})

	count := buf.ErrorCount(id)
	if count != 2 {
		t.Errorf("expected 2 errors, got %d", count)
	}
}

func TestBuffer_Clear(t *testing.T) {
	buf := NewBuffer(100)

	id1 := config.ServiceID{Project: "test", Service: "backend"}
	id2 := config.ServiceID{Project: "test", Service: "frontend"}

	buf.Add(Entry{ServiceID: id1, Line: "line1", Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id2, Line: "line2", Timestamp: time.Now()})

	buf.Clear(id1)

	if len(buf.Get(id1)) != 0 {
		t.Error("id1 should be empty after clear")
	}

	if len(buf.Get(id2)) != 1 {
		t.Error("id2 should still have 1 entry")
	}
}

func TestBuffer_ClearAll(t *testing.T) {
	buf := NewBuffer(100)

	id1 := config.ServiceID{Project: "test", Service: "backend"}
	id2 := config.ServiceID{Project: "test", Service: "frontend"}

	buf.Add(Entry{ServiceID: id1, Line: "line1", Timestamp: time.Now()})
	buf.Add(Entry{ServiceID: id2, Line: "line2", Timestamp: time.Now()})

	buf.ClearAll()

	if buf.TotalCount() != 0 {
		t.Error("total count should be 0 after ClearAll")
	}
}

func TestBuffer_Count(t *testing.T) {
	buf := NewBuffer(100)

	id := config.ServiceID{Project: "test", Service: "backend"}

	for i := 0; i < 5; i++ {
		buf.Add(Entry{ServiceID: id, Line: "line", Timestamp: time.Now()})
	}

	if buf.Count(id) != 5 {
		t.Errorf("expected count 5, got %d", buf.Count(id))
	}
}
