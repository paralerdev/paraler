package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MoveServiceModal is a dialog for selecting target project
type MoveServiceModal struct {
	visible     bool
	projects    []string
	selected    int
	serviceName string
	fromProject string
	width       int
	styles      MoveServiceStyles
}

// MoveServiceStyles contains styles for the modal
type MoveServiceStyles struct {
	Container    lipgloss.Style
	Title        lipgloss.Style
	ServiceName  lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
	Help         lipgloss.Style
}

// DefaultMoveServiceStyles returns default styles
func DefaultMoveServiceStyles() MoveServiceStyles {
	return MoveServiceStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8B5CF6")).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8B5CF6")),
		ServiceName: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Bold(true),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB")).
			Bold(true).
			PaddingLeft(2),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),
	}
}

// NewMoveServiceModal creates a new move service modal
func NewMoveServiceModal() *MoveServiceModal {
	return &MoveServiceModal{
		styles: DefaultMoveServiceStyles(),
	}
}

// SetSize sets the modal width
func (m *MoveServiceModal) SetSize(width int) {
	m.width = width
}

// Show shows the modal with available target projects
func (m *MoveServiceModal) Show(serviceName, fromProject string, allProjects []string) {
	m.serviceName = serviceName
	m.fromProject = fromProject
	m.selected = 0

	// Filter out the current project
	m.projects = make([]string, 0, len(allProjects)-1)
	for _, p := range allProjects {
		if p != fromProject {
			m.projects = append(m.projects, p)
		}
	}

	// Sort alphabetically
	sort.Strings(m.projects)

	m.visible = len(m.projects) > 0
}

// Hide hides the modal
func (m *MoveServiceModal) Hide() {
	m.visible = false
}

// IsVisible returns true if modal is visible
func (m *MoveServiceModal) IsVisible() bool {
	return m.visible
}

// HasTargets returns true if there are target projects available
func (m *MoveServiceModal) HasTargets() bool {
	return len(m.projects) > 0
}

// MoveUp moves selection up
func (m *MoveServiceModal) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

// MoveDown moves selection down
func (m *MoveServiceModal) MoveDown() {
	if m.selected < len(m.projects)-1 {
		m.selected++
	}
}

// SelectedProject returns the currently selected project
func (m *MoveServiceModal) SelectedProject() string {
	if m.selected < len(m.projects) {
		return m.projects[m.selected]
	}
	return ""
}

// ServiceName returns the service being moved
func (m *MoveServiceModal) ServiceName() string {
	return m.serviceName
}

// FromProject returns the source project
func (m *MoveServiceModal) FromProject() string {
	return m.fromProject
}

// View renders the modal
func (m *MoveServiceModal) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.styles.Title.Render("Move Service"))
	b.WriteString("\n\n")

	b.WriteString("Move ")
	b.WriteString(m.styles.ServiceName.Render(m.serviceName))
	b.WriteString(" to:\n\n")

	for i, project := range m.projects {
		if i == m.selected {
			b.WriteString(m.styles.SelectedItem.Render(fmt.Sprintf("→ %s", project)))
		} else {
			b.WriteString(m.styles.Item.Render(fmt.Sprintf("  %s", project)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("↑/↓ select • enter confirm • Esc cancel"))

	return m.styles.Container.
		Width(m.width).
		Render(b.String())
}
