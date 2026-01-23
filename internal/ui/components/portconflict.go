package components

import (
	"fmt"
	"strings"

	"github.com/paralerdev/paraler/internal/config"
	"github.com/paralerdev/paraler/internal/process"
	"github.com/charmbracelet/lipgloss"
)

// PortConflictModal shows port conflict information and options
type PortConflictModal struct {
	visible      bool
	conflict     *process.PortConflictInfo
	serviceID    config.ServiceID // The service we're trying to start
	width        int
	styles       PortConflictStyles
}

// PortConflictStyles contains styles for the modal
type PortConflictStyles struct {
	Container   lipgloss.Style
	Title       lipgloss.Style
	Port        lipgloss.Style
	ProcessInfo lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Help        lipgloss.Style
}

// DefaultPortConflictStyles returns default styles
func DefaultPortConflictStyles() PortConflictStyles {
	return PortConflictStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F59E0B")),
		Port: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB")),
		ProcessInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginTop(1),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
	}
}

// NewPortConflictModal creates a new port conflict modal
func NewPortConflictModal() *PortConflictModal {
	return &PortConflictModal{
		styles: DefaultPortConflictStyles(),
	}
}

// SetSize sets the modal width
func (m *PortConflictModal) SetSize(width int) {
	m.width = width
}

// Show shows the modal with conflict info
func (m *PortConflictModal) Show(serviceID config.ServiceID, conflict *process.PortConflictInfo) {
	m.visible = true
	m.serviceID = serviceID
	m.conflict = conflict
}

// Hide hides the modal
func (m *PortConflictModal) Hide() {
	m.visible = false
	m.conflict = nil
}

// IsVisible returns true if modal is visible
func (m *PortConflictModal) IsVisible() bool {
	return m.visible
}

// Conflict returns the current conflict info
func (m *PortConflictModal) Conflict() *process.PortConflictInfo {
	return m.conflict
}

// ServiceID returns the service we're trying to start
func (m *PortConflictModal) ServiceID() config.ServiceID {
	return m.serviceID
}

// View renders the modal
func (m *PortConflictModal) View() string {
	if !m.visible || m.conflict == nil {
		return ""
	}

	var b strings.Builder

	// Title
	title := fmt.Sprintf("Port %d is busy", m.conflict.Port)
	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n\n")

	// Process info
	if m.conflict.IsParalerService {
		// Another paraler service
		b.WriteString(m.styles.Label.Render("Used by: "))
		svcInfo := fmt.Sprintf("%s (%s)",
			m.conflict.ParalerServiceID.Service,
			m.conflict.ParalerServiceID.Project)
		b.WriteString(m.styles.Value.Render(svcInfo))
		b.WriteString("\n")
		b.WriteString(m.styles.ProcessInfo.Render("This is another service managed by paraler"))
	} else {
		// External process
		if m.conflict.ExternalProcess != "" {
			b.WriteString(m.styles.Label.Render("Process: "))
			b.WriteString(m.styles.Value.Render(m.conflict.ExternalProcess))
			b.WriteString("\n")
		}
		if m.conflict.ExternalPID > 0 {
			b.WriteString(m.styles.Label.Render("PID: "))
			b.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.conflict.ExternalPID)))
			b.WriteString("\n")
		}
		if m.conflict.ExternalCommand != "" {
			// Truncate long commands
			cmd := m.conflict.ExternalCommand
			maxLen := m.width - 15
			if maxLen < 20 {
				maxLen = 20
			}
			if len(cmd) > maxLen {
				cmd = cmd[:maxLen-3] + "..."
			}
			b.WriteString(m.styles.Label.Render("Command: "))
			b.WriteString(m.styles.Value.Render(cmd))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Help
	if m.conflict.IsParalerService {
		b.WriteString(m.styles.Help.Render("k kill & start • Esc cancel"))
	} else {
		b.WriteString(m.styles.Help.Render("k kill & start • Esc cancel"))
	}

	return m.styles.Container.
		Width(m.width).
		Render(b.String())
}
