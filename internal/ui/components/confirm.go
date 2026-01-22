package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmAction represents what action needs confirmation
type ConfirmAction int

const (
	ConfirmNone ConfirmAction = iota
	ConfirmDeleteService
	ConfirmDeleteProject
)

// ConfirmModal is a confirmation dialog
type ConfirmModal struct {
	action      ConfirmAction
	title       string
	message     string
	targetName  string
	projectName string
	width       int
	styles      ConfirmStyles
}

// ConfirmStyles contains styles for the modal
type ConfirmStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Message   lipgloss.Style
	Warning   lipgloss.Style
	Help      lipgloss.Style
}

// DefaultConfirmStyles returns default styles
func DefaultConfirmStyles() ConfirmStyles {
	return ConfirmStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444")),
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			MarginTop(1),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Italic(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
	}
}

// NewConfirmModal creates a new confirmation modal
func NewConfirmModal() *ConfirmModal {
	return &ConfirmModal{
		action: ConfirmNone,
		styles: DefaultConfirmStyles(),
	}
}

// SetSize sets the modal width
func (m *ConfirmModal) SetSize(width int) {
	m.width = width
}

// Show shows the confirmation dialog
func (m *ConfirmModal) Show(action ConfirmAction, projectName, serviceName string) {
	m.action = action
	m.projectName = projectName

	switch action {
	case ConfirmDeleteService:
		m.title = "Delete Service"
		m.targetName = serviceName
		m.message = fmt.Sprintf("Delete service '%s' from project '%s'?", serviceName, projectName)
	case ConfirmDeleteProject:
		m.title = "Delete Project"
		m.targetName = projectName
		m.message = fmt.Sprintf("Delete project '%s' and all its services?", projectName)
	}
}

// Hide hides the modal
func (m *ConfirmModal) Hide() {
	m.action = ConfirmNone
}

// IsVisible returns true if modal is visible
func (m *ConfirmModal) IsVisible() bool {
	return m.action != ConfirmNone
}

// Action returns the current action
func (m *ConfirmModal) Action() ConfirmAction {
	return m.action
}

// ProjectName returns the project name
func (m *ConfirmModal) ProjectName() string {
	return m.projectName
}

// TargetName returns the target name (service or project)
func (m *ConfirmModal) TargetName() string {
	return m.targetName
}

// View renders the modal
func (m *ConfirmModal) View() string {
	if !m.IsVisible() {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.styles.Title.Render(m.title))
	b.WriteString("\n")
	b.WriteString(m.styles.Message.Render(m.message))
	b.WriteString("\n\n")

	if m.action == ConfirmDeleteProject {
		b.WriteString(m.styles.Warning.Render("This will remove all services in this project."))
		b.WriteString("\n\n")
	}

	b.WriteString(m.styles.Help.Render("y confirm • n/Esc cancel"))

	return m.styles.Container.
		Width(m.width).
		Render(b.String())
}
