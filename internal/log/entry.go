package log

import (
	"time"

	"github.com/paralerdev/paraler/internal/config"
)

// Entry represents a single log entry
type Entry struct {
	ServiceID config.ServiceID
	Line      string
	IsStderr  bool
	Timestamp time.Time
}

// NewEntry creates a new log entry
func NewEntry(serviceID config.ServiceID, line string, isStderr bool) Entry {
	return Entry{
		ServiceID: serviceID,
		Line:      line,
		IsStderr:  isStderr,
		Timestamp: time.Now(),
	}
}
