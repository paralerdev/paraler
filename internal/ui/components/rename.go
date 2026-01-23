package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// RenameTarget represents what is being renamed
type RenameTarget int

const (
	RenameNone RenameTarget = iota
	RenameProject
	RenameService
)

// RenameModal is a dialog for renaming projects or services
type RenameModal struct {
	visible     bool
	target      RenameTarget
	projectName string
	serviceName string
	input       textinput.Model
	errorMsg    string
	width       int
	styles      RenameStyles
}

// RenameStyles contains styles for the modal
type RenameStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Label     lipgloss.Style
	Error     lipgloss.Style
	Help      lipgloss.Style
}

// DefaultRenameStyles returns default styles
func DefaultRenameStyles() RenameStyles {
	return RenameStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8B5CF6")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8B5CF6")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			MarginTop(1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
	}
}

// NewRenameModal creates a new rename modal
func NewRenameModal() *RenameModal {
	ti := textinput.New()
	ti.Placeholder = "new name"
	ti.CharLimit = 64
	ti.Width = 30

	return &RenameModal{
		input:  ti,
		styles: DefaultRenameStyles(),
	}
}

// SetSize sets the modal width
func (m *RenameModal) SetSize(width int) {
	m.width = width
	m.input.Width = width - 10
}

// ShowRenameProject shows the modal for renaming a project
func (m *RenameModal) ShowRenameProject(projectName string) {
	m.target = RenameProject
	m.projectName = projectName
	m.serviceName = ""
	m.errorMsg = ""
	m.input.SetValue(projectName)
	m.input.Focus()
	m.input.CursorEnd()
	m.visible = true
}

// ShowRenameService shows the modal for renaming a service
func (m *RenameModal) ShowRenameService(projectName, serviceName string) {
	m.target = RenameService
	m.projectName = projectName
	m.serviceName = serviceName
	m.errorMsg = ""
	m.input.SetValue(serviceName)
	m.input.Focus()
	m.input.CursorEnd()
	m.visible = true
}

// Hide hides the modal
func (m *RenameModal) Hide() {
	m.visible = false
	m.target = RenameNone
	m.input.Blur()
}

// IsVisible returns true if modal is visible
func (m *RenameModal) IsVisible() bool {
	return m.visible
}

// Target returns the rename target type
func (m *RenameModal) Target() RenameTarget {
	return m.target
}

// ProjectName returns the project name
func (m *RenameModal) ProjectName() string {
	return m.projectName
}

// ServiceName returns the service name (for service rename)
func (m *RenameModal) ServiceName() string {
	return m.serviceName
}

// NewName returns the entered new name
func (m *RenameModal) NewName() string {
	return strings.TrimSpace(m.input.Value())
}

// SetError sets an error message
func (m *RenameModal) SetError(err string) {
	m.errorMsg = err
}

// Input returns the text input model
func (m *RenameModal) Input() *textinput.Model {
	return &m.input
}

// View renders the modal
func (m *RenameModal) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	// Title
	var title string
	var currentName string
	switch m.target {
	case RenameProject:
		title = "Rename Project"
		currentName = m.projectName
	case RenameService:
		title = "Rename Service"
		currentName = m.serviceName
	}

	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Label.Render(fmt.Sprintf("Current: %s", currentName)))
	b.WriteString("\n\n")

	b.WriteString(m.input.View())

	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Error.Render(m.errorMsg))
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.Help.Render("enter confirm • Esc cancel"))

	return m.styles.Container.
		Width(m.width).
		Render(b.String())
}
