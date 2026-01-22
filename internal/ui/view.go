package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Main content area
	sidebar := m.sidebar.View(m.manager, m.logBuffer)
	logs := m.logPanel.View(m.logBuffer)

	// Join sidebar and logs horizontally
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, logs)

	// Status bar
	var statusBar string
	if m.showHelp {
		statusBar = m.statusBar.View(m.manager, true)
	} else {
		statusBar = m.statusBar.View(m.manager, false)
	}

	// Join vertically
	var b strings.Builder
	b.WriteString(mainArea)
	b.WriteString("\n")
	b.WriteString(statusBar)

	// Overlay modals if visible
	if m.showConfirm {
		return m.overlayConfirmModal(b.String())
	}

	if m.showAddProject {
		return m.overlayModal(b.String(), m.addProjectModal.View())
	}

	return b.String()
}

// overlayModal places a modal on top of the background
func (m *Model) overlayModal(background, modal string) string {
	// Calculate modal position (center of screen)
	modalWidth := m.width / 2
	modalHeight := m.height / 2

	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalHeight < 15 {
		modalHeight = 15
	}

	m.addProjectModal.SetSize(modalWidth, modalHeight)

	// Simple overlay: just show the modal centered
	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return modalStyle.Render(m.addProjectModal.View())
}

// overlayConfirmModal overlays the confirm modal
func (m *Model) overlayConfirmModal(background string) string {
	m.confirmModal.SetSize(m.width / 2)

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return modalStyle.Render(m.confirmModal.View())
}
