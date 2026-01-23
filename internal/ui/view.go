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

	// Update log panel with current service status
	m.updateLogPanelStatus()

	// Main content area
	var mainArea string
	if m.fullscreen {
		// Fullscreen mode: only logs
		mainArea = m.logPanel.View(m.logBuffer)
	} else {
		// Normal mode: sidebar + logs
		sidebar := m.sidebar.View(m.manager, m.logBuffer)
		logs := m.logPanel.View(m.logBuffer)
		mainArea = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, logs)
	}

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

	if m.showMoveService {
		return m.overlayMoveServiceModal(b.String())
	}

	if m.showRename {
		return m.overlayRenameModal(b.String())
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

// overlayMoveServiceModal overlays the move service modal
func (m *Model) overlayMoveServiceModal(background string) string {
	m.moveServiceModal.SetSize(m.width / 2)

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return modalStyle.Render(m.moveServiceModal.View())
}

// overlayRenameModal overlays the rename modal
func (m *Model) overlayRenameModal(background string) string {
	m.renameModal.SetSize(m.width / 2)

	modalStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return modalStyle.Render(m.renameModal.View())
}
